<script setup lang="ts">
// Session list inside the project sub-sidebar. Sessions live as long as the
// project itself; "Clear all" wipes the list, individual delete removes one.
// New sessions default to GET / no URL and the user takes it from there.
import {model} from '../../wailsjs/go/models';

const props = defineProps<{
    sessions: model.Session[];
    selectedId: string | null;
}>();

const emit = defineEmits<{
    (e: 'select', id: string): void;
    (e: 'new'): void;
    (e: 'delete', id: string): void;
    (e: 'clear'): void;
}>();

function preview(s: model.Session): string {
    if (s.name && s.name !== 'New request') return s.name;
    return s.url || 'New request';
}

function confirmClear() {
    if (!props.sessions.length) return;
    if (window.confirm('Clear all sessions?')) emit('clear');
}
</script>

<template>
    <div class="session-list">
        <div class="section-header">
            <span>Sessions</span>
            <div class="actions">
                <button class="primary tiny" title="New session" @click="emit('new')">+ New</button>
                <button
                    class="ghost"
                    title="Clear all sessions"
                    :disabled="!sessions.length"
                    @click="confirmClear"
                >⌫</button>
            </div>
        </div>
        <ul v-if="sessions.length">
            <li
                v-for="s in sessions"
                :key="s.id"
                :class="{selected: s.id === selectedId}"
                @click="emit('select', s.id)"
            >
                <span :class="`badge method-${String(s.method).toLowerCase()}`">{{ s.method }}</span>
                <span class="preview">{{ preview(s) }}</span>
                <button
                    class="ghost del"
                    title="Delete session"
                    @click.stop="emit('delete', s.id)"
                >×</button>
            </li>
        </ul>
        <div v-else class="empty">
            <p>No sessions yet.</p>
            <button class="primary" @click="emit('new')">+ New request</button>
        </div>
    </div>
</template>

<style scoped>
.session-list {
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.section-header {
    display: flex;
    align-items: center;
    padding: 10px 12px 6px;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: var(--text-dim);
}

.section-header span {
    flex: 1;
}

.actions {
    display: flex;
    gap: 2px;
}

.actions button {
    line-height: 1;
}

.actions .ghost {
    padding: 0 6px;
    font-size: 14px;
}

.tiny {
    padding: 3px 8px;
    font-size: 11px;
    text-transform: none;
    letter-spacing: 0;
}

.actions button:disabled {
    opacity: 0.3;
    cursor: not-allowed;
}

ul {
    list-style: none;
    margin: 0;
    padding: 0 6px;
    overflow-y: auto;
    flex: 1;
}

li {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
    font-size: 13px;
}

li:hover {
    background: var(--bg-hover);
}

li.selected {
    background: var(--bg-active);
}

.badge {
    font-size: 10px;
    font-weight: 700;
    padding: 1px 4px;
    border-radius: 2px;
    background: var(--bg);
    min-width: 38px;
    text-align: center;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
}

.method-get { color: #4cd964; }
.method-post { color: #ff9500; }
.method-put { color: #5ac8fa; }
.method-patch { color: #af52de; }
.method-delete { color: #ff3b30; }
.method-head, .method-options { color: var(--text-dim); }

.preview {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-dim);
}

li.selected .preview {
    color: var(--text);
}

.del {
    opacity: 0;
    padding: 0 6px;
    font-size: 14px;
    line-height: 1;
}

li:hover .del, li.selected .del {
    opacity: 1;
}

.del:hover {
    color: var(--danger);
}

.empty {
    padding: 16px 12px;
    color: var(--text-dim);
    font-size: 12px;
    text-align: center;
}

.empty p {
    margin: 0 0 10px;
}
</style>
