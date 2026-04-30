<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useJobStream } from '@/composables/useJobStream'
import { usePresetsStore } from '@/stores/presets'
import StatusBadge from '@/components/jobs/StatusBadge.vue'
import ProgressBar from '@/components/jobs/ProgressBar.vue'
import Button from '@/components/common/Button.vue'
import { api } from '@/api/client'

const route = useRoute()
const id = String(route.params.id)
const presets = usePresetsStore()
const { job, logs, error, loaded } = useJobStream(id)

if (!presets.items.length) presets.load()

const presetName = computed(() => {
	const p = job.value ? presets.byId[job.value.presetId] : undefined
	return p?.name ?? job.value?.presetId ?? '—'
})

const cancelling = ref(false)
async function cancelJob() {
	if (!job.value) return
	cancelling.value = true
	try {
		await api.jobs.cancel(job.value.id)
	} finally {
		cancelling.value = false
	}
}

const logPane = ref<HTMLDivElement | null>(null)
watch(
	() => logs.value.length,
	() => {
		nextTick(() => {
			if (logPane.value) logPane.value.scrollTop = logPane.value.scrollHeight
		})
	}
)
</script>

<template>
	<div v-if="!loaded" class="flex items-center gap-2 text-sm text-base-content/60">
		<span class="loading loading-spinner loading-sm" /> Loading…
	</div>
	<div v-else-if="error || !job" class="alert alert-error">
		<span>{{ error ?? 'Job not found' }}</span>
	</div>
	<div v-else class="space-y-6">
		<div class="flex flex-wrap items-center gap-3">
			<StatusBadge :status="job.status" />
			<span class="text-xs text-base-content/60">preset</span>
			<span class="font-mono text-xs">{{ presetName }}</span>
			<div class="flex-1" />
			<Button
				v-if="!['succeeded','failed','cancelled'].includes(job.status)"
				variant="danger"
				:loading="cancelling"
				@click="cancelJob"
			>Cancel</Button>
		</div>

		<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
			<div class="card bg-base-100 border border-base-300">
				<div class="card-body">
					<dl class="grid grid-cols-[110px_1fr] gap-x-4 gap-y-2 text-sm">
						<dt class="text-base-content/60">Source</dt>
						<dd class="break-all font-mono">{{ job.sourcePath }}</dd>
						<dt class="text-base-content/60">Output</dt>
						<dd class="break-all font-mono">{{ job.outputPath }}</dd>
						<dt class="text-base-content/60">Created</dt>
						<dd>{{ new Date(job.createdAt).toLocaleString() }}</dd>
						<dt class="text-base-content/60">Started</dt>
						<dd>{{ job.startedAt ? new Date(job.startedAt).toLocaleString() : '—' }}</dd>
						<dt class="text-base-content/60">Finished</dt>
						<dd>{{ job.finishedAt ? new Date(job.finishedAt).toLocaleString() : '—' }}</dd>
						<dt v-if="job.error" class="text-error">Error</dt>
						<dd v-if="job.error" class="text-error">{{ job.error }}</dd>
					</dl>
				</div>
			</div>

			<div class="card bg-base-100 border border-base-300">
				<div class="card-body">
					<ProgressBar :percent="job.progress.percent" label="Encode progress" />
					<div class="stats stats-horizontal mt-4 bg-base-200">
						<div class="stat py-2">
							<div class="stat-title text-xs">frame</div>
							<div class="stat-value text-base font-mono text-primary">{{ job.progress.frame }}</div>
						</div>
						<div class="stat py-2">
							<div class="stat-title text-xs">fps</div>
							<div class="stat-value text-base font-mono text-primary">{{ job.progress.fps.toFixed(1) }}</div>
						</div>
						<div class="stat py-2">
							<div class="stat-title text-xs">speed</div>
							<div class="stat-value text-base font-mono text-primary">{{ job.progress.speed.toFixed(2) }}x</div>
						</div>
					</div>
				</div>
			</div>
		</div>

		<div class="card bg-base-100 border border-base-300">
			<div class="px-4 py-2 border-b border-base-300 flex items-center justify-between">
				<span class="text-xs uppercase tracking-wide text-base-content/60">Logs ({{ logs.length }})</span>
			</div>
			<div ref="logPane" class="max-h-[480px] overflow-auto px-4 py-3 font-mono text-[12px] leading-relaxed bg-base-200/40">
				<div
					v-for="l in logs"
					:key="`${l.seq}-${l.emittedAt}`"
					:class="['whitespace-pre-wrap', l.stream === 'system' ? 'text-info' : l.stream === 'stderr' ? 'text-base-content' : 'text-base-content/80']"
				>
					{{ l.line }}
				</div>
				<div v-if="!logs.length" class="text-base-content/40">no logs yet</div>
			</div>
		</div>
	</div>
</template>
