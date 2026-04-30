<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useJobsStore } from '@/stores/jobs'
import { usePresetsStore } from '@/stores/presets'
import StatusBadge from '@/components/jobs/StatusBadge.vue'
import ProgressBar from '@/components/jobs/ProgressBar.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Button from '@/components/common/Button.vue'
import Modal from '@/components/common/Modal.vue'
import type { Job, JobStatus } from '@/api/types'

const TERMINAL: JobStatus[] = ['succeeded', 'failed', 'cancelled']
function isTerminal(s: JobStatus) { return TERMINAL.includes(s) }

const jobs = useJobsStore()
const presets = usePresetsStore()
const { items, loading, error } = storeToRefs(jobs)

const DEFAULT_CACHE = '/var/cache/tidal'

const showNew = ref(false)
const form = ref({
	presetId: '',
	sourcePath: '',
	outputPath: '',
	cachePath: DEFAULT_CACHE,
	sourceMovePath: ''
})
const submitting = ref(false)
const submitErr = ref<string | null>(null)

const cancelTarget = ref<Job | null>(null)
const cancelForce = ref(false)
const cancelling = ref(false)
const cancelErr = ref<string | null>(null)

onMounted(async () => {
	await Promise.all([jobs.load(), presets.load()])
	jobs.startFirehose()
	if (presets.items.length && !form.value.presetId) form.value.presetId = presets.items[0]!.id
})

const sortedJobs = computed(() =>
	[...items.value].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt))
)

// Terminal jobs hard-delete after a confirm; non-terminal jobs open the cancel
// dialog, which exposes the optional force-cancel toggle.
async function rowAction(j: Job, ev: MouseEvent) {
	ev.stopPropagation()
	if (isTerminal(j.status)) {
		if (!confirm(`Delete this job?\n${j.sourcePath}`)) return
		try {
			await jobs.cancel(j.id)
			items.value = items.value.filter((x) => x.id !== j.id)
		} catch (e) {
			alert(e instanceof Error ? e.message : String(e))
		}
		return
	}
	cancelTarget.value = j
	cancelForce.value = false
	cancelErr.value = null
}

async function confirmCancel() {
	const j = cancelTarget.value
	if (!j) return
	cancelling.value = true
	cancelErr.value = null
	try {
		await jobs.cancel(j.id, { force: cancelForce.value })
		cancelTarget.value = null
	} catch (e) {
		cancelErr.value = e instanceof Error ? e.message : String(e)
	} finally {
		cancelling.value = false
	}
}

async function submit() {
	submitErr.value = null
	if (!form.value.presetId || !form.value.sourcePath) {
		submitErr.value = 'preset and source required'
		return
	}
	submitting.value = true
	try {
		await jobs.enqueue({
			presetId: form.value.presetId,
			sourcePath: form.value.sourcePath.trim(),
			outputPath: form.value.outputPath.trim() || undefined,
			cachePath: form.value.cachePath.trim() || undefined,
			sourceMovePath: form.value.sourceMovePath.trim() || undefined
		})
		showNew.value = false
		form.value.sourcePath = ''
		form.value.outputPath = ''
		form.value.sourceMovePath = ''
		form.value.cachePath = DEFAULT_CACHE
	} catch (e) {
		submitErr.value = e instanceof Error ? e.message : String(e)
	} finally {
		submitting.value = false
	}
}
</script>

