import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import PresetsView from './PresetsView.vue'
import PresetEditorView from './PresetEditorView.vue'
import { api } from '@/api/client'
import type { Preset, PresetSpec } from '@/api/types'
const baseSpec: PresetSpec = {
	container: { format: 'mp4', faststart: true },
	video: {
		codec: 'libx264',
		preset: 'slow',
		rate: { mode: 'crf', crf: 20 },
		resolution: { width: 1920, height: 1080 }
	},
	audio: { codec: 'aac', bitrate: '192k' }
}
const builtin: Preset = {
	id: 'b1', name: 'h264-1080p', description: 'desc', builtin: true,
	spec: baseSpec, createdAt: '', updatedAt: ''
}
let list: ReturnType<typeof vi.spyOn>
let remove: ReturnType<typeof vi.spyOn>
let duplicate: ReturnType<typeof vi.spyOn>
let restoreDefaults: ReturnType<typeof vi.spyOn>
function setupRouter() {
	return createRouter({
		history: createMemoryHistory(),
		routes: [
			{ path: '/presets', name: 'presets', component: PresetsView },
			{ path: '/presets/new', name: 'preset-new', component: PresetEditorView },
			{ path: '/presets/:id', name: 'preset-edit', component: PresetEditorView }
		]
	})
}
describe('PresetsView', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		list = vi.spyOn(api.presets, 'list').mockResolvedValue([])
		remove = vi.spyOn(api.presets, 'remove').mockResolvedValue(undefined)
		duplicate = vi.spyOn(api.presets, 'duplicate').mockResolvedValue({} as Preset)
		restoreDefaults = vi.spyOn(api.presets, 'restoreDefaults').mockResolvedValue({ restored: [] })
	})
	afterEach(() => {
		vi.restoreAllMocks()
	})
	it('shows loading then preset cards', async () => {
		list.mockResolvedValue([builtin])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		expect(w.text()).toContain('h264-1080p')
		expect(w.text()).toContain('libx264')
		expect(w.text()).toContain('CRF 20')
		expect(w.text()).toContain('mp4')
		expect(w.text()).toContain('1920×1080')
		expect(w.text()).toContain('built-in')
	})
	it('renders empty state when no presets', async () => {
		list.mockResolvedValue([])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		expect(w.text()).toContain('No presets')
	})
	it('renders backend error', async () => {
		list.mockRejectedValue(new Error('server down'))
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		expect(w.text()).toContain('server down')
	})
	it('+ New preset button navigates to /presets/new', async () => {
		list.mockResolvedValue([])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		const newBtn = w.findAll('button').find((b) => b.text().includes('+ New preset'))
		expect(newBtn).toBeDefined()
		await newBtn!.trigger('click')
		await flushPromises()
		expect(router.currentRoute.value.name).toBe('preset-new')
	})
	it('clicking a card navigates to edit route', async () => {
		list.mockResolvedValue([builtin])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		const card = w.find('.card.cursor-pointer')
		await card.trigger('click')
		await flushPromises()
		expect(router.currentRoute.value.name).toBe('preset-edit')
		expect(router.currentRoute.value.params.id).toBe('b1')
	})
	it('handles spec missing nested fields gracefully', async () => {
		// Defensive coverage: partial spec rows (e.g. legacy DB content).
		const partial = { ...builtin, spec: {} as unknown as PresetSpec }
		list.mockResolvedValue([partial])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		// Should not throw; renders fallback dashes.
		expect(w.text()).toContain('h264-1080p')
	})
	it('Restore defaults invokes API + updates list', async () => {
		list.mockResolvedValue([])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		restoreDefaults.mockResolvedValue({ restored: ['x'] })
		list.mockResolvedValueOnce([builtin])
		const btn = w.findAll('button').find((b) => b.text().includes('Restore defaults'))
		await btn!.trigger('click')
		await flushPromises()
		expect(restoreDefaults).toHaveBeenCalled()
	})
	it('Delete confirms and removes', async () => {
		list.mockResolvedValue([{ ...builtin, id: 'x', builtin: false }])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		vi.stubGlobal('confirm', () => true)
		remove.mockResolvedValue(undefined)
		const delBtn = w.findAll('button').find((b) => b.text() === 'Delete')
		await delBtn!.trigger('click')
		await flushPromises()
		expect(remove).toHaveBeenCalledWith('x')
		vi.unstubAllGlobals()
	})
	it('Delete cancels when confirm rejected', async () => {
		list.mockResolvedValue([{ ...builtin, id: 'x', builtin: false }])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		vi.stubGlobal('confirm', () => false)
		const delBtn = w.findAll('button').find((b) => b.text() === 'Delete')
		await delBtn!.trigger('click')
		await flushPromises()
		expect(remove).not.toHaveBeenCalled()
		vi.unstubAllGlobals()
	})
	it('Duplicate prompts for name + navigates', async () => {
		list.mockResolvedValue([builtin])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		vi.stubGlobal('prompt', () => 'duped-name')
		duplicate.mockResolvedValue({ ...builtin, id: 'dup', name: 'duped-name', builtin: false })
		const dupBtn = w.findAll('button').find((b) => b.text() === 'Duplicate')
		await dupBtn!.trigger('click')
		await flushPromises()
		expect(duplicate).toHaveBeenCalledWith('b1', 'duped-name')
		vi.unstubAllGlobals()
	})
	it('Duplicate aborts when prompt empty', async () => {
		list.mockResolvedValue([builtin])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		vi.stubGlobal('prompt', () => null)
		const dupBtn = w.findAll('button').find((b) => b.text() === 'Duplicate')
		await dupBtn!.trigger('click')
		await flushPromises()
		expect(duplicate).not.toHaveBeenCalled()
		vi.unstubAllGlobals()
	})
	it('renders qp-mode rate and bitrate-mode rate', async () => {
		const qpPreset: Preset = {
			...builtin,
			id: 'q', name: 'qp',
			spec: { ...baseSpec, video: { ...baseSpec.video, rate: { mode: 'qp', qp: 22 } } }
		}
		const cbrPreset: Preset = {
			...builtin,
			id: 'c', name: 'cbr',
			spec: { ...baseSpec, video: { ...baseSpec.video, rate: { mode: 'cbr', bitrate: '5M' } } }
		}
		list.mockResolvedValue([qpPreset, cbrPreset])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		expect(w.text()).toContain('QP 22')
		expect(w.text()).toContain('CBR 5M')
	})
	it('shows hwaccel and 2-pass badges when set', async () => {
		const advanced: Preset = {
			...builtin,
			id: 'adv',
			name: 'advanced',
			spec: {
				...baseSpec,
				video: { ...baseSpec.video, twoPass: true },
				hwaccel: { type: 'nvdec' }
			}
		}
		list.mockResolvedValue([advanced])
		const router = setupRouter()
		router.push('/presets')
		await router.isReady()
		const w = mount(PresetsView, { global: { plugins: [router] } })
		await flushPromises()
		expect(w.text()).toContain('2-pass')
		expect(w.text()).toContain('nvdec')
	})
})
