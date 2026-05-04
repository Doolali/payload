<script setup lang="ts">
// Wrapper that parses the raw response text and hands the value tree to
// JsonNode. If parsing fails, falls back to a plain <pre>. Forwards the
// optional savePath callback into the tree via Vue's provide so primitives
// can offer a "save as variable" affordance.
import {computed, provide} from 'vue';
import JsonNode from './JsonNode.vue';

const props = defineProps<{
    raw: string;
    onSavePath?: (path: string, value: any) => void;
}>();

provide('savePath', props.onSavePath ?? null);

const parsed = computed<{ok: true; value: any} | {ok: false}>(() => {
    if (!props.raw || !props.raw.trim()) return {ok: true, value: null};
    try {
        return {ok: true, value: JSON.parse(props.raw)};
    } catch {
        return {ok: false};
    }
});
</script>

<template>
    <pre v-if="!parsed.ok" class="raw">{{ raw }}</pre>
    <div v-else class="json-view">
        <JsonNode :value="parsed.value" :depth="0" :is-last="true" :path="''" />
    </div>
</template>

<style scoped>
.json-view {
    padding: 12px 16px;
    overflow: auto;
}

.raw {
    margin: 0;
    padding: 12px 16px;
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    color: var(--text);
    white-space: pre-wrap;
    word-break: break-word;
}
</style>
