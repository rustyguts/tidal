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

onMounted(async () => {
	await Promise.all([jobs.load(), presets.load()])
	jobs.startFirehose()
	if (presets.items.length && !form.value.presetId) form.value.presetId = presets.items[0]!.id
})

const sortedJobs = computed(() =>
	[...items.value].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt))
)

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
					</tr>
				</tbody>
			</table>
		</div>

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
