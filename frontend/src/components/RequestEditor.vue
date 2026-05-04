<script setup lang="ts">
// The editor + response pair for a single Request (session or collection
// request). Mutating any field schedules an autosave; Send hits the backend
// HTTP client with merged variables and emits the response back to the
// parent so it can decide whether to persist it.
import {computed, ref} from 'vue';
import {SendRequest} from '../../wailsjs/go/main/App';
import {model} from '../../wailsjs/go/models';
import {useProject} from '../composables/useProject';
import {promptText} from '../composables/useDialogs';
import {defaultVarName, resolvePath, toVarString} from '../composables/jsonPath';
import {HEADER_NAMES, valuesForHeader} from '../composables/headerSuggestions';
import KVTable from './KVTable.vue';
import BodyEditor from './BodyEditor.vue';
import ResponsePane from './ResponsePane.vue';

const {state, scheduleProjectSave} = useProject();

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

const extractors = computed<model.Extractor[]>({
    get: () => ((props.request as any).extractors ?? []) as model.Extractor[],
    set: (v) => { (props.request as any).extractors = v; },
});

// Apply the request's saved extractors to the parsed response body. Each
// extractor that resolves to a non-undefined value writes to project.vars.
function applyExtractors(respBody: string): number {
    if (!state.project) return 0;
    if (!extractors.value.length) return 0;
    let parsed: any;
    try { parsed = JSON.parse(respBody); } catch { return 0; }
    let written = 0;
    for (const ex of extractors.value) {
        const v = resolvePath(parsed, ex.path);
        if (v === undefined) continue;
        state.project.variables[ex.varName] = toVarString(v);
        written++;
    }
    if (written) scheduleProjectSave();
    return written;
}

// Capture a value picked from the JSON tree. Prompts for a name, registers
// the extractor on the request, writes the current value to project vars
// immediately so subsequent {{name}} references work straight away.
async function captureExtractor(path: string, value: any) {
    if (!state.project) {
        state.lastError = 'Open a project to save variables.';
        return;
    }
    const initial = defaultVarName(path);
    const name = await promptText({
        title: 'Save as project variable',
        label: `Path: ${path || '$'}`,
        placeholder: 'variable name',
        initial,
        confirmText: 'Save',
    });
    if (!name?.trim()) return;
    const trimmed = name.trim();
    state.project.variables[trimmed] = toVarString(value);
    if (!(props.request as any).extractors) (props.request as any).extractors = [];
    const list = extractors.value;
    const existing = list.findIndex(e => e.varName === trimmed);
    const ex = {path, varName: trimmed} as model.Extractor;
    if (existing >= 0) list[existing] = ex;
    else list.push(ex);
    onChange();
    scheduleProjectSave();
}

function removeExtractor(varName: string) {
    extractors.value = extractors.value.filter(e => e.varName !== varName);
    onChange();
}

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
        if (resp.body) applyExtractors(resp.body);
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
                :key-suggestions="HEADER_NAMES"
                :values-for="valuesForHeader"
                @change="onChange"
            />
            <BodyEditor
                v-else
                :body="props.request.body"
                @change="onChange"
            />
        </div>

        <ResponsePane :response="response" :busy="sending" :on-save-path="captureExtractor" />

        <div v-if="extractors.length" class="extractors">
            <span class="ex-label">Saved on send:</span>
            <span v-for="ex in extractors" :key="ex.varName" class="ex-chip">
                <code>{{ ex.varName }}</code>
                <span class="ex-path">← {{ ex.path || '$' }}</span>
                <button class="ex-del" title="Remove" @click="removeExtractor(ex.varName)">×</button>
            </span>
        </div>
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

.extractors {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    padding: 6px 16px;
    border-top: 1px solid var(--border);
    background: var(--bg-elev);
    font-size: 11px;
}

.ex-label {
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.ex-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 2px 4px 2px 8px;
}

.ex-chip code {
    color: var(--accent);
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
}

.ex-path {
    color: var(--text-dim);
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 10px;
}

.ex-del {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    font-size: 14px;
    line-height: 1;
    padding: 0 4px;
    border-radius: 50%;
}

.ex-del:hover {
    color: var(--danger);
    background: var(--bg-hover);
}
</style>
