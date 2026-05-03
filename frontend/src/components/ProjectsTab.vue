<script setup lang="ts">
// Projects tab content for the left sidebar. Top: list of projects with
// + New / open / delete. Bottom: when a project is open, a Vars button and
// the collection tree for that project. Sessions live in their own tab.
import {computed, ref} from 'vue';
import {model} from '../../wailsjs/go/models';
import {useProject} from '../composables/useProject';
import NewProjectDialog from './NewProjectDialog.vue';
import CollectionTree from './CollectionTree.vue';

const {state, scheduleProjectSave} = useProject();

const props = defineProps<{
    projects: model.ProjectSummary[];
    activeId: string | null;
}>();

const emit = defineEmits<{
    (e: 'select', id: string): void;
    (e: 'create', payload: {name: string; dir: string}): void;
    (e: 'delete', id: string): void;
}>();

const showNew = ref(false);

const sorted = computed(() =>
    [...props.projects].sort((a, b) => a.name.localeCompare(b.name))
);

function confirmDelete(p: model.ProjectSummary) {
    if (window.confirm(`Delete project "${p.name}"? This cannot be undone.`)) {
        emit('delete', p.id);
    }
}

function openProjectVars() {
    state.selection = {kind: 'project-vars'};
}

function isProjectVars(): boolean {
    return state.selection?.kind === 'project-vars';
}

function newCollection() {
    if (!state.project) return;
    const name = window.prompt('Collection name', 'New collection');
    if (!name?.trim()) return;
    const now = new Date().toISOString() as any;
    const c: model.Collection = {
        id: crypto.randomUUID(),
        name: name.trim(),
        variables: {},
        requests: [],
        createdAt: now,
        updatedAt: now,
    } as any;
    state.project.collections.push(c);
    scheduleProjectSave();
}

function newRequest(collectionId: string) {
    if (!state.project) return;
    const c = state.project.collections.find(c => c.id === collectionId);
    if (!c) return;
    const now = new Date().toISOString() as any;
    const r: model.Request = {
        id: crypto.randomUUID(),
        name: 'New request',
        method: 'GET',
        url: '',
        headers: [],
        queryParams: [],
        body: {type: 'none', content: ''},
        createdAt: now,
        updatedAt: now,
    } as any;
    c.requests.push(r);
    state.selection = {kind: 'request', collectionId: c.id, requestId: r.id};
    state.transientResponse = null;
    scheduleProjectSave();
}

function selectRequest(collectionId: string, requestId: string) {
    state.selection = {kind: 'request', collectionId, requestId};
    state.transientResponse = null;
}

function selectCollectionVars(collectionId: string) {
    state.selection = {kind: 'collection-vars', collectionId};
}

function renameCollection(collectionId: string, name: string) {
    if (!state.project) return;
    const c = state.project.collections.find(c => c.id === collectionId);
    if (c) {
        c.name = name;
        scheduleProjectSave();
    }
}

function deleteCollection(collectionId: string) {
    if (!state.project) return;
    state.project.collections = state.project.collections.filter(c => c.id !== collectionId);
    const sel = state.selection;
    if (sel && (sel.kind === 'request' || sel.kind === 'collection-vars') && sel.collectionId === collectionId) {
        state.selection = null;
    }
    scheduleProjectSave();
}

function deleteRequest(collectionId: string, requestId: string) {
    if (!state.project) return;
    const c = state.project.collections.find(c => c.id === collectionId);
    if (!c) return;
    c.requests = c.requests.filter(r => r.id !== requestId);
    const sel = state.selection;
    if (sel?.kind === 'request' && sel.requestId === requestId) {
        state.selection = null;
    }
    scheduleProjectSave();
}
</script>

<template>
    <div class="projects-tab">
        <div class="actions-row">
            <button class="primary full" @click="showNew = true">+ New project</button>
        </div>

        <ul v-if="sorted.length" class="project-list">
            <li
                v-for="p in sorted"
                :key="p.id"
                :class="{active: p.id === activeId}"
                @click="emit('select', p.id)"
            >
                <span class="name">{{ p.name }}</span>
                <span v-if="p.id === activeId" class="open-badge">open</span>
                <button
                    class="ghost delete"
                    title="Delete project"
                    @click.stop="confirmDelete(p)"
                >×</button>
            </li>
        </ul>
        <div v-else class="empty">
            <p>No projects yet.</p>
            <button class="primary" @click="showNew = true">+ New project</button>
        </div>

        <div v-if="state.project" class="active-section">
            <header>
                <span class="title">{{ state.project.name }}</span>
                <button
                    class="ghost vars-btn"
                    :class="{active: isProjectVars()}"
                    title="Edit project variables"
                    @click="openProjectVars"
                >Vars</button>
                <span v-if="state.saving" class="saving">saving…</span>
            </header>
            <CollectionTree
                :collections="state.project.collections"
                :selection="state.selection"
                @new-collection="newCollection"
                @new-request="newRequest"
                @select-request="selectRequest"
                @select-vars="selectCollectionVars"
                @rename-collection="renameCollection"
                @delete-collection="deleteCollection"
                @delete-request="deleteRequest"
            />
        </div>

        <NewProjectDialog
            v-if="showNew"
            @cancel="showNew = false"
            @create="(payload) => { emit('create', payload); showNew = false; }"
        />
    </div>
</template>

<style scoped>
.projects-tab {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    flex: 1;
}

.actions-row {
    padding: 8px 12px;
    border-bottom: 1px solid var(--border);
}

.full {
    width: 100%;
}

.project-list {
    list-style: none;
    margin: 0;
    padding: 6px;
    max-height: 38vh;
    overflow-y: auto;
}

.project-list li {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 10px;
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
}

.project-list li:hover {
    background: var(--bg-hover);
}

.project-list li.active {
    background: var(--bg-active);
}

.project-list .name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.open-badge {
    font-size: 10px;
    color: var(--accent);
    background: rgba(255, 120, 73, 0.15);
    padding: 1px 5px;
    border-radius: 8px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 600;
}

.delete {
    opacity: 0;
    padding: 0 6px;
    font-size: 16px;
    line-height: 1;
}

.project-list li:hover .delete,
.project-list li.active .delete {
    opacity: 1;
}

.delete:hover {
    color: var(--danger);
}

.empty {
    padding: 24px;
    color: var(--text-dim);
    font-size: 13px;
    text-align: center;
}

.empty p { margin: 0 0 12px; }

.active-section {
    border-top: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    flex: 1;
}

.active-section header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-bottom: 1px solid var(--border);
    background: var(--bg);
}

.active-section .title {
    flex: 1;
    font-size: 13px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.vars-btn {
    padding: 3px 8px;
    font-size: 11px;
    font-weight: 600;
    border-radius: 3px;
}

.vars-btn.active {
    background: var(--bg-active);
    color: var(--accent);
}

.saving {
    font-size: 10px;
    color: var(--text-dim);
    font-style: italic;
}
</style>
