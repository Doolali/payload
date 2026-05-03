<script setup lang="ts">
// Edit a request's body. Type switcher + a single textarea for the content.
// JSON gets a "Format" button that pretty-prints whatever the user typed.
import {model} from '../../wailsjs/go/models';

const props = defineProps<{
    body: model.Body;
}>();

const emit = defineEmits<{
    (e: 'change'): void;
}>();

const TYPES: Array<{value: string; label: string}> = [
    {value: 'none', label: 'none'},
    {value: 'json', label: 'JSON'},
    {value: 'text', label: 'Text'},
    {value: 'urlencoded', label: 'x-www-form-urlencoded'},
];

function setType(t: string) {
    props.body.type = t;
    emit('change');
}

function setContent(v: string) {
    props.body.content = v;
    emit('change');
}

function format() {
    if (props.body.type !== 'json') return;
    try {
        const parsed = JSON.parse(props.body.content || 'null');
        props.body.content = JSON.stringify(parsed, null, 2);
        emit('change');
    } catch {
        // ignore — leave the user's content alone if it isn't valid JSON yet
    }
}
</script>

<template>
    <div class="body-editor">
        <div class="toolbar">
            <div class="types">
                <label v-for="t in TYPES" :key="t.value">
                    <input
                        type="radio"
                        :value="t.value"
                        :checked="body.type === t.value"
                        @change="setType(t.value)"
                    />
                    <span>{{ t.label }}</span>
                </label>
            </div>
            <button
                v-if="body.type === 'json'"
                class="ghost"
                @click="format"
            >Format JSON</button>
        </div>
        <textarea
            v-if="body.type !== 'none'"
            :value="body.content"
            :placeholder="body.type === 'json' ? '{\n  \&quot;key\&quot;: \&quot;value\&quot;\n}' : ''"
            spellcheck="false"
            @input="setContent(($event.target as HTMLTextAreaElement).value)"
        />
        <div v-else class="empty">This request has no body.</div>
    </div>
</template>

<style scoped>
.body-editor {
    display: flex;
    flex-direction: column;
    height: 100%;
}

.toolbar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 4px;
    border-bottom: 1px solid var(--border);
}

.types {
    display: flex;
    gap: 12px;
    flex: 1;
}

.types label {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    cursor: pointer;
    user-select: none;
}

textarea {
    flex: 1;
    min-height: 200px;
    border: none;
    background: var(--bg);
    color: var(--text);
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 12px;
    padding: 12px;
    resize: none;
    outline: none;
}

.empty {
    padding: 24px;
    text-align: center;
    color: var(--text-dim);
    font-size: 12px;
}
</style>
