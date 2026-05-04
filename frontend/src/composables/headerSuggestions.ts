// Common HTTP header names + canonical values, used to back the autocomplete
// datalists in the headers KV editor. Header names are case-insensitive on
// the wire; the values map is keyed lower-case so lookup is straightforward.

export const HEADER_NAMES: string[] = [
    'Accept',
    'Accept-Charset',
    'Accept-Encoding',
    'Accept-Language',
    'Authorization',
    'Cache-Control',
    'Connection',
    'Content-Disposition',
    'Content-Encoding',
    'Content-Language',
    'Content-Length',
    'Content-Location',
    'Content-Type',
    'Cookie',
    'Date',
    'ETag',
    'Expires',
    'Forwarded',
    'From',
    'Host',
    'If-Match',
    'If-Modified-Since',
    'If-None-Match',
    'If-Range',
    'If-Unmodified-Since',
    'Last-Modified',
    'Location',
    'Origin',
    'Pragma',
    'Range',
    'Referer',
    'Server',
    'Set-Cookie',
    'Transfer-Encoding',
    'User-Agent',
    'Vary',
    'WWW-Authenticate',
    'X-Api-Key',
    'X-Csrf-Token',
    'X-Forwarded-For',
    'X-Forwarded-Host',
    'X-Forwarded-Proto',
    'X-Real-IP',
    'X-Request-ID',
    'X-Requested-With',
];

export const HEADER_VALUE_SUGGESTIONS: Record<string, string[]> = {
    'accept': [
        'application/json',
        'application/xml',
        'application/octet-stream',
        'text/html',
        'text/plain',
        'image/*',
        '*/*',
    ],
    'accept-encoding': ['gzip, deflate, br', 'identity'],
    'authorization': ['Bearer ', 'Basic ', 'Token '],
    'cache-control': ['no-cache', 'no-store', 'max-age=0', 'public', 'private'],
    'connection': ['keep-alive', 'close'],
    'content-encoding': ['gzip', 'deflate', 'br', 'identity'],
    'content-type': [
        'application/json',
        'application/x-www-form-urlencoded',
        'application/xml',
        'multipart/form-data',
        'text/plain',
        'text/html',
        'application/octet-stream',
    ],
    'origin': ['*', 'null'],
    'x-requested-with': ['XMLHttpRequest'],
};

export function valuesForHeader(key: string): string[] {
    if (!key) return [];
    return HEADER_VALUE_SUGGESTIONS[key.toLowerCase()] ?? [];
}
