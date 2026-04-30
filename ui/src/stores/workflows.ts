import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'
import type { Workflow, WorkflowAction, WorkflowTrigger } from '@/api/types'

export const useWorkflowsStore = defineStore('workflows', () => {
	const items = ref<Workflow[]>([])
	const loading = ref(false)
	const error = ref<string | null>(null)

	async function load() {
		loading.value = true
		error.value = null
		try {
			items.value = await api.workflows.list()
		} catch (e) {
			error.value = e instanceof Error ? e.message : String(e)
		} finally {
			loading.value = false
		}
	}

	async function create(body: {
		name: string
		enabled: boolean
		trigger: WorkflowTrigger
		actions: WorkflowAction[]
		pollIntervalMs?: number
		stableThresholdMs?: number
	}) {
		const w = await api.workflows.create(body)
		items.value = [...items.value, w].sort((a, b) => a.name.localeCompare(b.name))
		return w
	}

	async function update(id: string, body: Parameters<typeof api.workflows.update>[1]) {
		const w = await api.workflows.update(id, body)
		items.value = items.value.map((x) => (x.id === id ? w : x))
		return w
	}

	async function remove(id: string) {
		await api.workflows.remove(id)
		items.value = items.value.filter((x) => x.id !== id)
	}

	async function setEnabled(id: string, enabled: boolean) {
		if (enabled) await api.workflows.enable(id)
		else await api.workflows.disable(id)
		// Refresh just this row.
		const w = await api.workflows.get(id)
		items.value = items.value.map((x) => (x.id === id ? w : x))
	}

	return { items, loading, error, load, create, update, remove, setEnabled }
})
