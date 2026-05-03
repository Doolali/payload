<script setup lang="ts">
// Sessions tab content for the left sidebar. Pure global session list — no
// dependency on whether a project is open. Sessions persist in their own
// file and survive project switches and app restarts.
import {model} from '../../wailsjs/go/models';
import {useProject} from '../composables/useProject';
import SessionList from './SessionList.vue';

const {state, scheduleSessionsSave} = useProject();

function newSession() {
    const now = new Date().toISOString() as any;
    const s = {
        id: crypto.randomUUID(),
        name: 'New request',
        method: 'GET',
        url: '',
        headers: [],
        queryParams: [],
        body: {type: 'none', content: ''},
        createdAt: now,
        updatedAt: now,
        lastResponse: undefined,
    } as unknown as model.Session;
    state.sessions.push(s);
    state.selection = {kind: 'session', sessionId: s.id};
    state.transientResponse = null;
    scheduleSessionsSave();
}

function selectSession(id: string) {
    state.selection = {kind: 'session', sessionId: id};
    state.transientResponse = null;
}

function deleteSession(id: string) {
    state.sessions = state.sessions.filter(s => s.id !== id);
    if (state.selection?.kind === 'session' && state.selection.sessionId === id) {
        state.selection = state.sessions[0]
            ? {kind: 'session', sessionId: state.sessions[0].id}
            : null;
    }
    scheduleSessionsSave();
}

function clearSessions() {
    state.sessions = [];
    if (state.selection?.kind === 'session') state.selection = null;
    scheduleSessionsSave();
}
</script>

<template>
    <div class="sessions-tab">
        <SessionList
            :sessions="state.sessions"
            :selected-id="state.selection?.kind === 'session' ? state.selection.sessionId : null"
            @select="selectSession"
            @new="newSession"
            @delete="deleteSession"
            @clear="clearSessions"
        />
    </div>
</template>

<style scoped>
.sessions-tab {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    flex: 1;
}
</style>
