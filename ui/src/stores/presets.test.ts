import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { usePresetsStore } from './presets'
import type { Preset, PresetSpecV2 } from '@/api/types'

function makePreset(id: string, name: string, builtin = false): Preset {
	const spec: PresetSpecV2 = {
		schemaVersion: 2,
		container: { format: 'mp4', faststart: true },
		video: {
			codec: 'libx264',
			rate: { mode: 'crf', crf: 23 }
		},
		audio: { codec: 'aac' }
	}
	return {
		id,
		name,
		description: '',
		builtin,
		spec,
		createdAt: '2026-01-01T00:00:00Z',
		updatedAt: '2026-01-01T00:00:00Z'
	}
}

const list = vi.fn()
const create = vi.fn()
const update = vi.fn()
const remove = vi.fn()
const duplicate = vi.fn()
const restoreDefaults = vi.fn()

vi.mock('@/api/client', () => ({
	api: {
		presets: {
			list: (...a: never[]) => list(...a),
			create: (...a: never[]) => create(...a),
			update: (...a: never[]) => update(...a),
			remove: (...a: never[]) => remove(...a),
			duplicate: (...a: never[]) => duplicate(...a),
			restoreDefaults: (...a: never[]) => restoreDefaults(...a)
		}
	}
}))

describe('usePresetsStore', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		list.mockReset()
		create.mockReset()
		update.mockReset()
		remove.mockReset()
		duplicate.mockReset()
		restoreDefaults.mockReset()
	})

	it('load populates items and clears error', async () => {
		list.mockResolvedValue([makePreset('a', 'alpha'), makePreset('b', 'beta')])
		const s = usePresetsStore()
		await s.load()
		expect(s.items).toHaveLength(2)
		expect(s.error).toBeNull()
	})

	it('load captures error message and stops loading', async () => {
		list.mockRejectedValue(new Error('boom'))
		const s = usePresetsStore()
		await s.load()
		expect(s.error).toBe('boom')
		expect(s.loading).toBe(false)
	})

	it('byId indexes loaded presets', async () => {
		list.mockResolvedValue([makePreset('id-1', 'one'), makePreset('id-2', 'two')])
		const s = usePresetsStore()
		await s.load()
		expect(s.byId['id-1'].name).toBe('one')
		expect(s.byId['id-2'].name).toBe('two')
	})

	it('create appends sorted by name', async () => {
		list.mockResolvedValue([makePreset('a', 'beta')])
		const s = usePresetsStore()
		await s.load()
		create.mockResolvedValue(makePreset('b', 'alpha'))
		await s.create({ name: 'alpha', spec: makePreset('b', 'alpha').spec })
		expect(s.items.map((p) => p.name)).toEqual(['alpha', 'beta'])
	})

	it('update replaces the existing preset in place', async () => {
		const existing = makePreset('id-1', 'orig')
		list.mockResolvedValue([existing])
		const s = usePresetsStore()
		await s.load()
		update.mockResolvedValue({ ...existing, name: 'renamed' })
		await s.update('id-1', { name: 'renamed' })
		expect(s.items[0].name).toBe('renamed')
	})

	it('remove drops the preset locally', async () => {
		list.mockResolvedValue([makePreset('a', 'one'), makePreset('b', 'two')])
		const s = usePresetsStore()
		await s.load()
		remove.mockResolvedValue(undefined)
		await s.remove('a')
		expect(s.items.map((p) => p.id)).toEqual(['b'])
	})

	it('duplicate appends + sorts', async () => {
		list.mockResolvedValue([makePreset('a', 'orig')])
		const s = usePresetsStore()
		await s.load()
		duplicate.mockResolvedValue(makePreset('b', 'aaa-clone'))
		await s.duplicate('a', 'aaa-clone')
		expect(s.items[0].name).toBe('aaa-clone')
	})

	it('restoreDefaults reloads list and returns server payload', async () => {
		list.mockResolvedValueOnce([])
		const s = usePresetsStore()
		await s.load()
		restoreDefaults.mockResolvedValue({ restored: ['h264-1080p'] })
		list.mockResolvedValueOnce([makePreset('a', 'h264-1080p', true)])
		const r = await s.restoreDefaults()
		expect(r.restored).toContain('h264-1080p')
		expect(s.items).toHaveLength(1)
	})
})
