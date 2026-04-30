import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { usePresetCatalogStore } from './presetCatalog'
import type { Catalog, SchemaResponse } from '@/api/types'

const fakeCatalog: Catalog = {
	containers: [
		{ format: 'mp4', mime: 'video/mp4' },
		{ format: 'mkv', mime: 'video/x-matroska' }
	],
	videoCodecs: [
		{
			name: 'libx264',
			family: 'cpu',
			displayName: 'H.264',
			description: 'Software H.264',
			presets: ['slow', 'medium'],
			rateModes: ['crf', 'cbr'],
			allowsTwoPass: true,
			crfRange: { min: 0, max: 51, default: 23 }
		},
		{
			name: 'h264_nvenc',
			family: 'nvenc',
			displayName: 'NVENC H.264',
			rateModes: ['cbr', 'qp'],
			allowsTwoPass: false
		}
	],
	audioCodecs: [
		{ name: 'aac', displayName: 'AAC', allowsVBR: true, sampleRates: [44100, 48000], channels: [1, 2, 6] }
	],
	pixelFormats: ['yuv420p', 'yuv420p10le'],
	hwaccels: [
		{ type: 'none', displayName: 'Software', codecNames: ['libx264'], outputFormats: [] },
		{ type: 'nvdec', displayName: 'NVDEC', codecNames: ['h264_nvenc'], outputFormats: ['cuda'] }
	],
	filters: [
		{ name: 'scale', scope: 'video', displayName: 'Scale', args: [{ name: 'w', type: 'string', required: true }] },
		{ name: 'loudnorm', scope: 'audio', displayName: 'Loudnorm' }
	],
	rawExtrasAllow: [],
	rawExtrasDeny: []
}

const fakeSchema: SchemaResponse = {
	schemaVersion: 2,
	catalog: fakeCatalog,
	schema: { title: 'PresetSpecV2' }
}

vi.mock('@/api/client', () => ({
	api: {
		presets: {
			schema: vi.fn(async () => fakeSchema)
		}
	}
}))

describe('usePresetCatalogStore', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('starts empty before load', () => {
		const s = usePresetCatalogStore()
		expect(s.catalog).toBeNull()
		expect(s.videoCodecs).toEqual([])
	})

	it('loads catalog from /api/presets/schema', async () => {
		const s = usePresetCatalogStore()
		await s.load()
		expect(s.catalog).not.toBeNull()
		expect(s.videoCodecs).toHaveLength(2)
		expect(s.containers).toHaveLength(2)
	})

	it('looks up video codec by name', async () => {
		const s = usePresetCatalogStore()
		await s.load()
		const c = s.videoCodec('libx264')
		expect(c?.displayName).toBe('H.264')
		expect(s.videoCodec('nope')).toBeUndefined()
	})

	it('looks up audio codec by name', async () => {
		const s = usePresetCatalogStore()
		await s.load()
		expect(s.audioCodec('aac')?.displayName).toBe('AAC')
		expect(s.audioCodec('vorbis')).toBeUndefined()
	})

	it('looks up hwaccel by type', async () => {
		const s = usePresetCatalogStore()
		await s.load()
		expect(s.hwaccel('nvdec')?.codecNames).toContain('h264_nvenc')
	})

	it('looks up filter spec', async () => {
		const s = usePresetCatalogStore()
		await s.load()
		const f = s.filterSpec('scale')
		expect(f?.scope).toBe('video')
		expect(f?.args).toHaveLength(1)
	})

	it('partitions filters by scope', async () => {
		const s = usePresetCatalogStore()
		await s.load()
		const v = s.filtersByScope('video')
		const a = s.filtersByScope('audio')
		expect(v).toHaveLength(1)
		expect(a).toHaveLength(1)
		expect(v[0].name).toBe('scale')
		expect(a[0].name).toBe('loudnorm')
	})

	it('skips reload when catalog already cached', async () => {
		const s = usePresetCatalogStore()
		await s.load()
		const first = s.catalog
		await s.load()
		expect(s.catalog).toBe(first)
	})
})
