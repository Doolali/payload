<script setup lang="ts">
// Edit a variable map (project.variables or collection.variables). Same
// blank-row UX as KVTable but stores into a Record<string, string> instead
// of an array of KV. Empty keys are pruned on commit.
import {computed, ref} from 'vue';

const props = defineProps<{
    title: string;
    subtitle?: string;
    vars: Record<string, string>;
}>();

const emit = defineEmits<{
    (e: 'change'): void;
    (e: 'close'): void;
}>();

interface Row {
    key: string;
    value: string;
    originalKey: string;
}

const sampleSyntax = '{' + '{name}' + '}';

const rows = ref<Row[]>(buildRows());

function buildRows(): Row[] {
    return Object.entries(props.vars ?? {}).map(([k, v]) => ({key: k, value: v, originalKey: k}));
}

const display = computed<Row[]>(() => [...rows.value, {key: '', value: '', originalKey: ''}]);

function commit() {
    const next: Record<string, string> = {};
    for (const r of rows.value) {
        const k = r.key.trim();
        if (!k) continue;
        next[k] = r.value;
    }
    // Mutate in place so Vue reactivity reaches the parent's project.
    for (const k of Object.keys(props.vars)) delete props.vars[k];
    Object.assign(props.vars, next);
    emit('change');
}

function onInput(i: number, field: 'key' | 'value', val: string) {
    if (i === rows.value.length) {
        if (!val) return;
        rows.value.push({key: '', value: '', originalKey: ''});
    }
    rows.value[i][field] = val;
    commit();
}

function remove(i: number) {
    if (i === rows.value.length) return;
    rows.value.splice(i, 1);
    commit();
}
</script>

<template>
    <section class="vars">
        <header>
            <div>
                <h2>{{ title }}</h2>
                <p v-if="subtitle" class="sub">{{ subtitle }}</p>
            </div>
            <button class="ghost" @click="emit('close')">Close</button>
        </header>

        <p class="hint">
            Reference these from any request as <code>{{ sampleSyntax }}</code> in the URL,
            headers, query params, or body.
        </p>

        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Value</th>
                    <th></th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="(row, i) in display" :key="i">
                    <td>
                        <input
                            type="text"
                            :value="row.key"
                            placeholder="name"
                            spellcheck="false"
                            @input="onInput(i, 'key', ($event.target as HTMLInputElement).value)"
                        />
                    </td>
                    <td>
                        <input
                            type="text"
                            :value="row.value"
                            placeholder="value"
                            spellcheck="false"
                            @input="onInput(i, 'value', ($event.target as HTMLInputElement).value)"
                        />
                    </td>
                    <td class="actions">
                        <button
                            v-if="i !== rows.length"
                            class="ghost"
                            title="Remove"
                            @click="remove(i)"
                        >×</button>
                    </td>
                </tr>
            </tbody>
        </table>
    </section>
</template>

<style scoped>
.vars {
    flex: 1;
    overflow: auto;
    padding: 16px 20px;
    display: flex;
    flex-direction: column;
}

header {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 8px;
}

header > div {
    flex: 1;
}

h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
}

.sub {
    margin: 4px 0 0;
    color: var(--text-dim);
    font-size: 12px;
}

.hint {
    margin: 0 0 12px;
    color: var(--text-dim);
    font-size: 12px;
}

.hint code {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 3px;
    padding: 1px 4px;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 11px;
    color: var(--accent);
}

table {
    border-collapse: collapse;
    width: 100%;
    max-width: 720px;
}

th, td {
    border-bottom: 1px solid var(--border);
    text-align: left;
    padding: 0;
}

th {
    padding: 6px 8px;
    color: var(--text-dim);
    font-weight: 500;
    text-transform: uppercase;
    font-size: 11px;
    letter-spacing: 0.5px;
}

td input[type="text"] {
    border: none;
    background: transparent;
    border-radius: 0;
    padding: 8px 10px;
    width: 100%;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 12px;
}

td input[type="text"]:focus {
    background: var(--bg);
    outline: 1px solid var(--accent);
    outline-offset: -1px;
}

.actions {
    width: 40px;
    text-align: center;
}

.actions button {
    padding: 0 6px;
    font-size: 16px;
    line-height: 1;
}
</style>
