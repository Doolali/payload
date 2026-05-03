<script setup lang="ts">
// Modal text-input prompt. Driven by the singleton state in useDialogs;
// rendered once at App.vue and shown whenever promptText() is called.
import {nextTick, ref, watch} from 'vue';
import {useDialogs} from '../composables/useDialogs';

const {promptState, resolvePrompt} = useDialogs();
const value = ref('');
const input = ref<HTMLInputElement | null>(null);

watch(() => promptState.open, (open) => {
    if (open) {
        value.value = promptState.initial;
        nextTick(() => {
            input.value?.focus();
            input.value?.select();
        });
    }
});

function ok() {
    const v = value.value;
    resolvePrompt(v);
}

function cancel() {
    resolvePrompt(null);
}
</script>

<template>
    <div v-if="promptState.open" class="backdrop" @click.self="cancel">
        <div class="dialog">
            <h2>{{ promptState.title }}</h2>
            <form @submit.prevent="ok">
                <label v-if="promptState.label">{{ promptState.label }}</label>
                <input
                    ref="input"
                    v-model="value"
                    type="text"
                    :placeholder="promptState.placeholder"
                    spellcheck="false"
                    @keydown.escape="cancel"
                />
                <div class="actions">
                    <button type="button" @click="cancel">Cancel</button>
                    <button type="submit" class="primary" :disabled="!value.trim()">{{ promptState.confirmText }}</button>
                </div>
            </form>
        </div>
    </div>
</template>

<style scoped>
.backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
}

.dialog {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 20px;
    width: 360px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

h2 {
    margin: 0 0 16px;
    font-size: 16px;
    font-weight: 600;
}

label {
    display: block;
    font-size: 12px;
    color: var(--text-dim);
    margin-bottom: 4px;
}

.actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 16px;
}

button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
</style>
