<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useWorkflowsStore } from '@/stores/workflows'
import EmptyState from '@/components/common/EmptyState.vue'
import Button from '@/components/common/Button.vue'
import { useRouter } from 'vue-router'

const store = useWorkflowsStore()
const { items, loading, error } = storeToRefs(store)
const router = useRouter()

let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
	store.load()
	// Auto-refresh stats every 5s.
	timer = setInterval(() => store.load(), 5000)
})
onBeforeUnmount(() => {
	if (timer) clearInterval(timer)
})

function successRate(w: { runsCount: number; successCount: number }) {
	if (!w.runsCount) return 0
	return Math.round((w.successCount / w.runsCount) * 100)
}

const removing = ref<string | null>(null)
async function remove(id: string, name: string) {
	if (!confirm(`Delete workflow "${name}"?`)) return
	removing.value = id
	try {
		await store.remove(id)
	} finally {
		removing.value = null
	}
}

async function toggle(id: string, current: boolean) {
	await store.setEnabled(id, !current)
}
</script>

<template>
	<div class="space-y-6">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<p class="max-w-prose text-sm text-base-content/60">
				Workflows define a trigger and the actions that follow. v1 ships
				<code class="font-mono">file_created</code> +
				<code class="font-mono">enqueue_transcode</code>; more in v2.
			</p>
			<Button @click="router.push('/workflows/new')">+ New workflow</Button>
		</div>

		<div v-if="error" class="alert alert-error"><span>{{ error }}</span></div>

		<div v-if="loading && !items.length" class="flex items-center gap-2 text-sm text-base-content/60">
			<span class="loading loading-spinner loading-sm" /> Loading…
		</div>

		<EmptyState
			v-else-if="!items.length"
			title="No workflows yet"
			hint="Define a trigger + action and Tidal will fire jobs as files arrive."
		>
			<Button @click="router.push('/workflows/new')">+ New workflow</Button>
		</EmptyState>

		<div v-else class="overflow-x-auto rounded-box border border-base-300 bg-base-100">
			<table class="table table-zebra">
				<thead>
					<tr>
						<th>Name</th>
						<th>Trigger</th>
						<th>Action</th>
						<th class="w-44">Stats</th>
						<th class="w-24">Enabled</th>
						<th class="w-20"></th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="w in items" :key="w.id" class="hover">
						<td class="font-medium">
							<router-link :to="`/workflows/${w.id}`" class="link link-hover">
								{{ w.name }}
							</router-link>
						</td>
						<td class="text-xs">
							<div class="font-mono">{{ w.trigger.type }}</div>
							<div class="text-base-content/60">{{ w.trigger.watchDir }} / {{ w.trigger.glob }}</div>
						</td>
						<td class="text-xs">
							<div class="font-mono">{{ w.actions[0]?.type }}</div>
							<div class="text-base-content/60 truncate max-w-xs">{{ w.actions[0]?.outputPath || '—' }}</div>
						</td>
						<td>
							<div class="flex items-center gap-2 text-xs">
								<div class="radial-progress text-success" :style="{ '--value': successRate(w), '--size': '2.5rem', '--thickness': '3px' }" role="progressbar">
									{{ successRate(w) }}%
								</div>
								<div>
									<div><span class="font-mono">{{ w.successCount }}</span> ok</div>
									<div class="text-base-content/60"><span class="font-mono">{{ w.runsCount }}</span> total</div>
								</div>
							</div>
						</td>
						<td>
							<input
								type="checkbox"
								class="toggle toggle-sm toggle-primary"
								:checked="w.enabled"
								@change="toggle(w.id, w.enabled)"
							/>
						</td>
						<td class="text-right">
							<button
								class="btn btn-ghost btn-xs btn-square text-error"
								:disabled="removing === w.id"
								title="Delete workflow"
								@click="remove(w.id, w.name)"
							>
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="size-4"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
							</button>
						</td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>
</template>
