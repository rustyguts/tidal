<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { usePresetsStore } from '@/stores/presets'
import EmptyState from '@/components/common/EmptyState.vue'
import Button from '@/components/common/Button.vue'

const store = usePresetsStore()
const router = useRouter()
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
	const created = await store.duplicate(id, next)
	router.push({ name: 'preset-edit', params: { id: created.id } })
}

function shortRate(rate: { mode: string; crf?: number; qp?: number; bitrate?: string }): string {
	if (rate.mode === 'crf' && rate.crf != null) return `crf ${rate.crf}`
	if (rate.mode === 'qp' && rate.qp != null) return `qp ${rate.qp}`
	if (rate.bitrate) return `${rate.mode} ${rate.bitrate}`
	return rate.mode
}
</script>

<template>
	<div class="space-y-6">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<p class="max-w-prose text-sm text-base-content/60">
				Presets are structured ffmpeg encode specs (codec, rate, container, filters,
				hardware accel). Built-in presets can be deleted; deletions persist across
				restarts.
			</p>
			<div class="flex gap-2">
				<Button variant="ghost" :loading="restoring" @click="restoreDefaults">Restore defaults</Button>
				<Button @click="router.push({ name: 'preset-new' })">+ New preset</Button>
			</div>
		</div>

		<div v-if="error" class="alert alert-error"><span>{{ error }}</span></div>

		<div v-if="loading" class="flex items-center gap-2 text-sm text-base-content/60">
			<span class="loading loading-spinner loading-sm" /> Loading…
		</div>
		<EmptyState v-else-if="!items.length" title="No presets" hint="Click Restore defaults to seed the built-in presets, or + New preset to create one." />

		<div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
			<div
				v-for="p in items"
				:key="p.id"
				class="card bg-base-100 border border-base-300 hover:border-primary/40 transition-colors cursor-pointer"
				@click="router.push({ name: 'preset-edit', params: { id: p.id } })"
			>
				<div class="card-body">
					<div class="flex items-start justify-between gap-2">
						<h3 class="card-title">{{ p.name }}</h3>
						<span v-if="p.builtin" class="badge badge-primary badge-outline">built-in</span>
					</div>
					<p class="text-sm text-base-content/60">{{ p.description || '—' }}</p>
					<dl class="grid grid-cols-2 gap-x-3 gap-y-1 text-xs mt-2">
						<dt class="text-base-content/50">codec</dt><dd class="font-mono">{{ p.spec.video.codec || (p.spec.video.disabled ? '(none)' : '—') }}</dd>
						<dt class="text-base-content/50">rate</dt><dd class="font-mono">{{ shortRate(p.spec.video.rate) }}</dd>
						<dt class="text-base-content/50">container</dt><dd class="font-mono">{{ p.spec.container.format }}</dd>
						<dt class="text-base-content/50">res</dt>
						<dd class="font-mono">{{ p.spec.video.resolution ? `${p.spec.video.resolution.width}×${p.spec.video.resolution.height}` : '—' }}</dd>
						<dt v-if="p.spec.video.twoPass" class="text-base-content/50">passes</dt>
						<dd v-if="p.spec.video.twoPass" class="font-mono">2-pass</dd>
						<dt v-if="p.spec.hwaccel?.type && p.spec.hwaccel.type !== 'none'" class="text-base-content/50">hwaccel</dt>
						<dd v-if="p.spec.hwaccel?.type && p.spec.hwaccel.type !== 'none'" class="font-mono">{{ p.spec.hwaccel.type }}</dd>
					</dl>
					<div class="card-actions justify-end mt-2" @click.stop>
						<Button variant="ghost" size="xs" @click="duplicate(p.id, p.name)">Duplicate</Button>
						<Button variant="danger" size="xs" @click="remove(p.id)">Delete</Button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
