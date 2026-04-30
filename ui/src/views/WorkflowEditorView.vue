<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { usePresetsStore } from '@/stores/presets'
import { useWorkflowsStore } from '@/stores/workflows'
import Button from '@/components/common/Button.vue'

const route = useRoute()
const router = useRouter()
const presets = usePresetsStore()
const workflows = useWorkflowsStore()
const { items: presetItems } = storeToRefs(presets)

const id = computed(() => (route.params.id as string) || '')
const isNew = computed(() => !id.value || id.value === 'new')

const form = ref({
	name: '',
	enabled: true,
	trigger: {
		type: 'file_created' as const,
		watchDir: '/media/incoming',
		glob: '*.mkv'
	},
	action: {
		type: 'enqueue_transcode' as const,
		presetId: '',
		outputPath: '/media/output/{{stem}}.mp4',
		sourceMovePath: '/media/dump/',
		cachePath: '/media/.tidal/'
	},
	pollIntervalMs: 30000,
	stableThresholdMs: 60000
})

const submitting = ref(false)
const submitErr = ref<string | null>(null)

onMounted(async () => {
	await presets.load()
	if (isNew.value) {
		// Default preset to first available.
		if (presetItems.value.length && !form.value.action.presetId) {
			form.value.action.presetId = presetItems.value[0]!.id
		}
		return
	}
	// Load existing workflow.
	const w = await (await import('@/api/client')).api.workflows.get(id.value)
	form.value.name = w.name
	form.value.enabled = w.enabled
	form.value.trigger.watchDir = w.trigger.watchDir ?? ''
	form.value.trigger.glob = w.trigger.glob ?? ''
	const a = w.actions[0]
	if (a) {
		form.value.action.presetId = a.presetId ?? ''
		form.value.action.outputPath = a.outputPath ?? ''
		form.value.action.sourceMovePath = a.sourceMovePath ?? ''
		form.value.action.cachePath = a.cachePath ?? ''
	}
	form.value.pollIntervalMs = w.pollIntervalMs
	form.value.stableThresholdMs = w.stableThresholdMs
})

async function submit() {
	submitErr.value = null
	if (!form.value.name.trim()) {
		submitErr.value = 'name required'
		return
	}
	submitting.value = true
	try {
		const body = {
			name: form.value.name.trim(),
			enabled: form.value.enabled,
			trigger: {
				type: form.value.trigger.type,
				watchDir: form.value.trigger.watchDir.trim(),
				glob: form.value.trigger.glob.trim() || '*'
			},
			actions: [{
				type: form.value.action.type,
				presetId: form.value.action.presetId,
				outputPath: form.value.action.outputPath.trim(),
				sourceMovePath: form.value.action.sourceMovePath.trim(),
				cachePath: form.value.action.cachePath.trim()
			}],
			pollIntervalMs: form.value.pollIntervalMs,
			stableThresholdMs: form.value.stableThresholdMs
		}
		if (isNew.value) await workflows.create(body)
		else await workflows.update(id.value, body)
		router.push('/workflows')
	} catch (e) {
		submitErr.value = e instanceof Error ? e.message : String(e)
	} finally {
		submitting.value = false
	}
}

const TEMPLATE_VARS = ['{{path}}', '{{dir}}', '{{filename}}', '{{stem}}', '{{ext}}', '{{date}}']
</script>

<template>
	<div class="space-y-6 max-w-3xl">
		<h2 class="text-lg font-semibold">{{ isNew ? 'New workflow' : 'Edit workflow' }}</h2>

		<form class="space-y-6" @submit.prevent="submit">
			<!-- General -->
			<section class="card bg-base-100 border border-base-300">
				<div class="card-body">
					<h3 class="card-title text-base">General</h3>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Name</legend>
						<input v-model="form.name" required class="input input-bordered w-full" />
					</fieldset>
					<label class="label cursor-pointer justify-start gap-3">
						<input v-model="form.enabled" type="checkbox" class="toggle toggle-primary" />
						<span class="label-text">Enabled</span>
					</label>
				</div>
			</section>

			<!-- Trigger -->
			<section class="card bg-base-100 border border-base-300">
				<div class="card-body">
					<h3 class="card-title text-base">Trigger</h3>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Type</legend>
						<select v-model="form.trigger.type" class="select select-bordered w-full" disabled>
							<option value="file_created">file_created</option>
						</select>
					</fieldset>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Watch directory</legend>
						<input v-model="form.trigger.watchDir" placeholder="/media/incoming" class="input input-bordered w-full font-mono text-sm" />
					</fieldset>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Glob</legend>
						<input v-model="form.trigger.glob" placeholder="*.mkv" class="input input-bordered w-full font-mono text-sm" />
					</fieldset>
				</div>
			</section>

			<!-- Action -->
			<section class="card bg-base-100 border border-base-300">
				<div class="card-body">
					<h3 class="card-title text-base">Action</h3>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Type</legend>
						<select v-model="form.action.type" class="select select-bordered w-full" disabled>
							<option value="enqueue_transcode">enqueue_transcode</option>
						</select>
					</fieldset>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Preset</legend>
						<select v-model="form.action.presetId" class="select select-bordered w-full">
							<option v-for="p in presetItems" :key="p.id" :value="p.id">{{ p.name }} — {{ p.description }}</option>
						</select>
					</fieldset>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Output path (templated)</legend>
						<input v-model="form.action.outputPath" placeholder="/media/output/{{stem}}.mp4" class="input input-bordered w-full font-mono text-sm" />
					</fieldset>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Source move path</legend>
						<input v-model="form.action.sourceMovePath" placeholder="/media/dump/" class="input input-bordered w-full font-mono text-sm" />
					</fieldset>
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Cache path</legend>
						<input v-model="form.action.cachePath" placeholder="/media/.tidal/" class="input input-bordered w-full font-mono text-sm" />
					</fieldset>
					<div class="text-xs text-base-content/60">
						Variables:
						<code v-for="v in TEMPLATE_VARS" :key="v" class="font-mono mr-2">{{ v }}</code>
					</div>
				</div>
			</section>

			<!-- Schedule -->
			<section class="card bg-base-100 border border-base-300">
				<div class="card-body">
					<h3 class="card-title text-base">Schedule</h3>
					<div class="grid grid-cols-2 gap-4">
						<fieldset class="fieldset">
							<legend class="fieldset-legend">Poll interval (ms)</legend>
							<input v-model.number="form.pollIntervalMs" type="number" class="input input-bordered w-full" />
						</fieldset>
						<fieldset class="fieldset">
							<legend class="fieldset-legend">Stable threshold (ms)</legend>
							<input v-model.number="form.stableThresholdMs" type="number" class="input input-bordered w-full" />
							<p class="text-xs text-base-content/50">Don't fire while file is still being written.</p>
						</fieldset>
					</div>
				</div>
			</section>

			<div v-if="submitErr" class="alert alert-error"><span>{{ submitErr }}</span></div>

			<div class="flex justify-end gap-2">
				<Button variant="ghost" type="button" @click="router.push('/workflows')">Cancel</Button>
				<Button type="submit" :loading="submitting">{{ isNew ? 'Create' : 'Save' }}</Button>
			</div>
		</form>
	</div>
</template>
