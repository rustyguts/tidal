<script setup lang="ts">
import { ref, computed, onMounted, watch, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'

import { api } from '@/api/client'
import type {
	PresetSpecV2,
	FilterStep,
	PreviewResponse,
	CatalogFilter
} from '@/api/types'
import { usePresetCatalogStore } from '@/stores/presetCatalog'
import { usePresetsStore } from '@/stores/presets'
import { blankSpec, debounce } from '@/composables/presetDraft'

import Button from '@/components/common/Button.vue'

const route = useRoute()
const router = useRouter()
const catalogStore = usePresetCatalogStore()
const presetsStore = usePresetsStore()
const { catalog } = storeToRefs(catalogStore)

const isNew = computed(() => route.params.id === 'new' || !route.params.id)

const name = ref('')
const description = ref('')
const builtin = ref(false)
const draft = reactive<PresetSpecV2>(blankSpec())
const activeTab = ref<'container' | 'video' | 'audio' | 'filters' | 'hwaccel' | 'advanced'>('video')

const previewArgv = ref<string[]>([])
const previewErrors = ref<string[]>([])
const previewLoading = ref(false)
const saveError = ref<string | null>(null)
const saving = ref(false)

const codecCatalog = computed(() => catalogStore.videoCodec(draft.video.codec))
const audioCodecCatalog = computed(() =>
	draft.audio.codec ? catalogStore.audioCodec(draft.audio.codec) : undefined
)

const showVideoFields = computed(() => !draft.video.disabled && draft.video.codec !== 'copy')
const showRateFields = computed(() => showVideoFields.value && draft.video.rate.mode !== 'none')

const videoFilterCatalog = computed(() => catalogStore.filtersByScope('video'))
const audioFilterCatalog = computed(() => catalogStore.filtersByScope('audio'))

onMounted(async () => {
	await catalogStore.load()
	if (!isNew.value) {
		const id = String(route.params.id)
		try {
			const p = await api.presets.get(id)
			name.value = p.name
			description.value = p.description
			builtin.value = p.builtin
			// Merge against blank to fill missing nested keys when the server
			// returns a partial spec (e.g. legacy v1 jsonb that hasn't been
			// upgraded server-side).
			const base = blankSpec()
			const merged: PresetSpecV2 = {
				...base,
				...p.spec,
				container: { ...base.container, ...(p.spec?.container ?? {}) },
				video: { ...base.video, ...(p.spec?.video ?? {}) },
				audio: { ...base.audio, ...(p.spec?.audio ?? {}) },
				filters: {
					video: p.spec?.filters?.video ?? [],
					audio: p.spec?.filters?.audio ?? []
				},
				subtitles: { mode: 'copy', ...(p.spec?.subtitles ?? {}) },
				rawExtras: p.spec?.rawExtras ?? []
			}
			Object.assign(draft, merged)
		} catch (e) {
			saveError.value = e instanceof Error ? e.message : String(e)
		}
	}
	runPreview()
})

const runPreview = debounce(async () => {
	previewLoading.value = true
	try {
		const r: PreviewResponse = await api.presets.preview(draft)
		previewArgv.value = r.argv ?? []
		previewErrors.value = r.errors ?? []
	} catch (e) {
		previewErrors.value = [e instanceof Error ? e.message : String(e)]
	} finally {
		previewLoading.value = false
	}
}, 300)

watch(draft, runPreview, { deep: true })

// Codec change: clear preset/profile/level if they're no longer valid for the
// new codec. Keeps the form from stalling on stale enum picks.
watch(
	() => draft.video.codec,
	(name) => {
		const c = catalogStore.videoCodec(name)
		if (!c) return
		if (draft.video.preset && c.presets && !c.presets.includes(draft.video.preset)) draft.video.preset = ''
		if (draft.video.tune && c.tunes && !c.tunes.includes(draft.video.tune)) draft.video.tune = ''
		if (draft.video.profile && c.profiles && !c.profiles.includes(draft.video.profile)) draft.video.profile = ''
		if (draft.video.level && c.levels && !c.levels.includes(draft.video.level)) draft.video.level = ''
		if (c.pixelFormats?.length && draft.video.pixelFormat && !c.pixelFormats.includes(draft.video.pixelFormat)) {
			draft.video.pixelFormat = c.pixelFormats[0]
		}
		if (c.rateModes && !c.rateModes.includes(draft.video.rate.mode)) {
			draft.video.rate.mode = c.rateModes[0] ?? 'crf'
		}
	}
)

watch(
	() => draft.audio.codec,
	(name) => {
		const c = name ? catalogStore.audioCodec(name) : undefined
		if (!c) return
		if (draft.audio.profile && c.profiles && !c.profiles.includes(draft.audio.profile)) draft.audio.profile = ''
	}
)

function setResolution(width: number | null, height: number | null) {
	if (width == null || height == null || width <= 0 || height <= 0) {
		draft.video.resolution = null
		return
	}
	draft.video.resolution = { width, height }
}

function addVideoFilter(name: string) {
	if (!name) return
	draft.filters!.video!.push({ name, args: {} })
}
function removeVideoFilter(idx: number) {
	draft.filters!.video!.splice(idx, 1)
}
function moveVideoFilter(idx: number, delta: number) {
	const arr = draft.filters!.video!
	const j = idx + delta
	if (j < 0 || j >= arr.length) return
	const [s] = arr.splice(idx, 1)
	arr.splice(j, 0, s)
}
function addAudioFilter(name: string) {
	if (!name) return
	draft.filters!.audio!.push({ name, args: {} })
}
function removeAudioFilter(idx: number) {
	draft.filters!.audio!.splice(idx, 1)
}

function rawExtrasText(): string {
	return (draft.rawExtras ?? []).join(' ')
}
function setRawExtras(text: string) {
	const tokens = text.split(/\s+/).filter(Boolean)
	draft.rawExtras = tokens
}

const newVideoFilterName = ref('')
const newAudioFilterName = ref('')
function commitNewVideoFilter() {
	if (newVideoFilterName.value) {
		addVideoFilter(newVideoFilterName.value)
		newVideoFilterName.value = ''
	}
}
function commitNewAudioFilter() {
	if (newAudioFilterName.value) {
		addAudioFilter(newAudioFilterName.value)
		newAudioFilterName.value = ''
	}
}

async function save() {
	saveError.value = null
	saving.value = true
	try {
		if (isNew.value || builtin.value) {
			const created = await presetsStore.create({ name: name.value, description: description.value, spec: draft })
			router.replace({ name: 'preset-edit', params: { id: created.id } })
		} else {
			await presetsStore.update(String(route.params.id), {
				name: name.value,
				description: description.value,
				spec: draft
			})
		}
	} catch (e) {
		saveError.value = e instanceof Error ? e.message : String(e)
	} finally {
		saving.value = false
	}
}

function reset() {
	Object.assign(draft, blankSpec())
}
</script>

<template>
	<div class="space-y-4">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<div>
				<h1 class="text-xl font-semibold">{{ isNew ? 'New preset' : `Edit: ${name}` }}</h1>
				<p v-if="builtin" class="text-xs text-base-content/60 mt-1">
					Built-in preset — saving creates a copy you can modify.
				</p>
			</div>
			<div class="flex gap-2">
				<Button variant="ghost" size="sm" @click="$router.push({ name: 'presets' })">Cancel</Button>
				<Button v-if="!isNew && !builtin" variant="ghost" size="sm" @click="reset">Reset</Button>
				<Button :loading="saving" size="sm" @click="save">
					{{ isNew || builtin ? 'Save as new' : 'Save' }}
				</Button>
			</div>
		</div>

		<div v-if="saveError" class="alert alert-error text-sm"><span>{{ saveError }}</span></div>
		<div v-if="catalog === null && !catalogStore.loading" class="alert alert-warning text-sm">
			<span>Catalog not loaded. Check /api/presets/schema.</span>
		</div>

		<div class="grid lg:grid-cols-[2fr_1fr] gap-4">
			<!-- LEFT: Form -->
			<div class="space-y-4">
				<!-- Identity -->
				<fieldset class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-3">
						<legend class="font-medium">Identity</legend>
						<label class="form-control">
							<span class="label-text text-xs">Name</span>
							<input v-model="name" class="input input-bordered input-sm" placeholder="my-h264-1080p" />
						</label>
						<label class="form-control">
							<span class="label-text text-xs">Description</span>
							<input v-model="description" class="input input-bordered input-sm" />
						</label>
					</div>
				</fieldset>

				<!-- Tabs -->
				<div role="tablist" class="tabs tabs-boxed">
					<button
						v-for="tab in (['container','video','audio','filters','hwaccel','advanced'] as const)"
						:key="tab"
						role="tab"
						class="tab"
						:class="{ 'tab-active': activeTab === tab }"
						@click="activeTab = tab"
					>{{ tab }}</button>
				</div>

				<!-- CONTAINER -->
				<section v-show="activeTab === 'container'" class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-3">
						<legend class="font-medium">Container</legend>
						<label class="form-control">
							<span class="label-text text-xs">Format</span>
							<select v-model="draft.container.format" class="select select-bordered select-sm">
								<option v-for="c in catalogStore.containers" :key="c.format" :value="c.format">
									{{ c.format }} ({{ c.mime }})
								</option>
							</select>
						</label>
						<label class="form-control">
							<input type="checkbox" class="toggle toggle-sm" v-model="draft.container.faststart" />
							<span class="label-text text-xs ml-2">Faststart (mp4 only — moves moov atom for streaming)</span>
						</label>
						<label class="form-control">
							<span class="label-text text-xs">Output filename suffix (e.g. _1080p)</span>
							<input v-model="draft.outputSuffix" class="input input-bordered input-sm" />
						</label>
					</div>
				</section>

				<!-- VIDEO -->
				<section v-show="activeTab === 'video'" class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-3">
						<legend class="font-medium">Video</legend>
						<label class="form-control">
							<input type="checkbox" class="toggle toggle-sm" v-model="draft.video.disabled" />
							<span class="label-text text-xs ml-2">Disable video stream (-vn)</span>
						</label>
						<template v-if="!draft.video.disabled">
							<label class="form-control">
								<span class="label-text text-xs">Codec</span>
								<select v-model="draft.video.codec" class="select select-bordered select-sm">
									<option v-for="c in catalogStore.videoCodecs" :key="c.name" :value="c.name">
										{{ c.displayName }}
									</option>
								</select>
								<span v-if="codecCatalog?.description" class="text-xs text-base-content/60 mt-1">
									{{ codecCatalog.description }}
								</span>
							</label>

							<template v-if="showVideoFields">
								<div class="grid grid-cols-2 gap-3">
									<label v-if="codecCatalog?.presets?.length" class="form-control">
										<span class="label-text text-xs">Preset</span>
										<select v-model="draft.video.preset" class="select select-bordered select-sm">
											<option value="">—</option>
											<option v-for="p in codecCatalog.presets" :key="p" :value="p">{{ p }}</option>
										</select>
									</label>
									<label v-if="codecCatalog?.tunes?.length" class="form-control">
										<span class="label-text text-xs">Tune</span>
										<select v-model="draft.video.tune" class="select select-bordered select-sm">
											<option value="">—</option>
											<option v-for="t in codecCatalog.tunes" :key="t" :value="t">{{ t }}</option>
										</select>
									</label>
									<label v-if="codecCatalog?.profiles?.length" class="form-control">
										<span class="label-text text-xs">Profile</span>
										<select v-model="draft.video.profile" class="select select-bordered select-sm">
											<option value="">—</option>
											<option v-for="p in codecCatalog.profiles" :key="p" :value="p">{{ p }}</option>
										</select>
									</label>
									<label v-if="codecCatalog?.levels?.length" class="form-control">
										<span class="label-text text-xs">Level</span>
										<select v-model="draft.video.level" class="select select-bordered select-sm">
											<option value="">—</option>
											<option v-for="l in codecCatalog.levels" :key="l" :value="l">{{ l }}</option>
										</select>
									</label>
									<label class="form-control">
										<span class="label-text text-xs">Pixel format</span>
										<select v-model="draft.video.pixelFormat" class="select select-bordered select-sm">
											<option value="">auto</option>
											<option v-for="pf in (codecCatalog?.pixelFormats?.length ? codecCatalog.pixelFormats : catalogStore.pixelFormats)" :key="pf" :value="pf">{{ pf }}</option>
										</select>
									</label>
									<label class="form-control">
										<span class="label-text text-xs">Frame rate</span>
										<input v-model="draft.video.frameRate" placeholder="e.g. 24, 30, 60000/1001" class="input input-bordered input-sm" />
									</label>
								</div>

								<!-- Rate mode tabs -->
								<div class="border-t border-base-300 pt-3 mt-2 space-y-2">
									<label class="form-control">
										<span class="label-text text-xs">Rate mode</span>
										<div class="join">
											<button
												v-for="mode in (codecCatalog?.rateModes ?? [])"
												:key="mode"
												class="btn btn-sm join-item"
												:class="{ 'btn-primary': draft.video.rate.mode === mode }"
												@click="draft.video.rate.mode = mode"
											>{{ mode }}</button>
										</div>
									</label>

									<div v-if="showRateFields" class="grid grid-cols-2 gap-3">
										<label v-if="draft.video.rate.mode === 'crf'" class="form-control">
											<span class="label-text text-xs">
												CRF
												<span v-if="codecCatalog?.crfRange" class="text-base-content/50">
													({{ codecCatalog.crfRange.min }}–{{ codecCatalog.crfRange.max }})
												</span>
											</span>
											<input
												type="number"
												v-model.number="draft.video.rate.crf"
												:min="codecCatalog?.crfRange?.min ?? 0"
												:max="codecCatalog?.crfRange?.max ?? 63"
												class="input input-bordered input-sm"
											/>
										</label>
										<label v-if="draft.video.rate.mode === 'qp'" class="form-control">
											<span class="label-text text-xs">QP</span>
											<input type="number" v-model.number="draft.video.rate.qp" class="input input-bordered input-sm" />
										</label>
										<label v-if="['cbr','vbr','abr'].includes(draft.video.rate.mode)" class="form-control">
											<span class="label-text text-xs">Bitrate (e.g. 5M, 5000k)</span>
											<input v-model="draft.video.rate.bitrate" class="input input-bordered input-sm" />
										</label>
										<label v-if="['vbr','abr','crf'].includes(draft.video.rate.mode)" class="form-control">
											<span class="label-text text-xs">Max rate</span>
											<input v-model="draft.video.rate.maxRate" class="input input-bordered input-sm" />
										</label>
										<label v-if="draft.video.rate.mode !== 'qp'" class="form-control">
											<span class="label-text text-xs">Buffer size</span>
											<input v-model="draft.video.rate.bufSize" class="input input-bordered input-sm" />
										</label>
									</div>

									<label v-if="codecCatalog?.allowsTwoPass" class="form-control">
										<input type="checkbox" class="toggle toggle-sm" v-model="draft.video.twoPass" />
										<span class="label-text text-xs ml-2">Two-pass encoding (better quality at target bitrate; doubles CPU)</span>
									</label>
								</div>

								<!-- Resolution -->
								<div class="border-t border-base-300 pt-3 mt-2">
									<div class="flex items-center gap-3">
										<input
											type="checkbox"
											class="toggle toggle-sm"
											:checked="!!draft.video.resolution"
											@change="(e) => (e.target as HTMLInputElement).checked
												? setResolution(1920, 1080)
												: setResolution(null, null)"
										/>
										<span class="label-text text-xs">Resize</span>
									</div>
									<div v-if="draft.video.resolution" class="grid grid-cols-2 gap-3 mt-2">
										<label class="form-control">
											<span class="label-text text-xs">Width</span>
											<input
												type="number"
												:value="draft.video.resolution.width"
												@input="(e) => setResolution(Number((e.target as HTMLInputElement).value), draft.video.resolution!.height)"
												class="input input-bordered input-sm"
											/>
										</label>
										<label class="form-control">
											<span class="label-text text-xs">Height</span>
											<input
												type="number"
												:value="draft.video.resolution.height"
												@input="(e) => setResolution(draft.video.resolution!.width, Number((e.target as HTMLInputElement).value))"
												class="input input-bordered input-sm"
											/>
										</label>
									</div>
								</div>
							</template>
						</template>
					</div>
				</section>

				<!-- AUDIO -->
				<section v-show="activeTab === 'audio'" class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-3">
						<legend class="font-medium">Audio</legend>
						<label class="form-control">
							<input type="checkbox" class="toggle toggle-sm" v-model="draft.audio.disabled" />
							<span class="label-text text-xs ml-2">Disable audio stream (-an)</span>
						</label>
						<template v-if="!draft.audio.disabled">
							<label class="form-control">
								<span class="label-text text-xs">Codec</span>
								<select v-model="draft.audio.codec" class="select select-bordered select-sm">
									<option v-for="c in catalogStore.audioCodecs" :key="c.name" :value="c.name">
										{{ c.displayName }}
									</option>
								</select>
							</label>
							<template v-if="draft.audio.codec && draft.audio.codec !== 'copy'">
								<div class="grid grid-cols-2 gap-3">
									<label class="form-control">
										<span class="label-text text-xs">
											Bitrate
											<span v-if="audioCodecCatalog?.bitrateKbps" class="text-base-content/50">
												({{ audioCodecCatalog.bitrateKbps.min }}–{{ audioCodecCatalog.bitrateKbps.max }} kbps)
											</span>
										</span>
										<input v-model="draft.audio.bitrate" placeholder="192k" class="input input-bordered input-sm" />
									</label>
									<label class="form-control">
										<span class="label-text text-xs">Sample rate</span>
										<select v-model.number="draft.audio.sampleRate" class="select select-bordered select-sm">
											<option :value="0">auto</option>
											<option v-for="sr in audioCodecCatalog?.sampleRates ?? []" :key="sr" :value="sr">{{ sr }}</option>
										</select>
									</label>
									<label class="form-control">
										<span class="label-text text-xs">Channels</span>
										<select v-model.number="draft.audio.channels" class="select select-bordered select-sm">
											<option :value="0">auto</option>
											<option v-for="ch in audioCodecCatalog?.channels ?? []" :key="ch" :value="ch">{{ ch }}</option>
										</select>
									</label>
									<label v-if="audioCodecCatalog?.profiles?.length" class="form-control">
										<span class="label-text text-xs">Profile</span>
										<select v-model="draft.audio.profile" class="select select-bordered select-sm">
											<option value="">—</option>
											<option v-for="p in audioCodecCatalog.profiles" :key="p" :value="p">{{ p }}</option>
										</select>
									</label>
								</div>
							</template>
						</template>
					</div>
				</section>

				<!-- FILTERS -->
				<section v-show="activeTab === 'filters'" class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-3">
						<legend class="font-medium">Filter chain</legend>

						<div class="space-y-2">
							<div class="text-xs text-base-content/60">Video filters (applied in order)</div>
							<div v-if="!draft.filters?.video?.length" class="text-xs text-base-content/50 italic">No video filters.</div>
							<div
								v-for="(step, i) in draft.filters?.video ?? []"
								:key="i"
								class="border border-base-300 rounded p-3 space-y-2"
							>
								<div class="flex items-center justify-between gap-2">
									<span class="font-mono text-sm">{{ step.name }}</span>
									<div class="flex gap-1">
										<button class="btn btn-xs btn-ghost" @click="moveVideoFilter(i, -1)" :disabled="i === 0">▲</button>
										<button class="btn btn-xs btn-ghost" @click="moveVideoFilter(i, 1)" :disabled="i === (draft.filters?.video?.length ?? 0) - 1">▼</button>
										<button class="btn btn-xs btn-error btn-outline" @click="removeVideoFilter(i)">✕</button>
									</div>
								</div>
								<FilterArgs :step="step" :catalog="catalogStore.filterSpec(step.name)" />
							</div>
							<div class="flex gap-2">
								<select class="select select-bordered select-sm flex-1" v-model="newVideoFilterName">
									<option value="">+ Add video filter…</option>
									<option v-for="f in videoFilterCatalog" :key="f.name" :value="f.name">
										{{ f.displayName }}
									</option>
								</select>
								<Button variant="ghost" size="xs" @click="commitNewVideoFilter">Add</Button>
							</div>
						</div>

						<div class="space-y-2 mt-4">
							<div class="text-xs text-base-content/60">Audio filters</div>
							<div v-if="!draft.filters?.audio?.length" class="text-xs text-base-content/50 italic">No audio filters.</div>
							<div
								v-for="(step, i) in draft.filters?.audio ?? []"
								:key="i"
								class="border border-base-300 rounded p-3 space-y-2"
							>
								<div class="flex items-center justify-between gap-2">
									<span class="font-mono text-sm">{{ step.name }}</span>
									<button class="btn btn-xs btn-error btn-outline" @click="removeAudioFilter(i)">✕</button>
								</div>
								<FilterArgs :step="step" :catalog="catalogStore.filterSpec(step.name)" />
							</div>
							<div class="flex gap-2">
								<select class="select select-bordered select-sm flex-1" v-model="newAudioFilterName">
									<option value="">+ Add audio filter…</option>
									<option v-for="f in audioFilterCatalog" :key="f.name" :value="f.name">
										{{ f.displayName }}
									</option>
								</select>
								<Button variant="ghost" size="xs" @click="commitNewAudioFilter">Add</Button>
							</div>
						</div>
					</div>
				</section>

				<!-- HWACCEL -->
				<section v-show="activeTab === 'hwaccel'" class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-3">
						<legend class="font-medium">Hardware acceleration</legend>
						<label class="form-control">
							<span class="label-text text-xs">Type</span>
							<select
								class="select select-bordered select-sm"
								:value="draft.hwaccel?.type ?? 'none'"
								@change="(e) => {
									const v = (e.target as HTMLSelectElement).value
									if (v === 'none') draft.hwaccel = null
									else draft.hwaccel = { type: v as any }
								}"
							>
								<option v-for="h in catalogStore.hwaccels" :key="h.type" :value="h.type">
									{{ h.displayName }}
								</option>
							</select>
						</label>
						<template v-if="draft.hwaccel && draft.hwaccel.type !== 'none'">
							<label class="form-control">
								<span class="label-text text-xs">Device path</span>
								<input
									v-model="draft.hwaccel.device"
									:placeholder="catalogStore.hwaccel(draft.hwaccel.type)?.defaultDevice ?? ''"
									class="input input-bordered input-sm"
								/>
							</label>
							<label class="form-control">
								<span class="label-text text-xs">Output format</span>
								<select v-model="draft.hwaccel.outputFormat" class="select select-bordered select-sm">
									<option value="">default</option>
									<option v-for="f in catalogStore.hwaccel(draft.hwaccel.type)?.outputFormats ?? []" :key="f" :value="f">{{ f }}</option>
								</select>
							</label>
							<p class="text-xs text-base-content/60">
								Hardware acceleration must pair with a matching encoder. Compatible codecs:
								<span class="font-mono">{{ catalogStore.hwaccel(draft.hwaccel.type)?.codecNames.join(', ') }}</span>
							</p>
						</template>
					</div>
				</section>

				<!-- ADVANCED -->
				<section v-show="activeTab === 'advanced'" class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-3">
						<legend class="font-medium">Advanced</legend>

						<div class="grid grid-cols-2 gap-3">
							<label class="form-control">
								<span class="label-text text-xs">Threads (0 = auto)</span>
								<input type="number" :value="draft.threading?.threads ?? 0" @input="(e) => { draft.threading = { ...(draft.threading ?? {}), threads: Number((e.target as HTMLInputElement).value) } }" class="input input-bordered input-sm" />
							</label>
							<label class="form-control">
								<span class="label-text text-xs">Subtitle handling</span>
								<select
									class="select select-bordered select-sm"
									:value="draft.subtitles?.mode ?? 'copy'"
									@change="(e) => { draft.subtitles = { mode: (e.target as HTMLSelectElement).value as any } }"
								>
									<option value="copy">copy</option>
									<option value="strip">strip</option>
									<option value="burn">burn-in (use subtitles= filter in video chain)</option>
								</select>
							</label>
						</div>

						<label class="form-control">
							<span class="label-text text-xs">Raw extra ffmpeg args (allowlisted; one flag + value pair per token)</span>
							<input
								:value="rawExtrasText()"
								@input="(e) => setRawExtras((e.target as HTMLInputElement).value)"
								placeholder="-map 0:v:0 -map 0:a:0 -metadata title=Encoded"
								class="input input-bordered input-sm font-mono"
							/>
							<span class="text-xs text-base-content/50 mt-1">
								Forbidden: -i, -y, -f, -c:v, shell metacharacters. Set
								TIDAL_PRESET_RAW_EXTRAS_PERMISSIVE=true on the server to relax the allowlist.
							</span>
						</label>
					</div>
				</section>
			</div>

			<!-- RIGHT: Live preview -->
			<aside class="space-y-3 lg:sticky lg:top-4 lg:self-start">
				<div class="card bg-base-100 border border-base-300">
					<div class="card-body p-4 gap-2">
						<div class="flex items-center justify-between">
							<h3 class="font-medium">FFmpeg command</h3>
							<span v-if="previewLoading" class="loading loading-spinner loading-xs" />
						</div>
						<pre class="text-xs font-mono whitespace-pre-wrap break-all bg-base-200 rounded p-3 max-h-96 overflow-auto">ffmpeg {{ previewArgv.join(' ') }}</pre>
					</div>
				</div>
				<div v-if="previewErrors.length" class="card bg-error/5 border border-error/30">
					<div class="card-body p-4 gap-2">
						<h3 class="font-medium text-error">Validation errors</h3>
						<ul class="text-xs space-y-1">
							<li v-for="(e, i) in previewErrors" :key="i" class="text-error">{{ e }}</li>
						</ul>
					</div>
				</div>
			</aside>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

