import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { api, HTTPError } from './client'

describe('api client', () => {
	beforeEach(() => {
		vi.stubGlobal('fetch', vi.fn())
	})
	afterEach(() => {
		vi.unstubAllGlobals()
	})

	function mockResponse(body: unknown, init: { status?: number } = {}) {
		const status = init.status ?? 200
		return {
			ok: status >= 200 && status < 300,
			status,
			statusText: status === 204 ? 'No Content' : 'OK',
			text: async () => (typeof body === 'string' ? body : JSON.stringify(body)),
			json: async () => body
		} as unknown as Response
	}

	it('GET /api/presets passes the path', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse([]))
		await api.presets.list()
		expect(fetchMock).toHaveBeenCalledWith('/api/presets', expect.objectContaining({}))
	})

	it('GET /api/presets/schema', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(
			mockResponse({ catalog: { videoCodecs: [] }, schema: {} })
		)
		const r = await api.presets.schema()
		expect(r.catalog).toBeDefined()
	})

	it('POST /api/presets/preview serializes spec body', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ argv: ['ffmpeg', '-i'], spec: {} }))
		const spec = {} as never
		await api.presets.preview(spec)
		const init = fetchMock.mock.calls[0][1] as RequestInit
		expect(init.method).toBe('POST')
		expect(init.body).toContain('"spec"')
	})

	it('throws HTTPError on non-2xx', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse('not authorized', { status: 401 }))
		await expect(api.presets.list()).rejects.toBeInstanceOf(HTTPError)
	})

	it('returns undefined on 204', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse(null, { status: 204 }))
		const r = await api.presets.remove('id')
		expect(r).toBeUndefined()
	})

	it('jobs.list builds query string', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse([]))
		await api.jobs.list({ status: 'running', limit: 50 })
		const url = fetchMock.mock.calls[0][0] as string
		expect(url).toContain('status=running')
		expect(url).toContain('limit=50')
	})

	it('jobs.cancel uses force param when set', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse(null, { status: 204 }))
		await api.jobs.cancel('id', { force: true })
		const url = fetchMock.mock.calls[0][0] as string
		expect(url).toContain('?force=true')
	})

	it('presets.duplicate sends new name in body', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ id: 'b', name: 'beta' }))
		await api.presets.duplicate('a', 'beta')
		const init = fetchMock.mock.calls[0][1] as RequestInit
		expect(init.method).toBe('POST')
		expect(init.body).toContain('"beta"')
	})

	it('presets.create sends POST with spec body', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ id: 'x' }))
		await api.presets.create({ name: 'foo', spec: {} as never })
		const [url, init] = fetchMock.mock.calls[0]
		expect(url).toBe('/api/presets')
		expect((init as RequestInit).method).toBe('POST')
	})

	it('presets.update sends PATCH', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ id: 'x' }))
		await api.presets.update('id', { name: 'renamed' })
		const init = fetchMock.mock.calls[0][1] as RequestInit
		expect(init.method).toBe('PATCH')
	})

	it('presets.get builds path with id', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ id: 'abc' }))
		await api.presets.get('abc')
		const url = fetchMock.mock.calls[0][0] as string
		expect(url).toBe('/api/presets/abc')
	})

	it('presets.restoreDefaults POSTs', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ restored: [] }))
		await api.presets.restoreDefaults()
		const [url, init] = fetchMock.mock.calls[0]
		expect(url).toBe('/api/presets/restore-defaults')
		expect((init as RequestInit).method).toBe('POST')
	})

	it('jobs.create posts source path', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ id: 'j' }))
		await api.jobs.create({ presetId: 'p', sourcePath: '/x.mov' })
		const init = fetchMock.mock.calls[0][1] as RequestInit
		expect(init.method).toBe('POST')
		expect(init.body).toContain('/x.mov')
	})

	it('jobs.get builds path with id', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({}))
		await api.jobs.get('jid')
		expect(fetchMock.mock.calls[0][0]).toBe('/api/jobs/jid')
	})

	it('jobs.logs constructs query string with from + limit', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse([]))
		await api.jobs.logs('jid', 10, 50)
		const url = fetchMock.mock.calls[0][0] as string
		expect(url).toContain('from=10')
		expect(url).toContain('limit=50')
	})

	it('jobs.cancel without force omits query', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse(null, { status: 204 }))
		await api.jobs.cancel('jid')
		const url = fetchMock.mock.calls[0][0] as string
		expect(url).toBe('/api/jobs/jid')
	})

	it('system.info hits /api/system/info', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({}))
		await api.system.info()
		expect(fetchMock.mock.calls[0][0]).toBe('/api/system/info')
	})

	it('workflows.list returns array', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse([]))
		const r = await api.workflows.list()
		expect(Array.isArray(r)).toBe(true)
	})

	it('workflows.create posts body', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ id: 'w' }))
		await api.workflows.create({
			name: 'nightly',
			enabled: true,
			trigger: { type: 'file_created', watchDir: '/x', glob: '*.mp4' },
			actions: [{ type: 'enqueue_transcode', presetId: 'p' }]
		})
		const init = fetchMock.mock.calls[0][1] as RequestInit
		expect(init.method).toBe('POST')
	})

	it('workflows.update sends PATCH', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse({ id: 'w' }))
		await api.workflows.update('w', { enabled: false })
		const init = fetchMock.mock.calls[0][1] as RequestInit
		expect(init.method).toBe('PATCH')
	})

	it('workflows.enable / disable POST', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse(null, { status: 204 }))
		await api.workflows.enable('w')
		await api.workflows.disable('w')
		expect(fetchMock).toHaveBeenCalledTimes(2)
		expect((fetchMock.mock.calls[0][1] as RequestInit).method).toBe('POST')
		expect((fetchMock.mock.calls[1][1] as RequestInit).method).toBe('POST')
	})

	it('workflows.runs uses limit', async () => {
		const fetchMock = vi.mocked(fetch)
		fetchMock.mockResolvedValue(mockResponse([]))
		await api.workflows.runs('w', 25)
		expect(fetchMock.mock.calls[0][0]).toContain('limit=25')
	})

	it('HTTPError exposes status code', () => {
		const err = new HTTPError(503, 'down')
		expect(err.status).toBe(503)
		expect(err.message).toBe('down')
	})
})
