package main

import (
	"context"
	"runtime"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"payload/internal/httpclient"
	"payload/internal/model"
	"payload/internal/store"
	"payload/internal/updater"
)

// Repo coordinates for the GitHub releases the updater checks against.
const (
	updateOwner = "Doolali"
	updateRepo  = "payload"
)

// App is bound into the Vue runtime. Each exported method becomes callable
// from TypeScript via the generated wailsjs bindings.
type App struct {
	ctx     context.Context
	store   *store.Store
	version string
}

func NewApp(s *store.Store, version string) *App {
	return &App{store: s, version: version}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.silentUpdateCheck()
}

// silentUpdateCheck runs once shortly after launch. If a newer release exists
// it emits "update:available" so the frontend can offer to install. Failures
// (offline, rate limit, etc.) are intentionally swallowed — a missed check is
// never worth surfacing as an error.
func (a *App) silentUpdateCheck() {
	select {
	case <-a.ctx.Done():
		return
	case <-time.After(5 * time.Second):
	}
	info := a.CheckForUpdate()
	if info.Available {
		wruntime.EventsEmit(a.ctx, "update:available", info)
	}
}

// shutdown flushes any pending debounced saves so a clean window close never
// loses the user's last edit.
func (a *App) shutdown(ctx context.Context) {
	_ = a.store.Flush()
}

// AppDir returns the absolute path of the payload config directory (where the
// index lives and the default project folder sits).
func (a *App) AppDir() string {
	return a.store.AppDir()
}

// DefaultProjectDir returns the directory the new-project dialog should
// pre-fill: the most recently used location, or the in-app default.
func (a *App) DefaultProjectDir() string {
	return a.store.DefaultProjectDir()
}

// ChooseProjectDir opens a native directory picker so the user can decide
// where their new project's JSON file should live. An empty return value
// means the user cancelled.
func (a *App) ChooseProjectDir() (string, error) {
	return wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:            "Choose where to save the project",
		DefaultDirectory: a.store.DefaultProjectDir(),
	})
}

func (a *App) ListProjects() ([]model.ProjectSummary, error) {
	return a.store.ListProjects()
}

// CreateProject creates a project and stores its file at <dir>/<slug>.json.
// An empty dir falls back to the default location.
func (a *App) CreateProject(name, dir string) (*model.Project, error) {
	return a.store.CreateProject(name, dir)
}

// ProjectPath returns the absolute file path of a project, or empty if the
// project isn't in the index.
func (a *App) ProjectPath(id string) string {
	return a.store.ProjectPath(id)
}

// MoveProject opens a directory picker and moves the project's JSON file
// into the chosen directory. Returns nil project (no error) if the user
// cancels the dialog.
func (a *App) MoveProject(id string) (*model.Project, error) {
	dir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:            "Move project file to…",
		DefaultDirectory: a.store.DefaultProjectDir(),
	})
	if err != nil {
		return nil, err
	}
	if dir == "" {
		return nil, nil
	}
	return a.store.MoveProjectFile(id, dir)
}

// OpenProjectFile pops a native file picker so the user can locate an
// existing project JSON file, then registers it with the index. Returns
// nil project (and nil error) if the user cancelled.
func (a *App) OpenProjectFile() (*model.Project, error) {
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:            "Open project file",
		DefaultDirectory: a.store.DefaultProjectDir(),
		Filters: []wruntime.FileFilter{
			{DisplayName: "Project JSON (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	return a.store.OpenProjectFile(path)
}

func (a *App) LoadProject(id string) (*model.Project, error) {
	return a.store.LoadProject(id)
}

func (a *App) RenameProject(id, name string) (*model.Project, error) {
	return a.store.RenameProject(id, name)
}

func (a *App) DeleteProject(id string) error {
	return a.store.DeleteProject(id)
}

// SaveProject persists the entire project. The frontend owns the project
// state while it's selected; this binding accepts the latest snapshot and
// schedules a debounced disk write. The saved snapshot is returned so the
// frontend can pick up any backend-side adjustments (e.g. UpdatedAt).
func (a *App) SaveProject(p model.Project) (*model.Project, error) {
	if err := a.store.SaveProject(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListSessions returns the global session list. Sessions are not tied to a
// project — they're a scratchpad available regardless of which project (if
// any) is currently open.
func (a *App) ListSessions() ([]model.Session, error) {
	return a.store.LoadSessions()
}

// GetUIState returns the last-saved UI snapshot (which tab / project /
// selection was active when the app last shut down).
func (a *App) GetUIState() store.UIState {
	return a.store.UIState()
}

// SaveUIState persists the current UI snapshot so the next launch can
// restore it.
func (a *App) SaveUIState(ui store.UIState) error {
	return a.store.SaveUIState(ui)
}

// SaveSessions persists the global session list. Frontend-driven; the same
// debounce-and-flush pattern as projects.
func (a *App) SaveSessions(sessions []model.Session) error {
	return a.store.SaveSessions(sessions)
}

// SendRequest performs the HTTP request described by r, applying {{name}}
// substitution from vars before dialing. Network or parse failures populate
// Response.Error so the UI always receives a structured result.
func (a *App) SendRequest(r model.Request, vars map[string]string) model.Response {
	ctx, cancel := context.WithTimeout(a.ctx, 65*time.Second)
	defer cancel()
	return httpclient.Send(ctx, r, vars)
}

// AppVersion returns the version baked into this build. Dev builds report
// "dev" so the updater can short-circuit comparisons.
func (a *App) AppVersion() string {
	return a.version
}

// CheckForUpdate queries GitHub for the latest release and returns a flat
// status struct the UI can render directly. Network errors are reported via
// Info.Notes so the frontend can show a single uniform "couldn't check"
// message instead of plumbing errors through Wails' promise rejection.
func (a *App) CheckForUpdate() updater.Info {
	info := updater.Info{CurrentVersion: a.version}

	ctx, cancel := context.WithTimeout(a.ctx, 20*time.Second)
	defer cancel()

	rel, err := updater.LatestRelease(ctx, updateOwner, updateRepo)
	if err != nil {
		info.Notes = "Couldn't reach GitHub to check for updates."
		return info
	}

	info.LatestVersion = rel.TagName
	info.ReleaseURL = rel.HTMLURL
	info.Notes = rel.Body

	if !updater.IsNewer(rel.TagName, a.version) {
		return info
	}
	info.Available = true

	if asset := updater.AssetForPlatform(rel, runtime.GOOS, runtime.GOARCH); asset != nil {
		info.AssetURL = asset.BrowserDownloadURL
		info.AssetName = asset.Name
		info.CanAutoInstall = updater.CanAutoInstall(asset.Name)
	}
	return info
}

// DownloadAndInstallUpdate downloads the asset at downloadURL into a temp
// directory, launches the platform installer, and quits the app a moment
// later so the installer can replace the running binary cleanly.
func (a *App) DownloadAndInstallUpdate(downloadURL, fileName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	asset := &updater.Asset{
		Name:               fileName,
		BrowserDownloadURL: downloadURL,
	}
	path, err := updater.Download(ctx, asset)
	if err != nil {
		return err
	}
	if err := updater.LaunchInstaller(path); err != nil {
		return err
	}
	// Give Windows a beat to surface the UAC dialog before our process exits.
	// Once we're gone, msiexec can replace the binary without "files in use".
	go func() {
		time.Sleep(500 * time.Millisecond)
		wruntime.Quit(a.ctx)
	}()
	return nil
}
