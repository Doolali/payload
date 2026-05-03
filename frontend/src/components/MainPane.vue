<script setup lang="ts">
// Right-hand pane. Renders the right component for the current selection:
// a request editor (sessions or collection requests), a variables editor
// (project or collection), or an empty hint.
import {computed} from 'vue';
import {model} from '../../wailsjs/go/models';
import {useProject} from '../composables/useProject';
import RequestEditor from './RequestEditor.vue';
import VariablesEditor from './VariablesEditor.vue';
import logo from '../assets/images/logo.png';

const {state, selectedSession, selectedRequest, selectedCollection, mergedVars, scheduleSave, setResponse, getResponse} = useProject();

const activeRequest = computed(() => selectedSession.value || selectedRequest.value);
const persistResponse = computed(() => state.selection?.kind === 'session');
const saveTargets = computed<model.Collection[] | undefined>(() => {
    if (state.selection?.kind !== 'session' || !state.project) return undefined;
    return state.project.collections;
});

function onResponse(resp: any) {
    setResponse(resp);
}

// Clone the currently selected session into the chosen collection as a
// fresh Request (new id, no lastResponse). The session stays put — this
// "promotes" rather than moves, matching how Postman handles save-to-collection.
function saveSessionToCollection(collectionId: string) {
    if (!state.project || !selectedSession.value) return;
    const sess = selectedSession.value;
    const c = state.project.collections.find(c => c.id === collectionId);
    if (!c) return;
    const now = new Date().toISOString() as any;
    const cloned: model.Request = {
        id: crypto.randomUUID(),
        name: sess.name || 'Untitled',
        method: sess.method,
        url: sess.url,
        headers: cloneArr(sess.headers),
        queryParams: cloneArr(sess.queryParams),
        body: {type: sess.body.type, content: sess.body.content} as model.Body,
        createdAt: now,
        updatedAt: now,
    } as any;
    c.requests.push(cloned);
    state.selection = {kind: 'request', collectionId: c.id, requestId: cloned.id};
    state.transientResponse = null;
    scheduleSave();
}

function cloneArr(arr: model.KV[]): model.KV[] {
    return arr.map(kv => ({key: kv.key, value: kv.value, enabled: kv.enabled}));
}
</script>

<template>
    <main class="main-pane">
        <RequestEditor
            v-if="activeRequest"
            :key="(activeRequest as any).id"
            :request="(activeRequest as any)"
            :vars="mergedVars"
            :response="getResponse()"
            :persist-response="persistResponse"
            :save-targets="saveTargets"
            @change="scheduleSave"
            @response="onResponse"
            @save-to-collection="saveSessionToCollection"
        />
        <VariablesEditor
            v-else-if="state.selection?.kind === 'project-vars' && state.project"
            title="Project variables"
            :subtitle="state.project.name"
            :vars="state.project.variables"
            @change="scheduleSave"
            @close="state.selection = null"
        />
        <VariablesEditor
            v-else-if="state.selection?.kind === 'collection-vars' && selectedCollection"
            title="Collection variables"
            :subtitle="selectedCollection.name"
            :vars="selectedCollection.variables"
            @change="scheduleSave"
            @close="state.selection = null"
        />
        <div v-else-if="state.project" class="empty">
            <p>Pick a session or request from the left, or create one.</p>
        </div>
        <div v-else class="empty">
            <img :src="logo" alt="payload" class="hero" />
            <p>Open a project from the Projects tab to begin.</p>
        </div>
    </main>
</template>

<style scoped>
.main-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
}

.empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    color: var(--text-dim);
    font-size: 14px;
    text-align: center;
    padding: 40px;
}

.hero {
    width: 200px;
    height: 200px;
    image-rendering: pixelated;
    border-radius: 12px;
    opacity: 0.85;
}
</style>
