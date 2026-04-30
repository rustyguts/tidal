import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { api } from '@/api/client';
export const usePresetCatalogStore = defineStore('presetCatalog', () => {
    const catalog = ref(null);
    const loading = ref(false);
    const error = ref(null);
    async function load() {
        if (catalog.value || loading.value)
            return;
        loading.value = true;
        error.value = null;
        try {
            const r = await api.presets.schema();
            catalog.value = r.catalog;
        }
        catch (e) {
            error.value = e instanceof Error ? e.message : String(e);
        }
        finally {
            loading.value = false;
        }
    }
    const videoCodecs = computed(() => catalog.value?.videoCodecs ?? []);
    const audioCodecs = computed(() => catalog.value?.audioCodecs ?? []);
    const containers = computed(() => catalog.value?.containers ?? []);
    const hwaccels = computed(() => catalog.value?.hwaccels ?? []);
    const filters = computed(() => catalog.value?.filters ?? []);
    const pixelFormats = computed(() => catalog.value?.pixelFormats ?? []);
    function videoCodec(name) {
        return videoCodecs.value.find((c) => c.name === name);
    }
    function audioCodec(name) {
        return audioCodecs.value.find((c) => c.name === name);
    }
    function hwaccel(type) {
        return hwaccels.value.find((h) => h.type === type);
    }
    function filterSpec(name) {
        return filters.value.find((f) => f.name === name);
    }
    function filtersByScope(scope) {
        return filters.value.filter((f) => f.scope === scope);
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
    };
});
