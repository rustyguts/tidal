import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import type { Catalog, CatalogVideoCodec, CatalogAudioCodec, CatalogFilter, CatalogHwaccel } from '@/api/types'

export const usePresetCatalogStore = defineStore('presetCatalog', () => {
	const catalog = ref<Catalog | null>(null)
	const loading = ref(false)
	const error = ref<string | null>(null)

	async function load() {
		if (catalog.value || loading.value) return
		loading.value = true
		error.value = null
		try {
			const r = await api.presets.schema()
			catalog.value = r.catalog
		} catch (e) {
			error.value = e instanceof Error ? e.message : String(e)
		} finally {
			loading.value = false
		}
	}

	const videoCodecs = computed(() => catalog.value?.videoCodecs ?? [])
	const audioCodecs = computed(() => catalog.value?.audioCodecs ?? [])
	const containers = computed(() => catalog.value?.containers ?? [])
	const hwaccels = computed(() => catalog.value?.hwaccels ?? [])
	const filters = computed(() => catalog.value?.filters ?? [])
	const pixelFormats = computed(() => catalog.value?.pixelFormats ?? [])

	function videoCodec(name: string): CatalogVideoCodec | undefined {
		return videoCodecs.value.find((c) => c.name === name)
	}
	function audioCodec(name: string): CatalogAudioCodec | undefined {
		return audioCodecs.value.find((c) => c.name === name)
	}
	function hwaccel(type: string): CatalogHwaccel | undefined {
		return hwaccels.value.find((h) => h.type === type)
	}
	function filterSpec(name: string): CatalogFilter | undefined {
		return filters.value.find((f) => f.name === name)
	}
	function filtersByScope(scope: 'video' | 'audio'): CatalogFilter[] {
		return filters.value.filter((f) => f.scope === scope)
	}

	return {
		catalog,
		loading,
		error,
		videoCodecs,
		audioCodecs,
		containers,
		hwaccels,
		filters,
		pixelFormats,
		videoCodec,
		audioCodec,
		hwaccel,
		filterSpec,
		filtersByScope,
		load
	}
})
