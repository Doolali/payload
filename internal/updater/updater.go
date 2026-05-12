// Package updater queries the project's GitHub releases and applies updates
// by downloading the platform-appropriate asset and launching the native
// installer. The current process is expected to exit shortly after launching
// the installer so it can replace the running binary cleanly.
package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	githubAPI = "https://api.github.com"
	userAgent = "payload-updater"
)

// Release is the subset of GitHub's /releases/latest payload we care about.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	Assets      []Asset   `json:"assets"`
}

// Asset is one downloadable file attached to a release.
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Info is the consolidated update status surfaced to the UI.
type Info struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	AssetURL       string `json:"assetUrl"`
	AssetName      string `json:"assetName"`
	Notes          string `json:"notes"`
	// CanAutoInstall is true when the platform's asset can be launched from
	// within the app (currently only Windows MSI). When false the UI should
	// fall back to opening ReleaseURL in the browser.
	CanAutoInstall bool `json:"canAutoInstall"`
}

// LatestRelease fetches the most recently published release for owner/repo.
func LatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPI, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github api %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	return &rel, nil
}

// IsNewer reports whether `latest` is strictly greater than `current` using
// dotted-numeric comparison (so 1.0.10 > 1.0.9). A current version of "" or
// "dev" always returns false — dev builds shouldn't pester to upgrade.
func IsNewer(latest, current string) bool {
	if current == "" || current == "dev" {
		return false
	}
	return cmpVersion(strings.TrimPrefix(latest, "v"), strings.TrimPrefix(current, "v")) > 0
}

func cmpVersion(a, b string) int {
	ap, bp := splitNum(a), splitNum(b)
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		var ai, bi int
		if i < len(ap) {
			ai = ap[i]
		}
		if i < len(bp) {
			bi = bp[i]
		}
		if ai != bi {
			if ai > bi {
				return 1
			}
			return -1
		}
	}
	return 0
}

func splitNum(v string) []int {
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out
		}
		out = append(out, n)
	}
	return out
}

// AssetForPlatform picks the release asset that best matches the running
// OS/arch. Returns nil if no asset is suitable.
func AssetForPlatform(rel *Release, goos, goarch string) *Asset {
	var prefer []string
	switch goos {
	case "windows":
		prefer = []string{".msi", ".zip", ".exe"}
	case "darwin":
		prefer = []string{".dmg", ".pkg", ".zip"}
	case "linux":
		prefer = []string{".appimage", ".tar.gz", ".tar.xz"}
	default:
		return nil
	}

	for _, ext := range prefer {
		for i := range rel.Assets {
			a := &rel.Assets[i]
			n := strings.ToLower(a.Name)
			if !strings.HasSuffix(n, ext) {
				continue
			}
			if !strings.Contains(n, goos) {
				continue
			}
			// macOS releases are typically published as a single "universal"
			// asset rather than per-arch; accept that as a match.
			if !strings.Contains(n, goarch) && !(goos == "darwin" && strings.Contains(n, "universal")) {
				continue
			}
			return a
		}
	}
	return nil
}

// CanAutoInstall reports whether the named asset can be launched in-process
// to perform the install. Currently only Windows MSI is supported.
func CanAutoInstall(assetName string) bool {
	return runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(assetName), ".msi")
}

// Download fetches asset.BrowserDownloadURL into a fresh temp directory and
// returns the absolute path of the downloaded file.
func Download(ctx context.Context, asset *Asset) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: %s", asset.BrowserDownloadURL, resp.Status)
	}

	dir, err := os.MkdirTemp("", "payload-update-*")
	if err != nil {
		return "", err
	}
	out := filepath.Join(dir, asset.Name)
	f, err := os.Create(out)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return out, nil
}

// LaunchInstaller spawns the platform's installer for the downloaded file.
// On Windows this runs `msiexec /i path /passive /norestart`, which triggers
// the MSI's built-in UAC elevation. The call returns immediately; the caller
// should quit the app shortly after so the installer can replace the binary.
func LaunchInstaller(path string) error {
	if runtime.GOOS != "windows" {
		return errors.New("auto-install is only supported on Windows; download the asset from the release page")
	}
	cmd := exec.Command("msiexec", "/i", path, "/passive", "/norestart")
	return cmd.Start()
}
