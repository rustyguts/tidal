<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useSystemStore } from '@/stores/system'
import { api } from '@/api/client'
import Button from '@/components/common/Button.vue'

const sys = useSystemStore()
const { info } = storeToRefs(sys)

const concurrency = ref<number | null>(null)
const concurrencyDraft = ref<number>(4)
const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)
const saved = ref(false)

async function loadSettings() {
	loading.value = true
	error.value = null
	try {
		const s = await api.settings.get()
		concurrency.value = s.transcodeConcurrency
		concurrencyDraft.value = s.transcodeConcurrency
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		loading.value = false
	}
}

async function save() {
	saving.value = true
	error.value = null
	saved.value = false
	try {
		const n = Number(concurrencyDraft.value)
		if (!Number.isFinite(n) || n < 1 || n > 64) {
			throw new Error('Concurrency must be between 1 and 64')
		}
		const s = await api.settings.update({ transcodeConcurrency: n })
		concurrency.value = s.transcodeConcurrency
		concurrencyDraft.value = s.transcodeConcurrency
		saved.value = true
		setTimeout(() => (saved.value = false), 2500)
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		saving.value = false
	}
}

onMounted(loadSettings)
</script>

<template>
	<div class="space-y-6">
		<div class="card bg-base-100 border border-base-300">
			<div class="card-body space-y-4">
				<h2 class="card-title">Transcode</h2>
				<p class="text-sm text-base-content/60">
					Maximum number of ffmpeg jobs running concurrently across this worker.
					Changes take effect within a few seconds; in-flight jobs are not interrupted.
				</p>
				<div class="flex items-end gap-3">
					<fieldset class="fieldset w-32">
						<legend class="fieldset-legend">Concurrency</legend>
						<input
							v-model.number="concurrencyDraft"
							type="number"
							min="1"
							max="64"
							class="input input-bordered w-full"
							:disabled="loading"
						/>
					</fieldset>
					<Button
						:loading="saving"
						:disabled="loading || concurrencyDraft === concurrency"
						@click="save"
					>Save</Button>
					<span v-if="saved" class="text-success text-sm self-center">Saved</span>
				</div>
				<p v-if="concurrency !== null" class="text-xs text-base-content/50">
					Current cap: <span class="font-mono">{{ concurrency }}</span>
				</p>
				<div v-if="error" class="alert alert-error">
					<span>{{ error }}</span>
				</div>
			</div>
		</div>

		<div class="card bg-base-100 border border-base-300">
			<div class="card-body">
				<h2 class="card-title">System</h2>
				<table v-if="info" class="table table-sm w-full">
					<tbody>
						<tr><th class="w-40 text-base-content/60">Version</th><td class="font-mono">{{ info.version.version }}</td></tr>
						<tr><th class="text-base-content/60">Commit</th><td class="font-mono">{{ info.version.commit }}</td></tr>
						<tr><th class="text-base-content/60">Built</th><td class="font-mono">{{ info.version.date }}</td></tr>
						<tr><th class="text-base-content/60">Environment</th><td>{{ info.env }}</td></tr>
						<tr><th class="text-base-content/60">Dispatcher</th><td>{{ info.dispatcher }}</td></tr>
						<tr><th class="text-base-content/60">Media roots</th><td class="font-mono">{{ info.mediaRoots.join(', ') }}</td></tr>
					</tbody>
				</table>
				<p v-else class="flex items-center gap-2 text-sm text-base-content/60">
					<span class="loading loading-spinner loading-sm" /> Loading…
				</p>
			</div>
		</div>
	</div>
</template>
