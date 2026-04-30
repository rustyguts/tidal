class HTTPError extends Error {
    status;
    constructor(status, message) {
        super(message);
        this.status = status;
    }
}
async function request(path, init) {
    const r = await fetch(path, {
        ...init,
        headers: {
            'Content-Type': 'application/json',
            ...(init?.headers ?? {})
        }
    });
    if (!r.ok) {
        const text = await r.text().catch(() => r.statusText);
        throw new HTTPError(r.status, text || r.statusText);
    }
    if (r.status === 204)
        return undefined;
    return (await r.json());
}
export const api = {
    system: {
        info: () => request('/api/system/info')
    },
    presets: {
        list: () => request('/api/presets'),
        get: (id) => request(`/api/presets/${id}`),
        create: (body) => request('/api/presets', { method: 'POST', body: JSON.stringify(body) }),
        update: (id, body) => request(`/api/presets/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
        remove: (id) => request(`/api/presets/${id}`, { method: 'DELETE' }),
        duplicate: (id, name) => request(`/api/presets/${id}/duplicate`, {
            method: 'POST',
            body: JSON.stringify({ name })
        }),
        restoreDefaults: () => request('/api/presets/restore-defaults', { method: 'POST' })
    },
    jobs: {
        list: (params = {}) => {
            const q = new URLSearchParams();
            if (params.status)
                q.set('status', params.status);
            if (params.limit)
                q.set('limit', String(params.limit));
            const qs = q.toString();
            return request(`/api/jobs${qs ? `?${qs}` : ''}`);
        },
        get: (id) => request(`/api/jobs/${id}`),
        create: (body) => request('/api/jobs', { method: 'POST', body: JSON.stringify(body) }),
        cancel: (id) => request(`/api/jobs/${id}`, { method: 'DELETE' }),
        logs: (id, fromSeq = 0, limit = 200) => {
            const q = new URLSearchParams();
            if (fromSeq)
                q.set('from', String(fromSeq));
            if (limit)
                q.set('limit', String(limit));
            const qs = q.toString();
            return request(`/api/jobs/${id}/logs${qs ? `?${qs}` : ''}`);
        }
    }
};
export { HTTPError };
