import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { api } from '@/api/client';
export const usePresetsStore = defineStore('presets', () => {
    const items = ref([]);
    const loading = ref(false);
    const error = ref(null);
    const byId = computed(() => Object.fromEntries(items.value.map((p) => [p.id, p])));
    async function load() {
        loading.value = true;
        error.value = null;
        try {
            items.value = await api.presets.list();
        }
        catch (e) {
            error.value = e instanceof Error ? e.message : String(e);
        }
        finally {
            loading.value = false;
        }
    }
    async function create(body) {
        const p = await api.presets.create(body);
        items.value = [...items.value, p].sort((a, b) => a.name.localeCompare(b.name));
        return p;
    }
    async function update(id, body) {
        const p = await api.presets.update(id, body);
        items.value = items.value.map((x) => (x.id === id ? p : x));
        return p;
    }
    async function remove(id) {
        await api.presets.remove(id);
        items.value = items.value.filter((x) => x.id !== id);
    }
    async function duplicate(id, name) {
        const p = await api.presets.duplicate(id, name);
        items.value = [...items.value, p].sort((a, b) => a.name.localeCompare(b.name));
        return p;
    }
    async function restoreDefaults() {
        const r = await api.presets.restoreDefaults();
        await load();
        return r;
    }
    return { items, loading, error, byId, load, create, update, remove, duplicate, restoreDefaults };
});
