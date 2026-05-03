<script setup lang="ts">
// Display the result of the last Send. Switches between Body and Headers.
// JSON bodies are pretty-printed; everything else is shown raw.
import {computed, ref} from 'vue';
import {model} from '../../wailsjs/go/models';

const props = defineProps<{
    response: model.Response | null;
    busy: boolean;
}>();

const tab = ref<'body' | 'headers'>('body');

const statusClass = computed(() => {
    const s = props.response?.status ?? 0;
    if (s >= 500) return 'status err';
    if (s >= 400) return 'status warn';
    if (s >= 300) return 'status info';
    if (s >= 200) return 'status ok';
    return 'status unknown';
});

const formattedBody = computed(() => {
    if (!props.response) return '';
    const ct = (props.response.headers ?? [])
        .find(h => h.key.toLowerCase() === 'content-type')?.value
        ?? '';
    const body = props.response.body ?? '';
    if (ct.includes('json') && body.trim().length > 0) {
        try {
            return JSON.stringify(JSON.parse(body), null, 2);
        } catch {
            return body;
        }
    }
    return body;
});
</script>

<template>
    <section class="response">
        <header v-if="response">
            <span :class="statusClass">{{ response.status || '—' }} {{ response.statusText }}</span>
            <span class="meta">{{ response.durationMs }} ms</span>
            <span v-if="response.error" class="error-msg">{{ response.error }}</span>
            <div class="tabs">
                <button :class="{active: tab === 'body'}" @click="tab = 'body'">Body</button>
                <button :class="{active: tab === 'headers'}" @click="tab = 'headers'">Headers ({{ response.headers?.length ?? 0 }})</button>
            </div>
        </header>
        <div v-if="busy" class="placeholder">Sending…</div>
        <div v-else-if="!response" class="placeholder">No response yet — click Send.</div>
        <div v-else class="content">
            <pre v-if="tab === 'body'" class="body">{{ formattedBody }}</pre>
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
