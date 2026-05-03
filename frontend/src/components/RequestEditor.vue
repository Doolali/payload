<script setup lang="ts">
// The editor + response pair for a single Request (session or collection
// request). Mutating any field schedules an autosave; Send hits the backend
// HTTP client with merged variables and emits the response back to the
// parent so it can decide whether to persist it.
import {ref} from 'vue';
import {SendRequest} from '../../wailsjs/go/main/App';
import {model} from '../../wailsjs/go/models';
import KVTable from './KVTable.vue';
import BodyEditor from './BodyEditor.vue';
import ResponsePane from './ResponsePane.vue';

const props = defineProps<{
    request: model.Request | model.Session;
    vars: Record<string, string>;
    response: model.Response | null;
    persistResponse: boolean;
    saveTargets?: model.Collection[];
}>();

const emit = defineEmits<{
    (e: 'change'): void;
    (e: 'response', resp: model.Response): void;
    (e: 'saveToCollection', collectionId: string): void;
}>();

function onSaveTargetChange(ev: Event) {
    const select = ev.target as HTMLSelectElement;
    const id = select.value;
    if (id) {
        emit('saveToCollection', id);
        select.value = '';
    }
}

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];
const tab = ref<'params' | 'headers' | 'body'>('params');
const sending = ref(false);

function methodClass(m: string) {
    return `method method-${m.toLowerCase()}`;
}

function onChange() {
    (props.request as any).updatedAt = new Date().toISOString();
    emit('change');
}

async function send() {
    sending.value = true;
    try {
        const req: model.Request = {
            id: props.request.id,
            name: props.request.name,
            method: props.request.method,
            url: props.request.url,
            headers: props.request.headers,
            queryParams: props.request.queryParams,
            body: props.request.body,
            createdAt: props.request.createdAt,
            updatedAt: props.request.updatedAt,
        } as any;
        const resp = await SendRequest(req, props.vars);
        emit('response', resp);
    } finally {
        sending.value = false;
    }
}
</script>

<template>
    <section class="editor">
        <div class="name-row">
            <input
                v-model="(props.request as any).name"
                type="text"
                class="name-input"
                placeholder="Untitled request"
                @input="onChange"
            />
            <select
                v-if="saveTargets"
                class="save-target"
                :disabled="!saveTargets.length"
                :title="saveTargets.length ? 'Save this session to a collection' : 'Create a collection first'"
                @change="onSaveTargetChange"
            >
                <option value="">Save to…</option>
                <option v-for="c in saveTargets" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
        </div>

        <div class="url-row">
            <select
                v-model="(props.request as any).method"
                :class="methodClass(String(props.request.method))"
                @change="onChange"
            >
                <option v-for="m in METHODS" :key="m" :value="m">{{ m }}</option>
            </select>
            <input
                v-model="(props.request as any).url"
                type="text"
                class="url-input"
                placeholder="https://api.example.com/path"
                spellcheck="false"
                @input="onChange"
                @keydown.enter="send"
            />
            <button
                class="primary send"
                :disabled="!props.request.url || sending"
                @click="send"
            >{{ sending ? 'Sending…' : 'Send' }}</button>
        </div>

        <div class="tabs">
            <button :class="{active: tab === 'params'}" @click="tab = 'params'">
                Params <span v-if="props.request.queryParams.length" class="count">{{ props.request.queryParams.length }}</span>
            </button>
            <button :class="{active: tab === 'headers'}" @click="tab = 'headers'">
                Headers <span v-if="props.request.headers.length" class="count">{{ props.request.headers.length }}</span>
            </button>
            <button :class="{active: tab === 'body'}" @click="tab = 'body'">Body</button>
        </div>

        <div class="tab-content">
            <KVTable
                v-if="tab === 'params'"
                :rows="props.request.queryParams"
                key-placeholder="param"
                @change="onChange"
            />
            <KVTable
                v-else-if="tab === 'headers'"
                :rows="props.request.headers"
                key-placeholder="header"
                @change="onChange"
            />
            <BodyEditor
                v-else
                :body="props.request.body"
                @change="onChange"
            />
        </div>

        <ResponsePane :response="response" :busy="sending" />
    </section>
</template>

<style scoped>
.editor {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
}

.name-row {
    padding: 12px 16px 0;
    display: flex;
    align-items: center;
    gap: 8px;
}

.name-input {
    border: none;
    background: transparent;
    color: var(--text);
    font-size: 16px;
    font-weight: 600;
    padding: 4px 6px;
    flex: 1;
    min-width: 0;
}

.save-target {
    background: var(--bg);
    color: var(--text-dim);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 4px 8px;
    font-size: 12px;
    font-family: inherit;
    cursor: pointer;
}

.save-target:hover:not(:disabled) {
    color: var(--text);
}

.save-target:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.name-input:hover, .name-input:focus {
    background: var(--bg);
    border-radius: 4px;
    outline: none;
}

.url-row {
    display: flex;
    gap: 6px;
    padding: 8px 16px;
}

.url-row select {
    border: 1px solid var(--border);
    background: var(--bg);
    color: var(--text);
    border-radius: 4px;
    padding: 6px 10px;
    font-weight: 600;
    font-family: inherit;
    cursor: pointer;
    min-width: 100px;
}

.method-get { color: #4cd964; }
.method-post { color: #ff9500; }
.method-put { color: #5ac8fa; }
.method-patch { color: #af52de; }
.method-delete { color: #ff3b30; }
.method-head, .method-options { color: var(--text-dim); }

.url-input {
    flex: 1;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 13px;
}

.send {
    min-width: 80px;
}

.send:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.tabs {
    display: flex;
    gap: 4px;
    padding: 0 16px;
    border-bottom: 1px solid var(--border);
}

.tabs button {
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-dim);
    padding: 8px 12px;
    cursor: pointer;
    font-size: 13px;
    border-radius: 0;
}

.tabs button:hover { color: var(--text); }

.tabs button.active {
    color: var(--text);
    border-bottom-color: var(--accent);
}

.count {
    background: var(--bg-active);
    color: var(--text);
    border-radius: 8px;
    padding: 1px 6px;
    font-size: 10px;
    margin-left: 4px;
}

.tab-content {
    padding: 8px 16px;
    overflow: auto;
    min-height: 160px;
    max-height: 40vh;
}
</style>
