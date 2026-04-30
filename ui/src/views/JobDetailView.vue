<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useJobStream } from '@/composables/useJobStream'
import { usePresetsStore } from '@/stores/presets'
import { useJobsStore } from '@/stores/jobs'
import StatusBadge from '@/components/jobs/StatusBadge.vue'
import ProgressBar from '@/components/jobs/ProgressBar.vue'
import Button from '@/components/common/Button.vue'
import Modal from '@/components/common/Modal.vue'
import { api } from '@/api/client'

const route = useRoute()
const router = useRouter()
const id = String(route.params.id)
const presets = usePresetsStore()
const jobsStore = useJobsStore()
const { job, logs, error, loaded } = useJobStream(id)

if (!presets.items.length) presets.load()

const presetName = computed(() => {
	const p = job.value ? presets.byId[job.value.presetId] : undefined
	return p?.name ?? job.value?.presetId ?? '—'
})

const showCancel = ref(false)
const cancelForce = ref(false)
const cancelling = ref(false)
const cancelErr = ref<string | null>(null)

function openCancel() {
	cancelForce.value = false
	cancelErr.value = null
	showCancel.value = true
}

async function confirmCancel() {
	if (!job.value) return
	cancelling.value = true
	cancelErr.value = null
	try {
		await api.jobs.cancel(job.value.id, { force: cancelForce.value })
		showCancel.value = false
	} catch (e) {
		cancelErr.value = e instanceof Error ? e.message : String(e)
	} finally {
		cancelling.value = false
	}
}

const showClone = ref(false)
const cloneForm = ref({
	presetId: '',
	sourcePath: '',
	outputPath: '',
	cachePath: '',
	sourceMovePath: ''
})
const cloning = ref(false)
const cloneErr = ref<string | null>(null)

function openClone() {
	if (!job.value) return
	cloneForm.value = {
		presetId: job.value.presetId,
		sourcePath: job.value.sourcePath,
		outputPath: job.value.outputPath,
		cachePath: job.value.cachePath ?? '',
		sourceMovePath: job.value.sourceMovePath ?? ''
	}
	cloneErr.value = null
	showClone.value = true
}

async function submitClone() {
	cloneErr.value = null
	if (!cloneForm.value.presetId || !cloneForm.value.sourcePath) {
		cloneErr.value = 'preset and source required'
		return
	}
	cloning.value = true
	try {
		const j = await jobsStore.enqueue({
			presetId: cloneForm.value.presetId,
			sourcePath: cloneForm.value.sourcePath.trim(),
			outputPath: cloneForm.value.outputPath.trim() || undefined,
			cachePath: cloneForm.value.cachePath.trim() || undefined,
			sourceMovePath: cloneForm.value.sourceMovePath.trim() || undefined
		})
		showClone.value = false
		router.push(`/jobs/${j.id}`)
	} catch (e) {
		cloneErr.value = e instanceof Error ? e.message : String(e)
	} finally {
		cloning.value = false
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
			<Button variant="ghost" @click="openClone">Clone job</Button>
			<Button
				v-if="!['succeeded','failed','cancelled'].includes(job.status)"
				variant="danger"
				@click="openCancel"
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

		<Modal :open="showCancel" title="Cancel job" @close="showCancel = false">
			<div class="space-y-4">
				<p class="text-sm text-base-content/70">
					{{ cancelForce
						? 'Force cancel marks the job cancelled immediately. The dispatcher will clean up the K8s Job on its next watch tick. Use only for stuck/orphaned jobs.'
						: 'Signals the dispatcher to stop the running ffmpeg and mark the job cancelled.' }}
				</p>
				<div class="rounded-box bg-base-200 px-3 py-2 font-mono text-xs break-all">
					{{ job.sourcePath }}
				</div>
				<label class="label cursor-pointer justify-start gap-3">
					<input v-model="cancelForce" type="checkbox" class="toggle toggle-warning" />
					<span class="label-text">Force cancel <span class="text-base-content/50">(skip dispatcher coordination)</span></span>
				</label>
				<div v-if="cancelErr" class="alert alert-error">
					<span>{{ cancelErr }}</span>
				</div>
				<div class="flex justify-end gap-2">
					<Button variant="ghost" @click="showCancel = false">Keep running</Button>
					<Button
						:variant="cancelForce ? 'danger' : 'primary'"
						:loading="cancelling"
						@click="confirmCancel"
					>{{ cancelForce ? 'Force cancel' : 'Cancel job' }}</Button>
				</div>
			</div>
		</Modal>

		<Modal :open="showClone" title="Clone job" @close="showClone = false">
			<form class="space-y-4" @submit.prevent="submitClone">
				<p class="text-sm text-base-content/60">
					Pre-filled from this job. Edit any field and submit to enqueue a new transcode.
				</p>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Preset</legend>
					<select v-model="cloneForm.presetId" class="select select-bordered w-full">
						<option v-for="p in presets.items" :key="p.id" :value="p.id">
							{{ p.name }}<template v-if="p.description"> — {{ p.description }}</template>
						</option>
					</select>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Source path</legend>
					<input
						v-model="cloneForm.sourcePath"
						class="input input-bordered w-full font-mono text-sm"
					/>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Output path <span class="text-base-content/50">(optional)</span></legend>
					<input
						v-model="cloneForm.outputPath"
						placeholder="leave empty for auto-derived"
						class="input input-bordered w-full font-mono text-sm"
					/>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Tidal cache path</legend>
					<input
						v-model="cloneForm.cachePath"
						placeholder="/var/cache/tidal"
						class="input input-bordered w-full font-mono text-sm"
					/>
				</fieldset>
				<fieldset class="fieldset">
					<legend class="fieldset-legend">Source move path <span class="text-base-content/50">(optional)</span></legend>
					<input
						v-model="cloneForm.sourceMovePath"
						placeholder="/media/archive/"
						class="input input-bordered w-full font-mono text-sm"
					/>
				</fieldset>
				<div v-if="cloneErr" class="alert alert-error">
					<span>{{ cloneErr }}</span>
				</div>
				<div class="flex justify-end gap-2">
					<Button variant="ghost" type="button" @click="showClone = false">Cancel</Button>
					<Button type="submit" :loading="cloning">Enqueue clone</Button>
				</div>
			</form>
		</Modal>
	</div>
</template>
