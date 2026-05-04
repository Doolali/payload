<script setup lang="ts">
// Custom-styled autocomplete input. Replaces HTML5 <datalist> in places
// where the WKWebView quirk (popup doesn't dismiss on click) gets in the
// way. Filters options by substring as the user types, supports keyboard
// nav (↑/↓/Enter/Esc), closes on blur or selection.
import {computed, ref} from 'vue';

const props = defineProps<{
    modelValue: string;
    options: string[];
    placeholder?: string;
    maxResults?: number;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', v: string): void;
    (e: 'change'): void;
}>();

const open = ref(false);
const highlight = ref(-1);
const inputRef = ref<HTMLInputElement | null>(null);

const filtered = computed(() => {
    const limit = props.maxResults ?? 12;
    const q = props.modelValue.toLowerCase().trim();
    if (!q) return props.options.slice(0, limit);
    return props.options
        .filter(o => o.toLowerCase().includes(q))
        .slice(0, limit);
});

function set(v: string) {
    emit('update:modelValue', v);
    emit('change');
}

function onInput(ev: Event) {
    set((ev.target as HTMLInputElement).value);
    open.value = true;
    highlight.value = -1;
}

function onFocus() {
    open.value = true;
}

// Delay close so a mousedown on an option lands first; mousedown.prevent on
// the option also blocks the blur, but the timeout makes other dismissals
// (clicking elsewhere in the document) feel snappy.
function onBlur() {
    setTimeout(() => { open.value = false; }, 120);
}

function pick(opt: string) {
    set(opt);
    open.value = false;
    inputRef.value?.blur();
}

function onKeydown(ev: KeyboardEvent) {
    if (ev.key === 'Escape') {
        open.value = false;
        return;
    }
    if (!open.value || !filtered.value.length) {
        if (ev.key === 'ArrowDown') {
            open.value = true;
            highlight.value = 0;
            ev.preventDefault();
        }
        return;
    }
    if (ev.key === 'ArrowDown') {
        ev.preventDefault();
        highlight.value = (highlight.value + 1) % filtered.value.length;
    } else if (ev.key === 'ArrowUp') {
        ev.preventDefault();
        highlight.value = (highlight.value - 1 + filtered.value.length) % filtered.value.length;
    } else if (ev.key === 'Enter') {
        if (highlight.value >= 0) {
            ev.preventDefault();
            pick(filtered.value[highlight.value]);
        }
    } else if (ev.key === 'Tab' && highlight.value >= 0) {
        pick(filtered.value[highlight.value]);
    }
}
</script>

<template>
    <div class="autocomplete">
        <input
            ref="inputRef"
            type="text"
            :value="modelValue"
            :placeholder="placeholder"
            spellcheck="false"
            autocomplete="off"
            @input="onInput"
            @focus="onFocus"
            @blur="onBlur"
            @keydown="onKeydown"
        />
        <div v-if="open && filtered.length" class="popup">
            <div
                v-for="(opt, i) in filtered"
                :key="opt"
                class="opt"
                :class="{highlighted: i === highlight}"
                @mousedown.prevent="pick(opt)"
                @mouseenter="highlight = i"
            >{{ opt }}</div>
        </div>
    </div>
</template>

<style scoped>
.autocomplete {
    position: relative;
    width: 100%;
}

input {
    border: none;
    background: transparent;
    border-radius: 0;
    padding: 8px 10px;
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    width: 100%;
    color: var(--text);
}

input:focus {
    background: var(--bg);
    outline: 1px solid var(--accent);
    outline-offset: -1px;
}

.popup {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 4px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
    max-height: 240px;
    overflow-y: auto;
    z-index: 50;
    margin-top: 2px;
}

.opt {
    padding: 6px 10px;
    cursor: pointer;
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    color: var(--text);
}

.opt.highlighted {
    background: var(--bg-active);
    color: var(--text);
}
</style>
