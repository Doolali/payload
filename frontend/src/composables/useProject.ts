// Single shared, reactive state for the app.
//
// Sessions are global and live in their own file on disk; Projects (with
// their nested Collections) are a separate concern. The composable owns two
// independent debounce timers — one for projects, one for sessions — so a
// session edit doesn't trigger a project write and vice versa.
import {computed, reactive} from 'vue';
import {
    ListSessions,
    LoadProject,
    SaveProject,
    SaveSessions,
} from '../../wailsjs/go/main/App';
import {model} from '../../wailsjs/go/models';

export type Selection =
    | {kind: 'session'; sessionId: string}
    | {kind: 'request'; collectionId: string; requestId: string}
    | {kind: 'project-vars'}
    | {kind: 'collection-vars'; collectionId: string}
    | null;

interface State {
    // Global sessions, available regardless of which (if any) project is open.
    sessions: model.Session[];
    sessionsLoaded: boolean;

    // The currently-open project (optional).
    project: model.Project | null;

    selection: Selection;
    // Last response for collection requests (which don't persist responses
    // the way sessions do).
    transientResponse: model.Response | null;
    saving: boolean;
    lastError: string | null;
}

const state = reactive<State>({
    sessions: [],
    sessionsLoaded: false,
    project: null,
    selection: null,
    transientResponse: null,
    saving: false,
    lastError: null,
});

let projectSaveTimer: number | undefined;
let sessionsSaveTimer: number | undefined;

const selectedSession = computed<model.Session | null>(() => {
    const sel = state.selection;
    if (sel?.kind !== 'session') return null;
    return state.sessions.find(s => s.id === sel.sessionId) ?? null;
});

const selectedCollection = computed<model.Collection | null>(() => {
    const sel = state.selection;
    if (!state.project) return null;
    if (sel?.kind === 'collection-vars' || sel?.kind === 'request') {
        return state.project.collections.find(c => c.id === sel.collectionId) ?? null;
    }
    return null;
});

const selectedRequest = computed<model.Request | null>(() => {
    const sel = state.selection;
    if (!state.project || sel?.kind !== 'request') return null;
    const c = state.project.collections.find(c => c.id === sel.collectionId);
    return c?.requests.find(r => r.id === sel.requestId) ?? null;
});

// mergedVars returns the variable scope for whatever is selected. Sessions
// only see project-level vars (when a project is open). Collection requests
// see project ∪ collection vars (collection wins).
const mergedVars = computed<Record<string, string>>(() => {
    const out: Record<string, string> = {};
    if (state.project?.variables) Object.assign(out, state.project.variables);
    if (selectedCollection.value?.variables) Object.assign(out, selectedCollection.value.variables);
    return out;
});

async function loadSessions() {
    try {
        state.sessions = await ListSessions();
        state.sessionsLoaded = true;
    } catch (e: any) {
        state.lastError = String(e);
    }
}

async function loadProject(id: string) {
    cancelProjectSave();
    state.transientResponse = null;
    state.project = await LoadProject(id);
    if (!state.project.variables) state.project.variables = {};
    if (!state.project.collections) state.project.collections = [];
    // Drop any selection that no longer makes sense in the new project.
    const sel = state.selection;
    if (sel?.kind === 'request' || sel?.kind === 'collection-vars') {
        state.selection = null;
    } else if (sel?.kind === 'project-vars') {
        // keep — project-vars is always valid when a project is open
    }
}

function clearProject() {
    cancelProjectSave();
    state.project = null;
    state.transientResponse = null;
    const sel = state.selection;
    if (sel?.kind === 'request' || sel?.kind === 'collection-vars' || sel?.kind === 'project-vars') {
        state.selection = null;
    }
}

function cancelProjectSave() {
    if (projectSaveTimer) {
        clearTimeout(projectSaveTimer);
        projectSaveTimer = undefined;
    }
}

function cancelSessionsSave() {
    if (sessionsSaveTimer) {
        clearTimeout(sessionsSaveTimer);
        sessionsSaveTimer = undefined;
    }
}

function scheduleProjectSave() {
    if (!state.project) return;
    cancelProjectSave();
    projectSaveTimer = window.setTimeout(flushProjectSave, 300);
}

function scheduleSessionsSave() {
    cancelSessionsSave();
    sessionsSaveTimer = window.setTimeout(flushSessionsSave, 300);
}

async function flushProjectSave() {
    if (!state.project) return;
    state.saving = true;
    try {
        await SaveProject(state.project);
    } catch (e: any) {
        state.lastError = String(e);
    } finally {
        state.saving = false;
    }
}

async function flushSessionsSave() {
    state.saving = true;
    try {
        await SaveSessions(state.sessions);
    } catch (e: any) {
        state.lastError = String(e);
    } finally {
        state.saving = false;
    }
}

// scheduleSave is the catch-all autosave the editors call after a mutation.
// It looks at the current selection to decide which scope to persist.
function scheduleSave() {
    const sel = state.selection;
    if (sel?.kind === 'session') {
        scheduleSessionsSave();
    } else {
        scheduleProjectSave();
    }
}

function setResponse(resp: model.Response) {
    const sel = state.selection;
    if (!sel) return;
    if (sel.kind === 'session') {
        const sess = state.sessions.find(s => s.id === sel.sessionId);
        if (sess) {
            sess.lastResponse = resp;
            scheduleSessionsSave();
        }
    } else if (sel.kind === 'request') {
        state.transientResponse = resp;
    }
}

function getResponse(): model.Response | null {
    if (state.selection?.kind === 'session') {
        return selectedSession.value?.lastResponse ?? null;
    }
    if (state.selection?.kind === 'request') {
        return state.transientResponse;
    }
    return null;
}

export function useProject() {
    return {
        state,
        selectedSession,
        selectedRequest,
        selectedCollection,
        mergedVars,
        loadSessions,
        loadProject,
        clearProject,
        scheduleSave,
        scheduleProjectSave,
        scheduleSessionsSave,
        flushProjectSave,
        flushSessionsSave,
        setResponse,
        getResponse,
    };
}
