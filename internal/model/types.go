// Package model defines the persisted data types for the app.
//
// A Project is the top-level container. It owns Collections (folders of
// Requests), Sessions (ad-hoc requests preserved across app restarts), and
// Variables. Variables are also defined per-Collection; collection variables
// shadow project variables of the same name when a request is sent.
package model

import "time"

type Method string

const (
	MethodGET     Method = "GET"
	MethodPOST    Method = "POST"
	MethodPUT     Method = "PUT"
	MethodDELETE  Method = "DELETE"
	MethodPATCH   Method = "PATCH"
	MethodHEAD    Method = "HEAD"
	MethodOPTIONS Method = "OPTIONS"
)

type BodyType string

const (
	BodyNone   BodyType = "none"
	BodyJSON   BodyType = "json"
	BodyText   BodyType = "text"
	BodyForm   BodyType = "form"
	BodyURLEnc BodyType = "urlencoded"
)

type KV struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

type Body struct {
	Type    BodyType `json:"type"`
	Content string   `json:"content"`
}

type Request struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Method      Method    `json:"method"`
	URL         string    `json:"url"`
	Headers     []KV      `json:"headers"`
	QueryParams []KV      `json:"queryParams"`
	Body        Body      `json:"body"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Collection struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	Requests  []Request         `json:"requests"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// Response holds the result of executing a request. Stored on Sessions so a
// user can scroll back through past results in the same way as a chat log.
type Response struct {
	Status     int       `json:"status"`
	StatusText string    `json:"statusText"`
	Headers    []KV      `json:"headers"`
	Body       string    `json:"body"`
	DurationMs int64     `json:"durationMs"`
	SentAt     time.Time `json:"sentAt"`
	Error      string    `json:"error,omitempty"`
}

// Session is a Request kept on the project's session list. Sessions persist
// across app restarts and can be promoted into a Collection.
type Session struct {
	Request
	LastResponse *Response `json:"lastResponse,omitempty"`
}

type Project struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Variables   map[string]string `json:"variables"`
	Collections []Collection      `json:"collections"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// ProjectSummary is the lightweight shape used by the project picker.
type ProjectSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updatedAt"`
}
