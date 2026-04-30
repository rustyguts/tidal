import { defineStore } from 'pinia';
import { ref } from 'vue';
import { api } from '@/api/client';
export const useSystemStore = defineStore('system', () => {
    const info = ref(null);
    const loaded = ref(false);
    const error = ref(null);
    async function load() {
        try {
            info.value = await api.system.info();
            error.value = null;
        }
        catch (e) {
            error.value = e instanceof Error ? e.message : String(e);
        }
        finally {
            loaded.value = true;
        }
    }
    return { info, loaded, error, load };
});
