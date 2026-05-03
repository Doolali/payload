<script setup lang="ts">
// Modal confirmation dialog. Driven by the singleton state in useDialogs;
// rendered once at App.vue and shown whenever confirmAction() is called.
import {useDialogs} from '../composables/useDialogs';

const {confirmState, resolveConfirm} = useDialogs();

function ok() { resolveConfirm(true); }
function cancel() { resolveConfirm(false); }
</script>

<template>
    <div v-if="confirmState.open" class="backdrop" @click.self="cancel">
        <div class="dialog">
            <h2>{{ confirmState.title }}</h2>
            <p>{{ confirmState.message }}</p>
            <div class="actions">
                <button type="button" @click="cancel">Cancel</button>
                <button
                    type="button"
                    :class="confirmState.danger ? 'danger' : 'primary'"
                    @click="ok"
                >{{ confirmState.confirmText }}</button>
            </div>
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
    width: 380px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

h2 {
    margin: 0 0 12px;
    font-size: 16px;
    font-weight: 600;
}

p {
    margin: 0 0 16px;
    color: var(--text);
    font-size: 13px;
    line-height: 1.5;
}

.actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
}

.danger {
    background: var(--danger);
    border-color: var(--danger);
    color: #fff;
    font-weight: 600;
}

.danger:hover {
    background: #f06a7e;
    border-color: #f06a7e;
}
</style>
