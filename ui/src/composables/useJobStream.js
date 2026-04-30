import { ref, onBeforeUnmount } from 'vue';
import { api } from '@/api/client';
import { streamEvents } from '@/api/sse';
const MAX_LOG_LINES = 1000;
export function useJobStream(id) {
    const job = ref(null);
    const logs = ref([]);
    const error = ref(null);
    const loaded = ref(false);
    let nextSeq = 0;
    api.jobs.get(id)
        .then(async (j) => {
        job.value = j;
        const initialLogs = await api.jobs.logs(id, 0, MAX_LOG_LINES);
        logs.value = initialLogs;
        if (initialLogs.length)
            nextSeq = initialLogs[initialLogs.length - 1].seq;
        loaded.value = true;
    })
        .catch((e) => {
        error.value = e instanceof Error ? e.message : String(e);
        loaded.value = true;
    });
    const sse = streamEvents(`/api/jobs/${id}/events`, {
        status(ev) {
            const d = ev.data;
            if (job.value) {
                job.value.status = d.status;
                if (d.error)
                    job.value.error = d.error;
            }
        },
        progress(ev) {
            if (job.value)
                job.value.progress = ev.data;
        },
        log(ev) {
            const log = ev.data;
            // Server-side seq is set on insert; SSE event currently carries 0.
            // Maintain a monotonic client seq so list keys stay stable.
            nextSeq += 1;
            logs.value.push({ ...log, seq: nextSeq });
            if (logs.value.length > MAX_LOG_LINES) {
                logs.value.splice(0, logs.value.length - MAX_LOG_LINES);
            }
        }
    });
    onBeforeUnmount(() => sse.close());
    return { job, logs, error, loaded };
}
