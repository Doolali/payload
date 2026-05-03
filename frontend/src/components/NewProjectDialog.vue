<script setup lang="ts">
// Modal for entering the name and save location of a new project. The
// location pre-fills with whatever the backend reports as the default
// (initially the in-app projects folder, then the most recently used dir).
// Emits `create` with name + dir, or `cancel` if the user backs out.
import {nextTick, onMounted, ref} from 'vue';
import {ChooseProjectDir, DefaultProjectDir} from '../../wailsjs/go/main/App';

const emit = defineEmits<{
    (e: 'cancel'): void;
    (e: 'create', payload: {name: string; dir: string}): void;
}>();

const name = ref('');
const dir = ref('');
const input = ref<HTMLInputElement | null>(null);
const browsing = ref(false);

onMounted(async () => {
    try {
        dir.value = await DefaultProjectDir();
    } catch {
        // empty dir means "use server default" on the backend
    }
    nextTick(() => input.value?.focus());
});

async function browse() {
    browsing.value = true;
    try {
        const picked = await ChooseProjectDir();
        if (picked) dir.value = picked;
    } finally {
        browsing.value = false;
    }
}

function submit() {
    const trimmed = name.value.trim();
    if (!trimmed) return;
    emit('create', {name: trimmed, dir: dir.value});
}
</script>

<template>
    <div class="backdrop" @click.self="emit('cancel')">
        <div class="dialog">
            <h2>New project</h2>
            <form @submit.prevent="submit">
                <label>
                    <span>Name</span>
                    <input
                        ref="input"
                        v-model="name"
                        type="text"
                        placeholder="My API"
                        @keydown.escape="emit('cancel')"
                    />
                </label>

                <label>
                    <span>Save location</span>
                    <div class="dir-row">
                        <input
                            v-model="dir"
                            type="text"
                            placeholder="Loading default…"
                            spellcheck="false"
                        />
                        <button type="button" :disabled="browsing" @click="browse">Choose…</button>
                    </div>
                    <p class="hint">A JSON file named after the project will be created here.</p>
                </label>

                <div class="actions">
                    <button type="button" @click="emit('cancel')">Cancel</button>
                    <button type="submit" class="primary" :disabled="!name.trim()">Create</button>
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
    z-index: 100;
}

.dialog {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 20px;
    width: 440px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

h2 {
    margin: 0 0 16px;
    font-size: 16px;
    font-weight: 600;
}

label {
    display: block;
    margin-bottom: 14px;
}

label > span {
    display: block;
    margin-bottom: 4px;
    font-size: 12px;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.dir-row {
    display: flex;
    gap: 6px;
}

.dir-row input {
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
}

.dir-row button {
    white-space: nowrap;
}

.hint {
    margin: 6px 0 0;
    font-size: 11px;
    color: var(--text-dim);
}

.actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 8px;
}

button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
</style>
