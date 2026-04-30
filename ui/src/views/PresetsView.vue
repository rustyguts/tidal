<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { usePresetsStore } from '@/stores/presets'
import EmptyState from '@/components/common/EmptyState.vue'
import Button from '@/components/common/Button.vue'

const store = usePresetsStore()
const { items, loading, error } = storeToRefs(store)

onMounted(() => store.load())

const restoring = ref(false)
async function restoreDefaults() {
	restoring.value = true
	try {
		await store.restoreDefaults()
	} finally {
		restoring.value = false
	}
}

async function remove(id: string) {
	if (!confirm('Delete this preset?')) return
	await store.remove(id)
}

async function duplicate(id: string, currentName: string) {
	const next = prompt('New preset name', `${currentName}-copy`)
	if (!next) return
	await store.duplicate(id, next)
}
</script>

<template>
	<div class="space-y-6">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<p class="max-w-prose text-sm text-base-content/60">
				Presets define ffmpeg parameters (codec, CRF, container, scale). Built-in
				presets can be deleted; deletions persist across restarts.
			</p>
			<div class="flex gap-2">
				<Button variant="ghost" :loading="restoring" @click="restoreDefaults">Restore defaults</Button>
				<Button :disabled="true">+ New preset</Button>
			</div>
		</div>

		<div v-if="error" class="alert alert-error"><span>{{ error }}</span></div>

		<div v-if="loading" class="flex items-center gap-2 text-sm text-base-content/60">
			<span class="loading loading-spinner loading-sm" /> Loading…
		</div>
		<EmptyState v-else-if="!items.length" title="No presets" hint="Click Restore defaults to seed the built-in presets." />

		<div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
			<div
				v-for="p in items"
				:key="p.id"
				class="card bg-base-100 border border-base-300 hover:border-primary/40 transition-colors"
			>
				<div class="card-body">
					<div class="flex items-start justify-between gap-2">
						<h3 class="card-title">{{ p.name }}</h3>
						<span v-if="p.builtin" class="badge badge-primary badge-outline">built-in</span>
					</div>
					<p class="text-sm text-base-content/60">{{ p.description || '—' }}</p>
					<dl class="grid grid-cols-2 gap-x-3 gap-y-1 text-xs mt-2">
						<dt class="text-base-content/50">codec</dt><dd class="font-mono">{{ p.spec.videoCodec }}</dd>
						<dt class="text-base-content/50">crf</dt><dd class="font-mono">{{ p.spec.crf }}</dd>
						<dt class="text-base-content/50">container</dt><dd class="font-mono">{{ p.spec.container }}</dd>
						<dt class="text-base-content/50">res</dt>
						<dd class="font-mono">{{ p.spec.resolution ? `${p.spec.resolution.width}×${p.spec.resolution.height}` : '—' }}</dd>
					</dl>
					<div class="card-actions justify-end mt-2">
						<Button variant="ghost" size="xs" @click="duplicate(p.id, p.name)">Duplicate</Button>
						<Button variant="danger" size="xs" @click="remove(p.id)">Delete</Button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
