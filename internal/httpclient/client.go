// Package httpclient executes a model.Request and converts the outcome into a
// model.Response. Network and parse failures populate Response.Error rather
// than returning a Go error, so the UI always receives a structured result it
// can display.
package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"payload/internal/model"
)

const defaultTimeout = 60 * time.Second

var varRE = regexp.MustCompile(`\{\{(\w+)\}\}`)

// expand replaces every {{name}} occurrence with vars[name]. Unknown names
// are left as-is so the user can spot the typo in the response or URL.
func expand(s string, vars map[string]string) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	return varRE.ReplaceAllStringFunc(s, func(match string) string {
		name := varRE.FindStringSubmatch(match)[1]
		if v, ok := vars[name]; ok {
			return v
		}
		return match
	})
}

// applyVars returns a copy of r with {{var}} substitutions applied to URL,
// header keys/values, query param keys/values, and body content.
func applyVars(r model.Request, vars map[string]string) model.Request {
	if len(vars) == 0 {
		return r
	}
	out := r
	out.URL = expand(r.URL, vars)
	out.Headers = make([]model.KV, len(r.Headers))
	for i, h := range r.Headers {
		out.Headers[i] = model.KV{
			Key:     expand(h.Key, vars),
			Value:   expand(h.Value, vars),
			Enabled: h.Enabled,
		}
	}
	out.QueryParams = make([]model.KV, len(r.QueryParams))
	for i, q := range r.QueryParams {
		out.QueryParams[i] = model.KV{
			Key:     expand(q.Key, vars),
			Value:   expand(q.Value, vars),
			Enabled: q.Enabled,
		}
	}
	out.Body = model.Body{Type: r.Body.Type, Content: expand(r.Body.Content, vars)}
	return out
}

// Send executes r and returns the structured Response. The optional vars map
// is applied with {{name}} substitution to every textual field on the request
// before it goes out on the wire.
func Send(ctx context.Context, r model.Request, vars map[string]string) model.Response {
	r = applyVars(r, vars)
	out := model.Response{SentAt: time.Now().UTC()}

	rawURL := strings.TrimSpace(r.URL)
	if rawURL == "" {
		out.Error = "url is required"
		return out
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		out.Error = fmt.Sprintf("invalid url: %v", err)
		return out
	}
	if parsed.Scheme == "" {
		// Default to https when the user types "example.com/foo" — friendlier
		// than failing the request outright.
		parsed, err = url.Parse("https://" + rawURL)
		if err != nil {
			out.Error = fmt.Sprintf("invalid url: %v", err)
			return out
		}
	}

	if len(r.QueryParams) > 0 {
		q := parsed.Query()
		for _, p := range r.QueryParams {
			if !p.Enabled || p.Key == "" {
				continue
			}
			q.Add(p.Key, p.Value)
		}
		parsed.RawQuery = q.Encode()
	}

	body, defaultCT := buildBody(r.Body)

	method := strings.ToUpper(strings.TrimSpace(string(r.Method)))
	if method == "" {
		method = http.MethodGet
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, parsed.String(), body)
	if err != nil {
		out.Error = err.Error()
		return out
	}

	userSetCT := false
	for _, h := range r.Headers {
		if !h.Enabled || h.Key == "" {
			continue
		}
		httpReq.Header.Add(h.Key, h.Value)
		if strings.EqualFold(h.Key, "Content-Type") {
			userSetCT = true
		}
	}
	if !userSetCT && defaultCT != "" {
		httpReq.Header.Set("Content-Type", defaultCT)
	}

	client := &http.Client{Timeout: defaultTimeout}
	start := time.Now()
	httpResp, err := client.Do(httpReq)
	out.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			out.Error = fmt.Sprintf("timeout after %s", defaultTimeout)
		} else {
			out.Error = err.Error()
		}
		return out
	}
	defer httpResp.Body.Close()

	bodyBytes, readErr := io.ReadAll(httpResp.Body)
	out.Status = httpResp.StatusCode
	out.StatusText = httpResp.Status
	for k, vs := range httpResp.Header {
		for _, v := range vs {
			out.Headers = append(out.Headers, model.KV{Key: k, Value: v, Enabled: true})
		}
	}
	if readErr != nil {
		out.Error = fmt.Sprintf("read body: %v", readErr)
		return out
	}
	out.Body = string(bodyBytes)
	return out
}

// buildBody returns an io.Reader for the configured body and the default
// Content-Type to apply if the user hasn't set one. Empty body returns nil.
func buildBody(b model.Body) (io.Reader, string) {
	if b.Content == "" {
		return nil, ""
	}
	switch b.Type {
	case model.BodyJSON:
		return strings.NewReader(b.Content), "application/json"
	case model.BodyText:
		return strings.NewReader(b.Content), "text/plain"
	case model.BodyURLEnc:
		return strings.NewReader(b.Content), "application/x-www-form-urlencoded"
	default:
		return nil, ""
	}
}
