<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
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

function shortRate(rate?: { mode: string; crf?: number; qp?: number; bitrate?: string }): string {
	if (!rate) return '—'
	if (rate.mode === 'crf' && rate.crf != null) return `CRF ${rate.crf}`
	if (rate.mode === 'qp' && rate.qp != null) return `QP ${rate.qp}`
	if (rate.bitrate) return `${rate.mode.toUpperCase()} ${rate.bitrate}`
	return rate.mode.toUpperCase()
}

const sorted = computed(() =>
	[...items.value].sort((a, b) => {
		if (a.builtin !== b.builtin) return a.builtin ? -1 : 1
		return a.name.localeCompare(b.name)
	})
)
</script>

<template>
	<div class="space-y-6">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<p class="max-w-prose text-sm text-base-content/60">
				Presets are structured ffmpeg encode specs (codec, rate, container, filters,
				hardware accel) plus optional default output / cache / archive paths. Built-in
				presets can be deleted; deletions persist across restarts.
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
				v-for="p in sorted"
				:key="p.id"
				class="card bg-base-100 border border-base-300 hover:border-primary/40 transition-colors cursor-pointer"
				@click="router.push({ name: 'preset-edit', params: { id: p.id } })"
			>
				<div class="card-body gap-3">
					<div class="flex items-start justify-between gap-2">
						<div class="min-w-0">
							<h3 class="card-title truncate">{{ p.name }}</h3>
							<p class="text-xs text-base-content/60 line-clamp-2">{{ p.description || '—' }}</p>
						</div>
						<span v-if="p.builtin" class="badge badge-primary badge-outline shrink-0">built-in</span>
					</div>

					<!-- Headline encode summary -->
					<div class="rounded-box bg-base-200/60 px-3 py-2">
						<div class="font-mono text-sm">
							<span class="font-semibold">{{ p.spec?.video?.disabled ? 'audio-only' : (p.spec?.video?.codec || '—') }}</span>
							<span v-if="!p.spec?.video?.disabled" class="text-base-content/60"> · {{ shortRate(p.spec?.video?.rate) }}</span>
						</div>
						<div class="flex flex-wrap gap-1.5 mt-1.5">
							<span class="badge badge-sm badge-ghost font-mono">{{ p.spec?.container?.format || '—' }}</span>
							<span v-if="p.spec?.video?.resolution" class="badge badge-sm badge-ghost font-mono">
								{{ p.spec.video.resolution.width }}×{{ p.spec.video.resolution.height }}
							</span>
							<span v-if="p.spec?.video?.preset" class="badge badge-sm badge-ghost font-mono">{{ p.spec.video.preset }}</span>
							<span v-if="p.spec?.video?.twoPass" class="badge badge-sm badge-warning">2-pass</span>
							<span v-if="p.spec?.hwaccel?.type && p.spec.hwaccel.type !== 'none'" class="badge badge-sm badge-info font-mono">
								{{ p.spec.hwaccel.type }}
							</span>
						</div>
					</div>

					<!-- Audio summary -->
					<div v-if="!p.spec?.audio?.disabled" class="text-xs">
						<span class="text-base-content/50">audio</span>
						<span class="font-mono ml-2">{{ p.spec?.audio?.codec || '—' }}<span v-if="p.spec?.audio?.bitrate"> @ {{ p.spec.audio.bitrate }}</span></span>
					</div>

					<!-- Paths (new) -->
					<div v-if="p.outputPath || p.cachePath || p.sourceMovePath" class="text-xs space-y-0.5 border-t border-base-300 pt-2">
						<div v-if="p.outputPath"><span class="text-base-content/50">→ output</span> <span class="font-mono ml-1">{{ p.outputPath }}</span></div>
						<div v-if="p.sourceMovePath"><span class="text-base-content/50">⇒ archive</span> <span class="font-mono ml-1">{{ p.sourceMovePath }}</span></div>
						<div v-if="p.cachePath"><span class="text-base-content/50">cache</span> <span class="font-mono ml-1">{{ p.cachePath }}</span></div>
					</div>

					<div class="card-actions justify-end" @click.stop>
						<Button variant="ghost" size="xs" @click="duplicate(p.id, p.name)">Duplicate</Button>
						<Button variant="danger" size="xs" @click="remove(p.id)">Delete</Button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
