import type { Preset, PresetSpec, Job, JobLog, SystemInfo } from './types'

class HTTPError extends Error {
	constructor(public status: number, message: string) {
		super(message)
	}
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const r = await fetch(path, {
		...init,
		headers: {
			'Content-Type': 'application/json',
			...(init?.headers ?? {})
		}
	})
	if (!r.ok) {
		const text = await r.text().catch(() => r.statusText)
		throw new HTTPError(r.status, text || r.statusText)
	}
	if (r.status === 204) return undefined as T
	return (await r.json()) as T
}

export const api = {
	system: {
		info: () => request<SystemInfo>('/api/system/info')
	},
	presets: {
		list: () => request<Preset[]>('/api/presets'),
		get: (id: string) => request<Preset>(`/api/presets/${id}`),
		create: (body: { name: string; description?: string; spec: PresetSpec }) =>
			request<Preset>('/api/presets', { method: 'POST', body: JSON.stringify(body) }),
		update: (id: string, body: { name?: string; description?: string; spec?: PresetSpec }) =>
			request<Preset>(`/api/presets/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
		remove: (id: string) =>
			request<void>(`/api/presets/${id}`, { method: 'DELETE' }),
		duplicate: (id: string, name: string) =>
			request<Preset>(`/api/presets/${id}/duplicate`, {
				method: 'POST',
				body: JSON.stringify({ name })
			}),
		restoreDefaults: () =>
			request<{ restored: string[] }>('/api/presets/restore-defaults', { method: 'POST' })
	},
	jobs: {
		list: (params: { status?: string; limit?: number } = {}) => {
			const q = new URLSearchParams()
			if (params.status) q.set('status', params.status)
			if (params.limit) q.set('limit', String(params.limit))
			const qs = q.toString()
			return request<Job[]>(`/api/jobs${qs ? `?${qs}` : ''}`)
		},
		get: (id: string) => request<Job>(`/api/jobs/${id}`),
		create: (body: {
			presetId: string
			sourcePath: string
			outputPath?: string
			cachePath?: string
			sourceMovePath?: string
		}) =>
			request<Job>('/api/jobs', { method: 'POST', body: JSON.stringify(body) }),
		cancel: (id: string) =>
			request<void>(`/api/jobs/${id}`, { method: 'DELETE' }),
		logs: (id: string, fromSeq = 0, limit = 200) => {
			const q = new URLSearchParams()
			if (fromSeq) q.set('from', String(fromSeq))
			if (limit) q.set('limit', String(limit))
			const qs = q.toString()
			return request<JobLog[]>(`/api/jobs/${id}/logs${qs ? `?${qs}` : ''}`)
		}
	}
}

export { HTTPError }
