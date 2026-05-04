<script setup lang="ts">
// Project settings modal. Lets the user rename the project (in-memory edit
// + autosave) and move its JSON file to a different directory (backend
// move + index update). The path field is read-only — the move action
// is the way to change it.
import {nextTick, onMounted, ref} from 'vue';
import {model} from '../../wailsjs/go/models';
import {MoveProject, ProjectPath, RenameProject} from '../../wailsjs/go/main/App';

const props = defineProps<{
    project: model.Project;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'updated', project: model.Project): void;
}>();

const name = ref(props.project.name);
const path = ref('');
const saving = ref(false);
const moving = ref(false);
const error = ref<string | null>(null);
const nameInput = ref<HTMLInputElement | null>(null);

onMounted(async () => {
    try {
        path.value = await ProjectPath(props.project.id);
    } catch (e: any) {
        error.value = String(e);
    }
    nextTick(() => nameInput.value?.focus());
});

async function save() {
    error.value = null;
    const trimmed = name.value.trim();
    if (!trimmed) return;
    if (trimmed === props.project.name) {
        emit('close');
        return;
    }
    saving.value = true;
    try {
        const updated = await RenameProject(props.project.id, trimmed);
        emit('updated', updated);
        emit('close');
    } catch (e: any) {
        error.value = String(e);
    } finally {
        saving.value = false;
    }
}

async function moveFile() {
    error.value = null;
    moving.value = true;
    try {
        const updated = await MoveProject(props.project.id);
        if (updated) {
            // Refresh the displayed path and let the parent know.
            path.value = await ProjectPath(props.project.id);
            emit('updated', updated);
        }
    } catch (e: any) {
        error.value = String(e);
    } finally {
        moving.value = false;
    }
}
</script>

<template>
    <div class="backdrop" @click.self="emit('close')">
        <div class="dialog">
            <h2>Project settings</h2>
            <form @submit.prevent="save">
                <label>
                    <span>Name</span>
                    <input
                        ref="nameInput"
                        v-model="name"
                        type="text"
                        spellcheck="false"
                        @keydown.escape="emit('close')"
                    />
                </label>

                <label>
                    <span>JSON file</span>
                    <div class="path-row">
                        <input
                            :value="path"
                            type="text"
                            readonly
                            spellcheck="false"
                            class="path"
                        />
                        <button
                            type="button"
                            :disabled="moving"
                            @click="moveFile"
                        >{{ moving ? 'Moving…' : 'Move…' }}</button>
                    </div>
                    <p class="hint">Pick a different directory; the file keeps its current name.</p>
                </label>

                <p v-if="error" class="error">{{ error }}</p>

                <div class="actions">
                    <button type="button" @click="emit('close')">Close</button>
                    <button
                        type="submit"
                        class="primary"
                        :disabled="!name.trim() || saving || name.trim() === project.name"
                    >{{ saving ? 'Saving…' : 'Save' }}</button>
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
    width: 480px;
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

.path-row {
    display: flex;
    gap: 6px;
}

.path {
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    color: var(--text-dim);
}

.hint {
    margin: 6px 0 0;
    font-size: 11px;
    color: var(--text-dim);
}

.error {
    color: var(--danger);
    font-size: 12px;
    margin: 0 0 12px;
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
