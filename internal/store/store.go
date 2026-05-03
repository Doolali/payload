// Package store persists Projects to disk as one JSON file per project.
//
// Project files can live anywhere on disk — the user picks a directory when
// they create a project. A small index file at <appDir>/index.json tracks
// the (id, path) pairs so the project picker can list everything regardless
// of where the user saved it. The most recently used directory is kept on
// the index so the next "new project" dialog defaults there.
//
// SaveProject debounces per-project so rapid edits from the UI collapse into
// a single disk write ~250ms after the last change. Flush forces any pending
// writes — call on app shutdown.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"payload/internal/model"
)

const (
	appDirName       = "payload"
	defaultSubdir    = "projects"
	indexFileName    = "index.json"
	sessionsFileName = "sessions.json"
	saveDebounce     = 250 * time.Millisecond
	maxNameAttempt   = 1000
)

type indexEntry struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type indexFile struct {
	LastUsedDir string       `json:"lastUsedDir"`
	Projects    []indexEntry `json:"projects"`
}

type Store struct {
	appDir       string
	defaultDir   string
	indexPath    string
	sessionsPath string

	mu      sync.Mutex
	index   indexFile
	pending map[string]*model.Project
	timers  map[string]*time.Timer

	// Sessions are global rather than nested under a project. Same debounced
	// save pattern as projects but only one timer is needed.
	sessionsPending []model.Session
	sessionsTimer   *time.Timer
	sessionsDirty   bool
}

// New constructs a Store rooted at the user-config app directory and ensures
// the default projects subdirectory and index file exist.
func New() (*Store, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate user config dir: %w", err)
	}
	appDir := filepath.Join(cfg, appDirName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return nil, fmt.Errorf("create app dir: %w", err)
	}
	defaultDir := filepath.Join(appDir, defaultSubdir)
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		return nil, fmt.Errorf("create default projects dir: %w", err)
	}
	s := &Store{
		appDir:       appDir,
		defaultDir:   defaultDir,
		indexPath:    filepath.Join(appDir, indexFileName),
		sessionsPath: filepath.Join(appDir, sessionsFileName),
		pending:      make(map[string]*model.Project),
		timers:       make(map[string]*time.Timer),
	}
	if err := s.loadIndex(); err != nil {
		return nil, fmt.Errorf("load index: %w", err)
	}
	if s.index.LastUsedDir == "" {
		s.index.LastUsedDir = defaultDir
	}
	if err := s.migrateLegacySessions(); err != nil {
		// Migration is best-effort — log via stored error rather than blocking
		// startup. Worst case the user keeps a few duplicate sessions.
		fmt.Fprintf(os.Stderr, "session migration failed: %v\n", err)
	}
	return s, nil
}

