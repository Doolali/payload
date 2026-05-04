<script setup lang="ts">
// Reusable key/value editor used for query params and headers. Always
// renders one extra blank row at the bottom — typing into it adds a real
// entry, matching the Postman pattern. When key-suggestions or per-key
// value suggestions are supplied, the Autocomplete component renders in
// place of a plain input so the user gets a popup that actually closes.
import {computed} from 'vue';
import {model} from '../../wailsjs/go/models';
import Autocomplete from './Autocomplete.vue';

const props = defineProps<{
    rows: model.KV[];
    keyPlaceholder?: string;
    valuePlaceholder?: string;
    keySuggestions?: string[];
    valuesFor?: (key: string) => string[];
}>();

const emit = defineEmits<{
    (e: 'change'): void;
}>();

const display = computed(() => [...props.rows, blankRow()]);

function blankRow(): model.KV {
    return {key: '', value: '', enabled: true};
}

function ensureRow(index: number) {
    if (index === props.rows.length) {
        props.rows.push({key: '', value: '', enabled: true});
    }
}

function setKey(index: number, value: string) {
    if (!value && index === props.rows.length) return;
    ensureRow(index);
    props.rows[index].key = value;
    pruneTrailingBlanks();
    emit('change');
}

function setValue(index: number, value: string) {
    if (!value && index === props.rows.length) return;
    ensureRow(index);
    props.rows[index].value = value;
    pruneTrailingBlanks();
    emit('change');
}

function toggle(index: number) {
    if (index === props.rows.length) return;
    props.rows[index].enabled = !props.rows[index].enabled;
    emit('change');
}

function remove(index: number) {
    if (index === props.rows.length) return;
    props.rows.splice(index, 1);
    emit('change');
}

function pruneTrailingBlanks() {
    while (props.rows.length > 0) {
        const last = props.rows[props.rows.length - 1];
        if (!last.key && !last.value) props.rows.pop();
        else break;
    }
}

function valueOptions(key: string): string[] {
    return props.valuesFor ? props.valuesFor(key) : [];
}
</script>

<template>
    <table class="kv">
        <thead>
            <tr>
                <th class="enabled"></th>
                <th>Key</th>
                <th>Value</th>
                <th class="actions"></th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(row, i) in display" :key="i" :class="{disabled: !row.enabled}">
                <td class="enabled">
                    <input
                        type="checkbox"
                        :checked="row.enabled"
                        :disabled="i === props.rows.length"
                        @change="toggle(i)"
                    />
                </td>
                <td>
                    <Autocomplete
                        v-if="keySuggestions && keySuggestions.length"
                        :model-value="row.key"
                        :options="keySuggestions"
                        :placeholder="keyPlaceholder ?? 'key'"
                        @update:model-value="(v: string) => setKey(i, v)"
                    />
                    <input
                        v-else
                        type="text"
                        :value="row.key"
                        :placeholder="keyPlaceholder ?? 'key'"
                        autocomplete="off"
                        spellcheck="false"
                        @input="(e) => setKey(i, (e.target as HTMLInputElement).value)"
                    />
                </td>
                <td>
                    <Autocomplete
                        v-if="valueOptions(row.key).length"
                        :model-value="row.value"
                        :options="valueOptions(row.key)"
                        :placeholder="valuePlaceholder ?? 'value'"
                        @update:model-value="(v: string) => setValue(i, v)"
                    />
                    <input
                        v-else
                        type="text"
                        :value="row.value"
                        :placeholder="valuePlaceholder ?? 'value'"
                        autocomplete="off"
                        spellcheck="false"
                        @input="(e) => setValue(i, (e.target as HTMLInputElement).value)"
                    />
                </td>
                <td class="actions">
                    <button
                        v-if="i !== props.rows.length"
                        class="ghost"
                        title="Remove"
                        @click="remove(i)"
                    >×</button>
                </td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
.kv {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
}

.kv th {
    text-align: left;
    padding: 6px 8px;
    font-weight: 500;
    color: var(--text-dim);
    text-transform: uppercase;
    font-size: 11px;
    letter-spacing: 0.5px;
    border-bottom: 1px solid var(--border);
}

.kv td {
    padding: 0;
    border-bottom: 1px solid var(--border);
    position: relative;
}

.kv .enabled {
    width: 32px;
    text-align: center;
    padding: 4px 0;
}

.kv .actions {
    width: 32px;
    text-align: center;
}

.kv input[type="text"] {
    border: none;
    background: transparent;
    border-radius: 0;
    padding: 8px 10px;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 12px;
    width: 100%;
    color: var(--text);
}

.kv input[type="text"]:focus {
    background: var(--bg);
    border: none;
    outline: 1px solid var(--accent);
    outline-offset: -1px;
}

.kv tr.disabled input[type="text"] {
    color: var(--text-dim);
    text-decoration: line-through;
}

.kv .ghost {
    padding: 0 6px;
    font-size: 16px;
    line-height: 1;
}
</style>
