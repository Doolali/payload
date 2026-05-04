<script setup lang="ts">
// Recursive tree node for the JSON viewer. Each node renders one of:
//   primitive    — `key: value` on one line, syntax-highlighted
//   container    — `key: ▾ {` ... children ... `}` (collapsible)
//   empty        — `key: {}` / `[]` on one line
// The component imports itself for recursion (Vue 3 SFC compiler tolerates
// the circular import). Each node also tracks its full JSON path from the
// root so primitives can offer a "save as variable" affordance.
import {computed, inject, ref} from 'vue';
import JsonNode from './JsonNode.vue';

const props = defineProps<{
    value: any;
    keyName?: string;
    isLast?: boolean;
    depth: number;
    path?: string;
}>();

const onSavePath = inject<((path: string, value: any) => void) | null>('savePath', null);

// Auto-expand the first few levels; deeper nodes start collapsed so big
// payloads aren't a wall of text on first load.
const expanded = ref(props.depth < 2);

const type = computed(() => {
    if (props.value === null) return 'null';
    if (Array.isArray(props.value)) return 'array';
    if (typeof props.value === 'object') return 'object';
    return typeof props.value;
});

const isContainer = computed(() => type.value === 'array' || type.value === 'object');
const isEmpty = computed(() => isContainer.value && entries.value.length === 0);

const entries = computed<Array<{k?: string; v: any}>>(() => {
    if (type.value === 'array') return (props.value as any[]).map(v => ({v}));
    if (type.value === 'object') return Object.entries(props.value).map(([k, v]) => ({k, v}));
    return [];
});

const open = computed(() => (type.value === 'array' ? '[' : '{'));
const close = computed(() => (type.value === 'array' ? ']' : '}'));
const comma = computed(() => (props.isLast ? '' : ','));

const formatted = computed(() => {
    if (type.value === 'string') return JSON.stringify(props.value);
    return String(props.value);
});

const summary = computed(() => {
    const n = entries.value.length;
    if (type.value === 'array') return `${n} item${n === 1 ? '' : 's'}`;
    return `${n} key${n === 1 ? '' : 's'}`;
});

function toggle() {
    if (isContainer.value && !isEmpty.value) expanded.value = !expanded.value;
}

// Build the path of a child node. Object keys append with a dot (or bare
// when at root); array indices append in brackets.
function childPath(seg: string | number): string {
    const base = props.path ?? '';
    if (typeof seg === 'number') return `${base}[${seg}]`;
    return base ? `${base}.${seg}` : seg;
}

function saveSelf() {
    if (!onSavePath) return;
    onSavePath(props.path ?? '', props.value);
}
</script>

<template>
    <div class="node">
        <!-- Primitive: one line. -->
        <div v-if="!isContainer" class="line">
            <span v-if="keyName !== undefined" class="key">"{{ keyName }}":&nbsp;</span><span :class="`val v-${type}`">{{ formatted }}</span><span class="punct">{{ comma }}</span><button
                v-if="onSavePath && type !== 'null'"
                class="save-as-var"
                :title="`Save as project variable (${path || '$'})`"
                @click="saveSelf"
            >+ var</button>
        </div>

        <!-- Container collapsed (or empty). -->
        <div v-else-if="!expanded || isEmpty" class="line">
            <span v-if="keyName !== undefined" class="key">"{{ keyName }}":&nbsp;</span><span class="caret" :class="{disabled: isEmpty}" @click="toggle">{{ isEmpty ? '·' : '▸' }}</span><span class="bracket">{{ open }}</span><span v-if="!isEmpty" class="summary clickable" @click="toggle">{{ summary }}</span><span class="bracket">{{ close }}</span><span class="punct">{{ comma }}</span>
        </div>

        <!-- Container expanded. -->
        <template v-else>
            <div class="line">
                <span v-if="keyName !== undefined" class="key">"{{ keyName }}":&nbsp;</span><span class="caret" @click="toggle">▾</span><span class="bracket">{{ open }}</span>
            </div>
            <div class="children">
                <JsonNode
                    v-for="(e, i) in entries"
                    :key="i"
                    :value="e.v"
                    :key-name="e.k"
                    :is-last="i === entries.length - 1"
                    :depth="depth + 1"
                    :path="childPath(type === 'array' ? i : (e.k as string))"
                />
            </div>
            <div class="line">
                <span class="bracket">{{ close }}</span><span class="punct">{{ comma }}</span>
            </div>
        </template>
    </div>
</template>

<style scoped>
.node {
    line-height: 1.55;
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
}

.line {
    white-space: pre-wrap;
    word-break: break-word;
    overflow-wrap: anywhere;
}

.key { color: #5ac8fa; }
.val.v-string  { color: #4cd964; }
.val.v-number  { color: #ff9500; }
.val.v-boolean { color: #af52de; }
.val.v-null    { color: var(--text-dim); font-style: italic; }
.bracket { color: var(--text); }
.punct { color: var(--text-dim); }

.caret {
    display: inline-block;
    width: 14px;
    color: var(--text-dim);
    cursor: pointer;
    user-select: none;
    text-align: center;
}
.caret:hover:not(.disabled) { color: var(--text); }
.caret.disabled { cursor: default; opacity: 0.5; }

.clickable { cursor: pointer; user-select: none; }
.summary {
    color: var(--text-dim);
    font-style: italic;
    margin: 0 6px;
}
.summary:hover { color: var(--text); }

.children {
    padding-left: 18px;
    border-left: 1px dashed var(--border);
    margin-left: 6px;
}

.save-as-var {
    margin-left: 8px;
    padding: 0 6px;
    font-size: 10px;
    line-height: 16px;
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 3px;
    color: var(--text-dim);
    cursor: pointer;
    opacity: 0;
    transition: opacity 80ms ease;
    font-family: inherit;
}

.line:hover .save-as-var {
    opacity: 1;
}

.save-as-var:hover {
    color: var(--accent);
    border-color: var(--accent);
}
</style>