// migrateLegacySessions hoists any sessions stored under the legacy
// per-project layout into the global sessions file, then drops them from
// the project files on next save (the new Project struct has no Sessions
// field, so the next Marshal of each project will exclude them).
func (s *Store) migrateLegacySessions() error {
	s.mu.Lock()
	entries := append([]indexEntry(nil), s.index.Projects...)
	s.mu.Unlock()

	current, err := s.LoadSessions()
	if err != nil {
		return err
	}
	added := 0
	for _, e := range entries {
		raw, err := os.ReadFile(e.Path)
		if err != nil {
			continue
		}
		var doc struct {
			Sessions []model.Session `json:"sessions"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			continue
		}
		if len(doc.Sessions) > 0 {
			current = append(current, doc.Sessions...)
			added += len(doc.Sessions)
		}
	}
	if added == 0 {
		return nil
	}
	return s.writeSessions(current)
}

func (s *Store) loadIndex() error {
	raw, err := os.ReadFile(s.indexPath)
	if errors.Is(err, os.ErrNotExist) {
		s.index = indexFile{}
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &s.index)
}

// saveIndexLocked persists the index. Caller must hold s.mu.
func (s *Store) saveIndexLocked() error {
	raw, err := json.MarshalIndent(s.index, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.appDir, "index.*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, s.indexPath)
}

// AppDir is the root payload config directory (parent of the default project
// folder and the index).
func (s *Store) AppDir() string {
	return s.appDir
}

// DefaultProjectDir returns the directory to suggest when the user is asked
// where to save a new project. Initially the in-app default; thereafter the
// directory of the most recently created project.
func (s *Store) DefaultProjectDir() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.index.LastUsedDir != "" {
		return s.index.LastUsedDir
	}
	return s.defaultDir
}

func (s *Store) lookupPath(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.index.Projects {
		if e.ID == id {
			return e.Path, true
		}
	}
	return "", false
}

// ListProjects returns project summaries sorted by most recently updated.
// Index entries whose files are missing or unreadable are skipped — they will
// surface as load errors if the user clicks them, but should not break the
// whole picker.
func (s *Store) ListProjects() ([]model.ProjectSummary, error) {
	s.mu.Lock()
	entries := append([]indexEntry(nil), s.index.Projects...)
	s.mu.Unlock()

	out := make([]model.ProjectSummary, 0, len(entries))
	for _, e := range entries {
		raw, err := os.ReadFile(e.Path)
		if err != nil {
			continue
		}
		var summary model.ProjectSummary
		if err := json.Unmarshal(raw, &summary); err != nil {
			continue
		}
		if summary.ID == "" {
			summary.ID = e.ID
		}
		out = append(out, summary)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

// LoadProject reads and parses a single project file by id. If a debounced
// save is still pending for the project, the in-memory snapshot is returned
// instead of the on-disk copy so the UI sees its own writes after switching
// away and back.
func (s *Store) LoadProject(id string) (*model.Project, error) {
	if id == "" {
		return nil, errors.New("project id is required")
	}
	s.mu.Lock()
	if pending, ok := s.pending[id]; ok {
		s.mu.Unlock()
		return cloneProject(pending)
	}
	s.mu.Unlock()
	path, ok := s.lookupPath(id)
	if !ok {
		return nil, fmt.Errorf("project %s not in index", id)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read project: %w", err)
	}
	var p model.Project
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("parse project: %w", err)
	}
	return &p, nil
}

// cloneProject deep-copies a project via json round-trip so a returned value
// can never alias the in-memory pending snapshot.
func cloneProject(p *model.Project) (*model.Project, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("clone marshal: %w", err)
	}
	var c model.Project
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("clone unmarshal: %w", err)
	}
	return &c, nil
}

var slugRE = regexp.MustCompile(`[^a-z0-9]+`)

func slug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRE.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "project"
	}
	return s
}

// uniqueFilename picks <slug>.json in dir, falling back to <slug>-2.json etc
// if a file already exists. The final fallback uses a short uuid suffix in
// the unlikely case all numeric suffixes are taken.
func uniqueFilename(dir, base string) string {
	candidate := filepath.Join(dir, base+".json")
	if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
		return candidate
	}
	for i := 2; i < maxNameAttempt; i++ {
		candidate = filepath.Join(dir, fmt.Sprintf("%s-%d.json", base, i))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%s.json", base, uuid.NewString()[:8]))
}

// CreateProject creates a fresh Project with the given name in the chosen
// directory. An empty dir falls back to DefaultProjectDir. The directory is
// created if it doesn't yet exist, and is recorded as the new default.
func (s *Store) CreateProject(name, dir string) (*model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = s.DefaultProjectDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	now := time.Now().UTC()
	p := &model.Project{
		ID:          uuid.NewString(),
		Name:        name,
		Variables:   map[string]string{},
		Collections: []model.Collection{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	path := uniqueFilename(dir, slug(name))
	if err := s.writeProjectAt(p, path); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.index.Projects = append(s.index.Projects, indexEntry{ID: p.ID, Path: path})
	s.index.LastUsedDir = dir
	err := s.saveIndexLocked()
	s.mu.Unlock()
	if err != nil {
		// Best-effort cleanup so we don't leave an orphan file when the index
		// fails to record the new project.
		_ = os.Remove(path)
		return nil, fmt.Errorf("save index: %w", err)
	}
	return p, nil
}

// RenameProject updates a project's name and persists it. The file path is
// not changed — renaming on disk would invalidate any external references.
func (s *Store) RenameProject(id, name string) (*model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}
	p, err := s.LoadProject(id)
	if err != nil {
		return nil, err
	}
	p.Name = name
	p.UpdatedAt = time.Now().UTC()
	path, ok := s.lookupPath(id)
	if !ok {
		return nil, fmt.Errorf("project %s not in index", id)
	}
	if err := s.writeProjectAt(p, path); err != nil {
		return nil, err
	}
	return p, nil
}

// DeleteProject removes a project from the index and deletes its file. Any
// pending debounced save for that project is cancelled.
func (s *Store) DeleteProject(id string) error {
	if id == "" {
		return errors.New("project id is required")
	}
	s.mu.Lock()
	if t, ok := s.timers[id]; ok {
		t.Stop()
		delete(s.timers, id)
		delete(s.pending, id)
	}
	var path string
	for i, e := range s.index.Projects {
		if e.ID == id {
			path = e.Path
			s.index.Projects = append(s.index.Projects[:i], s.index.Projects[i+1:]...)
			break
		}
	}
	err := s.saveIndexLocked()
	s.mu.Unlock()
	if err != nil {
		return fmt.Errorf("save index: %w", err)
	}
	if path != "" {
		if rmErr := os.Remove(path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			return fmt.Errorf("delete file: %w", rmErr)
		}
	}
	return nil
}

// SaveProject schedules a debounced write of the project. Multiple calls
// within the debounce window collapse into a single disk write of the latest
// snapshot.
func (s *Store) SaveProject(p *model.Project) error {
	if p == nil || p.ID == "" {
		return errors.New("project with id is required")
	}
	if _, ok := s.lookupPath(p.ID); !ok {
		return fmt.Errorf("project %s not in index", p.ID)
	}
	p.UpdatedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[p.ID] = p
	if t, ok := s.timers[p.ID]; ok {
		t.Stop()
	}
	id := p.ID
	s.timers[id] = time.AfterFunc(saveDebounce, func() {
		_ = s.flushOne(id)
	})
	return nil
}

// Flush writes any pending debounced saves immediately. Call on shutdown.
func (s *Store) Flush() error {
	s.mu.Lock()
	ids := make([]string, 0, len(s.timers))
	for id, t := range s.timers {
		t.Stop()
		ids = append(ids, id)
	}
	if s.sessionsTimer != nil {
		s.sessionsTimer.Stop()
		s.sessionsTimer = nil
	}
	pendingSessions := s.sessionsPending
	dirty := s.sessionsDirty
	s.sessionsPending = nil
	s.sessionsDirty = false
	s.mu.Unlock()

	var firstErr error
	for _, id := range ids {
		if err := s.flushOne(id); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if dirty {
		if err := s.writeSessions(pendingSessions); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *Store) flushOne(id string) error {
	s.mu.Lock()
	p, ok := s.pending[id]
	if ok {
		delete(s.pending, id)
		delete(s.timers, id)
	}
	s.mu.Unlock()
	if !ok || p == nil {
		return nil
	}
	path, ok := s.lookupPath(id)
	if !ok {
		return fmt.Errorf("project %s not in index", id)
	}
	return s.writeProjectAt(p, path)
}

// LoadSessions reads the global sessions file. Missing file is treated as
// an empty list. The pending in-memory snapshot wins over the on-disk copy
// so the UI sees its own writes after restart-style reloads.
func (s *Store) LoadSessions() ([]model.Session, error) {
	s.mu.Lock()
	if s.sessionsDirty {
		out := append([]model.Session(nil), s.sessionsPending...)
		s.mu.Unlock()
		return out, nil
	}
	s.mu.Unlock()
	raw, err := os.ReadFile(s.sessionsPath)
	if errors.Is(err, os.ErrNotExist) {
		return []model.Session{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read sessions: %w", err)
	}
	var out []model.Session
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}
	if out == nil {
		out = []model.Session{}
	}
	return out, nil
}

// SaveSessions schedules a debounced write of the global sessions file.
// Multiple calls within the debounce window collapse into one write of the
// latest snapshot.
func (s *Store) SaveSessions(sessions []model.Session) error {
	if sessions == nil {
		sessions = []model.Session{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionsPending = append(sessions[:0:0], sessions...)
	s.sessionsDirty = true
	if s.sessionsTimer != nil {
		s.sessionsTimer.Stop()
	}
	s.sessionsTimer = time.AfterFunc(saveDebounce, func() {
		s.mu.Lock()
		snap := s.sessionsPending
		s.sessionsPending = nil
		s.sessionsDirty = false
		s.sessionsTimer = nil
		s.mu.Unlock()
		_ = s.writeSessions(snap)
	})
	return nil
}

// writeSessions persists the sessions list atomically.
func (s *Store) writeSessions(sessions []model.Session) error {
	if sessions == nil {
		sessions = []model.Session{}
	}
	raw, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sessions: %w", err)
	}
	tmp, err := os.CreateTemp(s.appDir, "sessions.*.tmp")
	if err != nil {
		return fmt.Errorf("create temp sessions: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp sessions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp sessions: %w", err)
	}
	if err := os.Rename(tmpName, s.sessionsPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename sessions: %w", err)
	}
	return nil
}

// writeProjectAt persists the project atomically: write to a temp file in
// the destination directory, then rename. This avoids leaving truncated
// files if the process dies mid-write.
func (s *Store) writeProjectAt(p *model.Project, path string) error {
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal project: %w", err)
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}