const FilterArgs = defineComponent({
	name: 'FilterArgs',
	props: {
		step: { type: Object as PropType<FilterStep>, required: true },
		catalog: { type: Object as PropType<CatalogFilter | undefined>, default: undefined }
	},
	setup(props) {
		function setArg(key: string, value: string) {
			if (!props.step.args) props.step.args = {}
			if (value === '') delete props.step.args[key]
			else props.step.args[key] = value
		}
		return { setArg }
	},
	template: `
		<div class="space-y-2">
			<div v-if="!catalog?.args?.length" class="text-xs text-base-content/50 italic">No arguments.</div>
			<div v-else class="grid grid-cols-2 gap-2">
				<label v-for="arg in catalog.args" :key="arg.name" class="form-control">
					<span class="label-text text-xs">
						{{ arg.name }}
						<span v-if="arg.required" class="text-error">*</span>
						<span v-if="arg.description" class="text-base-content/50">{{ ' — ' + arg.description }}</span>
					</span>
					<select
						v-if="arg.type === 'enum' && arg.enum?.length"
						class="select select-bordered select-xs"
						:value="step.args?.[arg.name] ?? arg.default ?? ''"
						@change="setArg(arg.name, ($event.target as HTMLSelectElement).value)"
					>
						<option value="">—</option>
						<option v-for="v in arg.enum" :key="v" :value="v">{{ v }}</option>
					</select>
					<input
						v-else
						:type="arg.type === 'int' || arg.type === 'float' ? 'number' : 'text'"
						class="input input-bordered input-xs"
						:placeholder="arg.default ?? ''"
						:value="step.args?.[arg.name] ?? ''"
						@input="setArg(arg.name, ($event.target as HTMLInputElement).value)"
					/>
				</label>
			</div>
		</div>
	`
})
export default { components: { FilterArgs } }
</script>
