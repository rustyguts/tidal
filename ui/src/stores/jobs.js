import { defineStore } from 'pinia';
import { ref, computed, onScopeDispose } from 'vue';
import { api } from '@/api/client';
import { streamEvents } from '@/api/sse';
export const useJobsStore = defineStore('jobs', () => {
    const items = ref([]);
    const loading = ref(false);
    const error = ref(null);
    let sse = null;
    const byId = computed(() => Object.fromEntries(items.value.map((j) => [j.id, j])));
    async function load() {
        loading.value = true;
        error.value = null;
        try {
            items.value = await api.jobs.list({ limit: 200 });
        }
        catch (e) {
            error.value = e instanceof Error ? e.message : String(e);
        }
        finally {
            loading.value = false;
        }
    }
    function startFirehose() {
        if (sse)
            return;
        sse = streamEvents('/api/jobs/events', {
            status(ev) {
                const data = ev.data;
                const j = items.value.find((x) => x.id === data.jobId);
                if (j) {
                    j.status = data.status;
                    if (data.error)
                        j.error = data.error;
                }
                else {
                    // new job appeared elsewhere; refresh list
                    load();
                }
            },
            progress(ev) {
                // firehose carries progress for all jobs; cheap to update
                const job = items.value.find((x) => ev.topic === `job:${x.id}`);
                if (job && ev.data && typeof ev.data === 'object') {
                    job.progress = ev.data;
                }
            }
        });
    }
    function stopFirehose() {
        sse?.close();
        sse = null;
    }
    async function enqueue(body) {
        const j = await api.jobs.create(body);
        items.value = [j, ...items.value];
        return j;
    }
    async function cancel(id, opts = {}) {
        await api.jobs.cancel(id, opts);
    }
    onScopeDispose(stopFirehose);
    return { items, loading, error, byId, load, enqueue, cancel, startFirehose, stopFirehose };
});
