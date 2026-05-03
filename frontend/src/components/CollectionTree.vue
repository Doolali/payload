<script setup lang="ts">
// Collections section of the project sub-sidebar. Each collection is a
// foldable node with its requests inside. Per-collection actions: add
// request, edit variables, delete. Top-level: add new collection.
import {ref} from 'vue';
import {model} from '../../wailsjs/go/models';
import type {Selection} from '../composables/useProject';
import {confirmAction, promptText} from '../composables/useDialogs';

const props = defineProps<{
    collections: model.Collection[];
    selection: Selection;
}>();

// Array-form emits — wider Vue 3.2 compatibility than the typed-call form,
// which silently dropped some events in earlier 3.2.x releases.
const emit = defineEmits([
    'newCollection',
    'newRequest',
    'selectRequest',
    'selectVars',
    'renameCollection',
    'deleteCollection',
    'deleteRequest',
]);

const expanded = ref<Record<string, boolean>>({});

function toggle(id: string) {
    expanded.value[id] = !(expanded.value[id] ?? true);
}

function isExpanded(id: string): boolean {
    return expanded.value[id] ?? true;
}

function isSelectedRequest(collectionId: string, requestId: string): boolean {
    const s = props.selection;
    return s?.kind === 'request' && s.collectionId === collectionId && s.requestId === requestId;
}

function isSelectedVars(collectionId: string): boolean {
    const s = props.selection;
    return s?.kind === 'collection-vars' && s.collectionId === collectionId;
}

async function startRename(c: model.Collection, ev: MouseEvent) {
    ev.stopPropagation();
    const name = await promptText({
        title: 'Rename collection',
        placeholder: 'Collection name',
        initial: c.name,
        confirmText: 'Rename',
    });
    if (name && name.trim()) emit('renameCollection', c.id, name.trim());
}

async function confirmDeleteCollection(c: model.Collection, ev: MouseEvent) {
    ev.stopPropagation();
    const ok = await confirmAction({
        title: 'Delete collection',
        message: `Delete "${c.name}" and its ${c.requests?.length ?? 0} request(s)?`,
        confirmText: 'Delete',
        danger: true,
    });
    if (ok) emit('deleteCollection', c.id);
}
</script>

<template>
    <div class="tree">
        <div class="section-header">
            <span>Collections</span>
            <div class="actions">
                <button class="primary tiny" title="New collection" @click="emit('newCollection')">+ New</button>
            </div>
        </div>

        <ul v-if="collections.length" class="collections">
            <li v-for="c in collections" :key="c.id" class="collection">
                <div class="row coll-row" @click="toggle(c.id)">
                    <span class="caret">{{ isExpanded(c.id) ? '▾' : '▸' }}</span>
                    <span class="name" :title="c.name">{{ c.name }}</span>
                    <button class="ghost mini" title="New request" @click.stop="emit('newRequest', c.id)">+</button>
                    <button
                        class="ghost mini"
                        :class="{active: isSelectedVars(c.id)}"
                        title="Edit variables"
                        @click.stop="emit('selectVars', c.id)"
                    >{}</button>
                    <button class="ghost mini" title="Rename" @click.stop="startRename(c, $event)">✎</button>
                    <button class="ghost mini del" title="Delete" @click.stop="confirmDeleteCollection(c, $event)">×</button>
                </div>
                <ul v-if="isExpanded(c.id) && c.requests?.length" class="requests">
                    <li
                        v-for="r in c.requests"
                        :key="r.id"
                        class="row req-row"
                        :class="{selected: isSelectedRequest(c.id, r.id)}"
                        @click="emit('selectRequest', c.id, r.id)"
                    >
                        <span :class="`badge method-${String(r.method).toLowerCase()}`">{{ r.method }}</span>
                        <span class="name">{{ r.name || r.url || 'Untitled' }}</span>
                        <button
                            class="ghost mini del"
                            title="Delete request"
                            @click.stop="emit('deleteRequest', c.id, r.id)"
                        >×</button>
                    </li>
                </ul>
                <div v-else-if="isExpanded(c.id)" class="empty-coll">No requests yet.</div>
            </li>
        </ul>
        <div v-else class="empty">
            <p>No collections yet.</p>
            <button class="primary" @click="emit('newCollection')">+ New collection</button>
        </div>
    </div>
</template>

<style scoped>
.tree {
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.section-header {
    display: flex;
    align-items: center;
    padding: 14px 12px 6px;
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
    border-radius: 3px;
}

.collections, .requests {
    list-style: none;
    margin: 0;
    padding: 0 6px;
}

.requests {
    padding: 0 6px 0 22px;
}

.row {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 5px 8px;
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
    font-size: 13px;
}

.row:hover {
    background: var(--bg-hover);
}

.req-row.selected {
    background: var(--bg-active);
}

.coll-row {
    color: var(--text);
    font-weight: 500;
}

.caret {
    font-size: 10px;
    width: 10px;
    color: var(--text-dim);
}

.name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.req-row .name {
    color: var(--text-dim);
}

.req-row.selected .name {
    color: var(--text);
}

.badge {
    font-size: 9px;
    font-weight: 700;
    padding: 1px 4px;
    border-radius: 2px;
    background: var(--bg);
    min-width: 36px;
    text-align: center;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
}

.method-get { color: #4cd964; }
.method-post { color: #ff9500; }
.method-put { color: #5ac8fa; }
.method-patch { color: #af52de; }
.method-delete { color: #ff3b30; }
.method-head, .method-options { color: var(--text-dim); }

.mini {
    opacity: 0;
    padding: 0 4px;
    font-size: 12px;
    line-height: 1;
}

.row:hover .mini, .mini.active {
    opacity: 1;
}

.mini.active {
    color: var(--accent);
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

.empty p { margin: 0 0 10px; }

.empty-coll {
    padding: 6px 12px 6px 28px;
    color: var(--text-dim);
    font-size: 12px;
    font-style: italic;
}
</style>
