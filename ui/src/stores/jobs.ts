import { defineStore } from 'pinia'
import { ref, computed, onScopeDispose } from 'vue'
import { api } from '@/api/client'
import { streamEvents, type SSEHandle } from '@/api/sse'
import type { Job, JobStatus } from '@/api/types'

export const useJobsStore = defineStore('jobs', () => {
	const items = ref<Job[]>([])
	const loading = ref(false)
	const error = ref<string | null>(null)
	let sse: SSEHandle | null = null

	const byId = computed(() => Object.fromEntries(items.value.map((j) => [j.id, j])) as Record<string, Job>)

	async function load() {
		loading.value = true
		error.value = null
		try {
			items.value = await api.jobs.list({ limit: 200 })
		} catch (e) {
			error.value = e instanceof Error ? e.message : String(e)
		} finally {
			loading.value = false
		}
	}

	function startFirehose() {
		if (sse) return
		sse = streamEvents('/api/jobs/events', {
			status(ev) {
				const data = ev.data as { jobId: string; status: JobStatus; error?: string }
				const j = items.value.find((x) => x.id === data.jobId)
				if (j) {
					j.status = data.status
					if (data.error) j.error = data.error
				} else {
					// new job appeared elsewhere; refresh list
					load()
				}
			},
			progress(ev) {
				// firehose carries progress for all jobs; cheap to update
				const job = items.value.find((x) => ev.topic === `job:${x.id}`)
				if (job && ev.data && typeof ev.data === 'object') {
					job.progress = ev.data as Job['progress']
				}
			}
		})
	}

	function stopFirehose() {
		sse?.close()
		sse = null
	}

	async function enqueue(body: {
		presetId: string
		sourcePath: string
		outputPath?: string
		cachePath?: string
		sourceMovePath?: string
	}) {
		const j = await api.jobs.create(body)
		items.value = [j, ...items.value]
		return j
	}

	async function cancel(id: string, opts: { force?: boolean } = {}) {
		await api.jobs.cancel(id, opts)
	}

	onScopeDispose(stopFirehose)
	return { items, loading, error, byId, load, enqueue, cancel, startFirehose, stopFirehose }
})
