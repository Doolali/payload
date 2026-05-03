// Singleton dialog state used in place of window.prompt / window.confirm,
// which are no-ops in Wails' macOS WKWebview. Components call promptText()
// or confirmAction() and await the user's response. App.vue renders the
// dialog hosts based on this shared state.
import {reactive} from 'vue';

interface PromptState {
    open: boolean;
    title: string;
    label: string;
    placeholder: string;
    initial: string;
    confirmText: string;
    resolve: ((v: string | null) => void) | null;
}

interface ConfirmState {
    open: boolean;
    title: string;
    message: string;
    confirmText: string;
    danger: boolean;
    resolve: ((v: boolean) => void) | null;
}

const promptState = reactive<PromptState>({
    open: false,
    title: '',
    label: '',
    placeholder: '',
    initial: '',
    confirmText: 'OK',
    resolve: null,
});

const confirmState = reactive<ConfirmState>({
    open: false,
    title: '',
    message: '',
    confirmText: 'OK',
    danger: false,
    resolve: null,
});

interface PromptOptions {
    title: string;
    label?: string;
    placeholder?: string;
    initial?: string;
    confirmText?: string;
}

export function promptText(opts: PromptOptions): Promise<string | null> {
    promptState.title = opts.title;
    promptState.label = opts.label ?? '';
    promptState.placeholder = opts.placeholder ?? '';
    promptState.initial = opts.initial ?? '';
    promptState.confirmText = opts.confirmText ?? 'OK';
    promptState.open = true;
    return new Promise(resolve => {
        promptState.resolve = resolve;
    });
}

export function resolvePrompt(value: string | null) {
    promptState.resolve?.(value);
    promptState.resolve = null;
    promptState.open = false;
}

interface ConfirmOptions {
    title?: string;
    message: string;
    confirmText?: string;
    danger?: boolean;
}

export function confirmAction(opts: ConfirmOptions): Promise<boolean> {
    confirmState.title = opts.title ?? 'Are you sure?';
    confirmState.message = opts.message;
    confirmState.confirmText = opts.confirmText ?? 'Confirm';
    confirmState.danger = opts.danger ?? false;
    confirmState.open = true;
    return new Promise(resolve => {
        confirmState.resolve = resolve;
    });
}

export function resolveConfirm(value: boolean) {
    confirmState.resolve?.(value);
    confirmState.resolve = null;
    confirmState.open = false;
}

export function useDialogs() {
    return {promptState, confirmState, resolvePrompt, resolveConfirm};
}
