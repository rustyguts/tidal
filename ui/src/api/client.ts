import type {
	Preset,
	PresetSpec,
	PreviewResponse,
	SchemaResponse,
	Job,
	JobLog,
	SystemInfo,
	Workflow,
	WorkflowRun,
	WorkflowTrigger,
	WorkflowAction
} from './types'

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
	settings: {
		get: () => request<{ transcodeConcurrency: number }>('/api/settings'),
		update: (body: { transcodeConcurrency?: number }) =>
			request<{ transcodeConcurrency: number }>('/api/settings', {
				method: 'PATCH',
				body: JSON.stringify(body)
			})
	},
	presets: {
		list: () => request<Preset[]>('/api/presets'),
		get: (id: string) => request<Preset>(`/api/presets/${id}`),
		create: (body: {
			name: string
			description?: string
			spec: PresetSpec
			outputPath?: string
			cachePath?: string
			sourceMovePath?: string
		}) =>
			request<Preset>('/api/presets', { method: 'POST', body: JSON.stringify(body) }),
		update: (
			id: string,
			body: {
				name?: string
				description?: string
				spec?: PresetSpec
				outputPath?: string
				cachePath?: string
				sourceMovePath?: string
			}
		) =>
			request<Preset>(`/api/presets/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
		remove: (id: string) =>
			request<void>(`/api/presets/${id}`, { method: 'DELETE' }),
		duplicate: (id: string, name: string) =>
			request<Preset>(`/api/presets/${id}/duplicate`, {
				method: 'POST',
				body: JSON.stringify({ name })
			}),
		restoreDefaults: () =>
			request<{ restored: string[] }>('/api/presets/restore-defaults', { method: 'POST' }),
		schema: () => request<SchemaResponse>('/api/presets/schema'),
		preview: (spec: PresetSpec) =>
			request<PreviewResponse>('/api/presets/preview', {
				method: 'POST',
				body: JSON.stringify({ spec })
			})
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
		cancel: (id: string, opts: { force?: boolean } = {}) => {
			const q = opts.force ? '?force=true' : ''
			return request<void>(`/api/jobs/${id}${q}`, { method: 'DELETE' })
		},
		logs: (id: string, fromSeq = 0, limit = 200) => {
			const q = new URLSearchParams()
			if (fromSeq) q.set('from', String(fromSeq))
			if (limit) q.set('limit', String(limit))
			const qs = q.toString()
			return request<JobLog[]>(`/api/jobs/${id}/logs${qs ? `?${qs}` : ''}`)
		}
	},
	workflows: {
		list: () => request<Workflow[]>('/api/workflows'),
		get: (id: string) => request<Workflow>(`/api/workflows/${id}`),
		create: (body: {
			name: string
			enabled: boolean
			trigger: WorkflowTrigger
			actions: WorkflowAction[]
			pollIntervalMs?: number
			stableThresholdMs?: number
		}) => request<Workflow>('/api/workflows', { method: 'POST', body: JSON.stringify(body) }),
		update: (
			id: string,
			body: Partial<{
				name: string
				enabled: boolean
				trigger: WorkflowTrigger
				actions: WorkflowAction[]
				pollIntervalMs: number
				stableThresholdMs: number
			}>
		) => request<Workflow>(`/api/workflows/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
		remove: (id: string) => request<void>(`/api/workflows/${id}`, { method: 'DELETE' }),
		enable: (id: string) =>
			request<void>(`/api/workflows/${id}/enable`, { method: 'POST' }),
		disable: (id: string) =>
			request<void>(`/api/workflows/${id}/disable`, { method: 'POST' }),
		runs: (id: string, limit = 50) =>
			request<WorkflowRun[]>(`/api/workflows/${id}/runs?limit=${limit}`)
	}
}

export { HTTPError }
