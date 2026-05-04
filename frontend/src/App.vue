<script setup lang="ts">
// Root layout. Tabbed left sidebar (Projects | Sessions) plus a main pane
// that renders whichever editor matches the current selection. Sessions are
// global and load once on mount; projects load on demand when picked.
//
// On launch we restore the last-used tab / project / selection from the
// backend's UI snapshot, and re-save it (debounced) whenever any of those
// change so a relaunch puts the user back where they were.
import {onMounted, ref, watch} from 'vue';
import type {model} from '../wailsjs/go/models';
import {
    CreateProject,
    DeleteProject,
    GetUIState,
    ListProjects,
    SaveUIState,
} from '../wailsjs/go/main/App';
import ProjectsTab from './components/ProjectsTab.vue';
import SessionsTab from './components/SessionsTab.vue';
import MainPane from './components/MainPane.vue';
import PromptDialog from './components/PromptDialog.vue';
import ConfirmDialog from './components/ConfirmDialog.vue';
import {useProject, type Selection} from './composables/useProject';
import logo from './assets/images/logo.png';

type Tab = 'projects' | 'sessions';

const {state, loadProject, loadSessions, clearProject, flushProjectSave} = useProject();

const projects = ref<model.ProjectSummary[]>([]);
const error = ref<string | null>(null);
const tab = ref<Tab>('sessions');
const restored = ref(false);
let uiSaveTimer: number | undefined;

async function refresh() {
    try {
        projects.value = await ListProjects();
    } catch (e: any) {
        error.value = String(e);
    }
}

async function selectProject(id: string) {
    if (state.project?.id === id) return;
    await flushProjectSave();
    try {
        await loadProject(id);
    } catch (e: any) {
        error.value = String(e);
    }
}

async function createProject(payload: {name: string; dir: string}) {
    try {
        const p = await CreateProject(payload.name, payload.dir);
        await refresh();
        await loadProject(p.id);
    } catch (e: any) {
        error.value = String(e);
    }
}

async function openedProject(id: string) {
    try {
        await refresh();
        await loadProject(id);
    } catch (e: any) {
        error.value = String(e);
    }
}

async function deleteProject(id: string) {
    try {
        await DeleteProject(id);
        if (state.project?.id === id) clearProject();
        await refresh();
    } catch (e: any) {
        error.value = String(e);
    }
}

watch(() => state.lastError, (msg) => {
    if (msg) {
        error.value = msg;
        state.lastError = null;
    }
});

// Reconstruct a Selection from the flat UISelection shape returned by Go.
// Returns null for empty/unknown kinds.
function selectionFromUI(s: any): Selection {
    if (!s || !s.kind) return null;
    switch (s.kind) {
        case 'session': return {kind: 'session', sessionId: s.sessionId};
        case 'request': return {kind: 'request', collectionId: s.collectionId, requestId: s.requestId};
        case 'project-vars': return {kind: 'project-vars'};
        case 'collection-vars': return {kind: 'collection-vars', collectionId: s.collectionId};
        default: return null;
    }
}

// Inverse of selectionFromUI — flatten a Selection for the backend snapshot.
function uiFromSelection(sel: Selection): any {
    if (!sel) return {};
    switch (sel.kind) {
        case 'session': return {kind: 'session', sessionId: sel.sessionId};
        case 'request': return {kind: 'request', collectionId: sel.collectionId, requestId: sel.requestId};
        case 'project-vars': return {kind: 'project-vars'};
        case 'collection-vars': return {kind: 'collection-vars', collectionId: sel.collectionId};
    }
}

// Drop a restored selection that no longer points at a real entity (the
// session was deleted, the collection was renamed, etc).
function selectionStillValid(sel: Selection): boolean {
    if (!sel) return false;
    if (sel.kind === 'session') return state.sessions.some(s => s.id === sel.sessionId);
    if (!state.project) return false;
    if (sel.kind === 'project-vars') return true;
    const c = state.project.collections.find(c => c.id === sel.collectionId);
    if (!c) return false;
    if (sel.kind === 'collection-vars') return true;
    return c.requests.some(r => r.id === sel.requestId);
}

async function restoreUI() {
    let ui: any = null;
    try { ui = await GetUIState(); } catch { /* fresh install */ }
    if (ui?.projectId) {
        try { await loadProject(ui.projectId); }
        catch { /* project file gone — fall through */ }
    }
    if (ui?.tab === 'projects' || ui?.tab === 'sessions') {
        tab.value = ui.tab;
    }
    const sel = selectionFromUI(ui?.selection);
    if (sel && selectionStillValid(sel)) {
        state.selection = sel;
    }
    restored.value = true;
}

function scheduleUISave() {
    if (!restored.value) return;
    if (uiSaveTimer) clearTimeout(uiSaveTimer);
    uiSaveTimer = window.setTimeout(async () => {
        try {
            await SaveUIState({
                tab: tab.value,
                projectId: state.project?.id ?? '',
                selection: uiFromSelection(state.selection),
            } as any);
        } catch { /* non-fatal */ }
    }, 250);
}

watch([tab, () => state.project?.id ?? null, () => state.selection], scheduleUISave, {deep: true});

onMounted(async () => {
    await Promise.all([refresh(), loadSessions()]);
    await restoreUI();
});
</script>

<template>
    <div class="layout">
        <aside class="sidebar">
            <header class="brand">
                <img :src="logo" alt="payload" class="logo" />
            </header>
            <nav class="tabs">
                <button
                    :class="{active: tab === 'sessions'}"
                    @click="tab = 'sessions'"
                >Sessions</button>
                <button
                    :class="{active: tab === 'projects'}"
                    @click="tab = 'projects'"
                >Projects</button>
            </nav>
            <SessionsTab v-if="tab === 'sessions'" />
            <ProjectsTab
                v-else
                :projects="projects"
                :active-id="state.project?.id ?? null"
                @select="selectProject"
                @create="createProject"
                @delete="deleteProject"
                @opened="openedProject"
            />
        </aside>
        <MainPane />
        <div v-if="error" class="error" @click="error = null">
            {{ error }}
            <span class="dismiss">click to dismiss</span>
        </div>
        <PromptDialog />
        <ConfirmDialog />
    </div>
</template>

<style scoped>
.layout {
    display: flex;
    flex: 1;
    min-height: 0;
    position: relative;
}

.sidebar {
    width: 280px;
    background: var(--bg-elev);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    flex-shrink: 0;
}

.brand {
    padding: 12px 12px 8px;
}

.logo {
    height: 32px;
    width: auto;
    image-rendering: pixelated;
    border-radius: 4px;
    display: block;
}

.tabs {
    display: flex;
    border-bottom: 1px solid var(--border);
}

.tabs button {
    flex: 1;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-dim);
    padding: 8px 12px;
    cursor: pointer;
    font-size: 13px;
    border-radius: 0;
    font-weight: 500;
}

.tabs button:hover:not(:disabled) {
    color: var(--text);
}

.tabs button.active {
    color: var(--text);
    border-bottom-color: var(--accent);
}

.error {
    position: absolute;
    bottom: 16px;
    right: 16px;
    background: var(--danger);
    color: #fff;
    padding: 10px 14px;
    border-radius: 6px;
    max-width: 360px;
    cursor: pointer;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
    z-index: 10;
}

.dismiss {
    display: block;
    font-size: 11px;
    opacity: 0.8;
    margin-top: 4px;
}
</style>