<template>
	<div class="space-y-6">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<p class="max-w-prose text-sm text-base-content/60">
				Live view of every transcode in flight. Status updates stream over SSE.
			</p>
			<Button @click="showNew = true">+ New job</Button>
		</div>

		<div v-if="error" class="alert alert-error">
			<span>{{ error }}</span>
		</div>

		<div v-if="loading" class="flex items-center gap-2 text-sm text-base-content/60">
			<span class="loading loading-spinner loading-sm" /> Loading…
		</div>

		<EmptyState
			v-else-if="!sortedJobs.length"
			title="No jobs yet"
			hint="Enqueue your first transcode to see live progress, logs, and history."
		>
			<Button @click="showNew = true">+ New job</Button>
		</EmptyState>

		<div v-else class="overflow-x-auto rounded-box border border-base-300 bg-base-100">
			<table class="table table-zebra">
				<thead>
					<tr>
						<th>Status</th>
						<th>Source</th>
						<th>Output</th>
						<th class="w-72">Progress</th>
						<th>Created</th>
						<th class="w-12 text-right"></th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="j in sortedJobs"
						:key="j.id"
						class="cursor-pointer hover"
						@click="$router.push(`/jobs/${j.id}`)"
					>
						<td><StatusBadge :status="j.status" /></td>
						<td class="font-mono text-xs">{{ j.sourcePath }}</td>
						<td class="font-mono text-xs">{{ j.outputPath }}</td>
						<td><ProgressBar :percent="j.progress.percent" /></td>
						<td class="text-xs text-base-content/60 whitespace-nowrap">{{ new Date(j.createdAt).toLocaleString() }}</td>
						<td class="text-right">
							<button
								class="btn btn-ghost btn-xs btn-square"
								:class="isTerminal(j.status) ? 'text-error hover:bg-error/10' : 'text-warning hover:bg-warning/10'"
								:title="isTerminal(j.status) ? 'Delete job' : 'Cancel job'"
								@click="rowAction(j, $event)"
							>
								<!-- trash for terminal, stop-square for running -->
								<svg v-if="isTerminal(j.status)" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="size-4"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
								<svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="size-4"><path stroke-linecap="round" stroke-linejoin="round" d="M5.25 7.5A2.25 2.25 0 0 1 7.5 5.25h9a2.25 2.25 0 0 1 2.25 2.25v9a2.25 2.25 0 0 1-2.25 2.25h-9a2.25 2.25 0 0 1-2.25-2.25v-9Z" /></svg>
							</button>
						</td>
					</tr>
				</tbody>
			</table>
		</div>

		<Modal :open="!!cancelTarget" title="Cancel job" @close="cancelTarget = null">
			<div v-if="cancelTarget" class="space-y-4">
				<p class="text-sm text-base-content/70">
					{{ cancelForce
						? 'Force cancel marks the job cancelled immediately. The dispatcher will clean up the K8s Job on its next watch tick. Use only for stuck/orphaned jobs.'
						: 'Signals the dispatcher to stop the running ffmpeg and mark the job cancelled.' }}
				</p>
				<div class="rounded-box bg-base-200 px-3 py-2 font-mono text-xs break-all">
					{{ cancelTarget.sourcePath }}
				</div>
				<label class="label cursor-pointer justify-start gap-3">
					<input v-model="cancelForce" type="checkbox" class="toggle toggle-warning" />
					<span class="label-text">Force cancel <span class="text-base-content/50">(skip dispatcher coordination)</span></span>
				</label>
				<div v-if="cancelErr" class="alert alert-error">
					<span>{{ cancelErr }}</span>
				</div>
				<div class="flex justify-end gap-2">
					<Button variant="ghost" @click="cancelTarget = null">Keep running</Button>
					<Button
						:variant="cancelForce ? 'danger' : 'primary'"
						:loading="cancelling"
						@click="confirmCancel"
					>{{ cancelForce ? 'Force cancel' : 'Cancel job' }}</Button>
				</div>
			</div>
		</Modal>

		<Modal :open="showNew" title="New transcode job" @close="showNew = false">
			<form class="space-y-4" @submit.prevent="submit">
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Preset</legend>
					<select v-model="form.presetId" class="select select-bordered w-full">
						<option v-for="p in presets.items" :key="p.id" :value="p.id">
							{{ p.name }} — {{ p.description }}
						</option>
					</select>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Source path</legend>
					<input
						v-model="form.sourcePath"
						placeholder="/media/incoming/clip.mkv"
						class="input input-bordered w-full font-mono text-sm"
					/>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Output path <span class="text-base-content/50">(optional)</span></legend>
					<input
						v-model="form.outputPath"
						placeholder="leave empty for auto-derived"
						class="input input-bordered w-full font-mono text-sm"
					/>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Tidal cache path</legend>
					<input
						v-model="form.cachePath"
						placeholder="/var/cache/tidal"
						class="input input-bordered w-full font-mono text-sm"
					/>
					<p class="text-xs text-base-content/50">Working dir for ffmpeg temp files inside the worker container.</p>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Source move path <span class="text-base-content/50">(optional)</span></legend>
					<input
						v-model="form.sourceMovePath"
						placeholder="/media/archive/  (or full path)"
						class="input input-bordered w-full font-mono text-sm"
					/>
					<p class="text-xs text-base-content/50">If set, source file moves here on success.</p>
				</fieldset>
				<div v-if="submitErr" class="alert alert-error">
					<span>{{ submitErr }}</span>
				</div>
				<div class="flex justify-end gap-2">
					<Button variant="ghost" type="button" @click="showNew = false">Cancel</Button>
					<Button type="submit" :loading="submitting">Enqueue</Button>
				</div>
			</form>
		</Modal>
	</div>
</template>
