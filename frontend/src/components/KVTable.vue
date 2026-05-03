<script setup lang="ts">
// Reusable key/value editor. Used for query params and headers.
//
// Always renders one extra blank row at the bottom — typing into it adds a
// real entry. This is the same UX pattern Postman uses.
import {computed} from 'vue';
import {model} from '../../wailsjs/go/models';

const props = defineProps<{
    rows: model.KV[];
    keyPlaceholder?: string;
    valuePlaceholder?: string;
}>();

const emit = defineEmits<{
    (e: 'change'): void;
}>();

const display = computed(() => [...props.rows, blankRow()]);

function blankRow(): model.KV {
    return {key: '', value: '', enabled: true};
}

function onInput(index: number, field: 'key' | 'value', value: string) {
    if (index === props.rows.length) {
        if (!value) return;
        props.rows.push({key: '', value: '', enabled: true});
    }
    props.rows[index][field] = value;
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
                    <input
                        type="text"
                        :value="row.key"
                        :placeholder="keyPlaceholder ?? 'key'"
                        @input="(e) => onInput(i, 'key', (e.target as HTMLInputElement).value)"
                    />
                </td>
                <td>
                    <input
                        type="text"
                        :value="row.value"
                        :placeholder="valuePlaceholder ?? 'value'"
                        @input="(e) => onInput(i, 'value', (e.target as HTMLInputElement).value)"
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
