<script setup lang="ts">
// Display the result of the last Send. Switches between Body and Headers.
// JSON bodies render in a collapsible, syntax-highlighted tree; other
// content types fall back to a plain <pre>.
import {computed, ref} from 'vue';
import {model} from '../../wailsjs/go/models';
import JsonView from './JsonView.vue';

const props = defineProps<{
    response: model.Response | null;
    busy: boolean;
    onSavePath?: (path: string, value: any) => void;
}>();

const tab = ref<'body' | 'headers' | 'raw'>('body');

const statusClass = computed(() => {
    const s = props.response?.status ?? 0;
    if (s >= 500) return 'status err';
    if (s >= 400) return 'status warn';
    if (s >= 300) return 'status info';
    if (s >= 200) return 'status ok';
    return 'status unknown';
});

// True when the response declares a JSON content type. Drives whether we
// hand the body to JsonView or the plain raw renderer.
const isJson = computed(() => {
    if (!props.response) return false;
    const ct = (props.response.headers ?? [])
        .find(h => h.key.toLowerCase() === 'content-type')?.value
        ?? '';
    return ct.includes('json');
});

const copyState = ref<'idle' | 'copied' | 'failed'>('idle');

// Copy the response body to the clipboard. JSON gets pretty-printed first
// so the pasted text matches what the tree viewer shows visually.
async function copyBody() {
    if (!props.response) return;
    const body = props.response.body ?? '';
    let text = body;
    if (isJson.value) {
        try {
            text = JSON.stringify(JSON.parse(body), null, 2);
        } catch {
            text = body;
        }
    }
    try {
        await navigator.clipboard.writeText(text);
        copyState.value = 'copied';
    } catch {
        copyState.value = 'failed';
    }
    setTimeout(() => { copyState.value = 'idle'; }, 1500);
}
</script>

<template>
    <section class="response">
        <header v-if="response">
            <span :class="statusClass">{{ response.status || '—' }} {{ response.statusText }}</span>
            <span class="meta">{{ response.durationMs }} ms</span>
            <span v-if="response.error" class="error-msg">{{ response.error }}</span>
            <div class="tabs">
                <button :class="{active: tab === 'body'}" @click="tab = 'body'">Body</button>
                <button v-if="isJson" :class="{active: tab === 'raw'}" @click="tab = 'raw'">Raw</button>
                <button :class="{active: tab === 'headers'}" @click="tab = 'headers'">Headers ({{ response.headers?.length ?? 0 }})</button>
                <button
                    class="copy"
                    :title="isJson ? 'Copy formatted JSON' : 'Copy body'"
                    @click="copyBody"
                >{{ copyState === 'copied' ? 'Copied!' : copyState === 'failed' ? 'Failed' : 'Copy' }}</button>
            </div>
        </header>
        <div v-if="busy" class="placeholder">Sending…</div>
        <div v-else-if="!response" class="placeholder">No response yet — click Send.</div>
        <div v-else class="content">
            <JsonView v-if="tab === 'body' && isJson" :raw="response.body" :on-save-path="onSavePath" />
            <pre v-else-if="tab === 'body' || tab === 'raw'" class="body">{{ response.body }}</pre>
            <table v-else class="headers">
                <thead><tr><th>Key</th><th>Value</th></tr></thead>
                <tbody>
                    <tr v-for="(h, i) in response.headers" :key="i">
                        <td>{{ h.key }}</td>
                        <td>{{ h.value }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
</template>

<style scoped>
.response {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
    border-top: 1px solid var(--border);
}

header {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 12px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
}

.status {
    font-weight: 600;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 3px;
}

.status.ok { background: #1f6f4a; color: #d3f9e1; }
.status.info { background: #1f4f6f; color: #d3eaf9; }
.status.warn { background: #6f5a1f; color: #f9eed3; }
.status.err { background: #6f1f2f; color: #f9d3d8; }
.status.unknown { background: var(--bg); color: var(--text-dim); }

.meta {
    font-size: 12px;
    color: var(--text-dim);
}

.error-msg {
    color: var(--danger);
    font-size: 12px;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.tabs {
    display: flex;
    gap: 2px;
    margin-left: auto;
}

.tabs button {
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 12px;
    padding: 4px 10px;
    border-radius: 3px;
    cursor: pointer;
}

.tabs button:hover { color: var(--text); }
.tabs button.active { color: var(--text); background: var(--bg); }

.tabs button.copy {
    margin-left: 8px;
    border: 1px solid var(--border);
    background: var(--bg);
    min-width: 60px;
}

.tabs button.copy:hover { background: var(--bg-hover); }

.content {
    flex: 1;
    overflow: auto;
}

.placeholder {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-dim);
    font-size: 13px;
}

pre.body {
    margin: 0;
    padding: 12px 16px;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 12px;
    color: var(--text);
    white-space: pre-wrap;
    word-break: break-word;
}

table.headers {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
}

table.headers th, table.headers td {
    text-align: left;
    padding: 6px 12px;
    border-bottom: 1px solid var(--border);
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
}

table.headers th {
    color: var(--text-dim);
    font-weight: 500;
    text-transform: uppercase;
    font-size: 10px;
    letter-spacing: 0.5px;
}
</style>
