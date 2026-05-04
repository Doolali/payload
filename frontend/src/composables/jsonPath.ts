// Tiny JSON-path resolver for the extractor feature. Supports dotted keys
// and bracketed indices: "data.token", "items[0].id", "list[2][1].name".
// Unknown paths yield undefined; never throws.
export function resolvePath(root: unknown, path: string): unknown {
    const parts = path.match(/[^.[\]]+/g);
    if (!parts) return path === '' ? root : undefined;
    let cur: any = root;
    for (const p of parts) {
        if (cur == null) return undefined;
        const idx = /^\d+$/.test(p) ? parseInt(p, 10) : null;
        cur = idx !== null ? cur[idx] : cur[p];
    }
    return cur;
}

// Choose a sensible default variable name from a path: the trailing key
// segment, or "var" if none exists.
export function defaultVarName(path: string): string {
    const parts = path.match(/[^.[\]]+/g) ?? [];
    for (let i = parts.length - 1; i >= 0; i--) {
        if (!/^\d+$/.test(parts[i])) return parts[i];
    }
    return 'var';
}

// Convert a value to the string we store in project.variables.
export function toVarString(v: unknown): string {
    if (v === null || v === undefined) return '';
    if (typeof v === 'string') return v;
    if (typeof v === 'object') return JSON.stringify(v);
    return String(v);
}
