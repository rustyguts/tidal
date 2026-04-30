// Hand-mirrored from Go domain types in internal/domain.
// Phase 8 will replace this with openapi-typescript codegen from api/openapi.yaml.

export interface Resolution {
	width: number
	height: number
}

// ===== Preset spec =====
export type RateMode = 'crf' | 'cbr' | 'vbr' | 'qp' | 'abr' | 'none'

export interface VideoRate {
	mode: RateMode
	crf?: number
	qp?: number
	bitrate?: string
	maxRate?: string
	bufSize?: string
	minRate?: string
}

export interface GopSpec {
	keyintSec?: number
	keyintMin?: number
	sceneCut?: boolean
	bFrames?: number
	refFrames?: number
}

export interface ColorSpec {
	range?: string
	primaries?: string
	transfer?: string
	matrix?: string
}

export interface KV {
	key: string
	value: string
}

export interface VideoSpec {
	disabled?: boolean
	codec: string
	preset?: string
	tune?: string
	profile?: string
	level?: string
	pixelFormat?: string
	rate: VideoRate
	gop?: GopSpec
	twoPass?: boolean
	resolution?: Resolution | null
	frameRate?: string
	color?: ColorSpec | null
	codecExtra?: KV[]
}

export interface AudioSpec {
	disabled?: boolean
	codec?: string
	bitrate?: string
	sampleRate?: number
	channels?: number
	profile?: string
	vbrQuality?: number
}

export interface ContainerSpec {
	format: string
	faststart?: boolean
	movflags?: string[]
	fragmentDurMs?: number
}

export interface InputSpec {
	seekStart?: string
	duration?: string
	readRateLimit?: number
	protocolWhitelist?: string[]
}

export interface HwaccelSpec {
	type: 'none' | 'nvdec' | 'vaapi' | 'qsv' | 'videotoolbox'
	device?: string
	outputFormat?: string
}

export interface FilterStep {
	name: string
	args?: Record<string, string>
	enabled?: boolean
}

export interface FilterChain {
	video?: FilterStep[]
	audio?: FilterStep[]
}

export interface MappingSpec {
	video?: string
	audio?: string
	subtitle?: string
	streams?: string[]
}

export interface SubtitleSpec {
	mode?: 'copy' | 'burn' | 'strip'
	burnTrack?: number
}

export interface ThreadingSpec {
	threads?: number
	filterThreads?: number
}

export interface PresetSpec {
	container: ContainerSpec
	input?: InputSpec
	hwaccel?: HwaccelSpec | null
	video: VideoSpec
	audio: AudioSpec
	filters?: FilterChain
	mapping?: MappingSpec
	subtitles?: SubtitleSpec
	outputSuffix?: string
	rawExtras?: string[]
	threading?: ThreadingSpec
}

export interface Preset {
	id: string
	name: string
	description: string
	builtin: boolean
	spec: PresetSpec
	createdAt: string
	updatedAt: string
}

// ===== Catalog (served by GET /api/presets/schema) =====
export interface CatalogIntRange {
	min: number
	max: number
	default: number
}

export interface CatalogVideoCodec {
	name: string
	family: 'cpu' | 'nvenc' | 'vaapi' | 'qsv' | 'passthrough'
	displayName: string
	description?: string
	presets?: string[]
	tunes?: string[]
	profiles?: string[]
	levels?: string[]
	pixelFormats?: string[]
	crfRange?: CatalogIntRange
	qpRange?: CatalogIntRange
	rateModes: RateMode[]
	allowsTwoPass: boolean
	paramFlag?: string
}

export interface CatalogAudioCodec {
	name: string
	displayName: string
	description?: string
	sampleRates?: number[]
	channels?: number[]
	bitrateKbps?: CatalogIntRange
	allowsVBR: boolean
	vbrQuality?: CatalogIntRange
	profiles?: string[]
}

export interface CatalogContainer {
	format: string
	faststart?: boolean
	movflags?: string[]
	mime: string
}

export interface CatalogHwaccel {
	type: string
	displayName: string
	description?: string
	codecNames: string[]
	outputFormats: string[]
	defaultDevice?: string
}

export interface CatalogFilterArg {
	name: string
	type: 'string' | 'int' | 'float' | 'bool' | 'enum'
	required?: boolean
	default?: string
	description?: string
	enum?: string[]
	min?: number
	max?: number
}

export interface CatalogFilter {
	name: string
	scope: 'video' | 'audio'
	displayName: string
	description?: string
	args?: CatalogFilterArg[]
}

export interface Catalog {
	containers: CatalogContainer[]
	videoCodecs: CatalogVideoCodec[]
	audioCodecs: CatalogAudioCodec[]
	pixelFormats: string[]
	hwaccels: CatalogHwaccel[]
	filters: CatalogFilter[]
	rawExtrasAllow: string[]
	rawExtrasDeny: string[]
}

export interface SchemaResponse {
	catalog: Catalog
	schema: Record<string, unknown>
}

export interface PreviewResponse {
	argv: string[]
	errors?: string[]
	warnings?: string[]
	spec: PresetSpec
}

// ===== Jobs / Workflows / System =====
export type JobStatus =
	| 'queued'
	| 'dispatched'
	| 'running'
	| 'cancelling'
	| 'succeeded'
	| 'failed'
	| 'cancelled'

export interface FFmpegProgress {
	frame: number
	fps: number
	timeMs: number
	bitrate: number
	speed: number
	sizeBytes: number
	percent: number
	updatedAt: string
}

export interface Job {
	id: string
	asynqId?: string
	presetId: string
	specSnapshot?: PresetSpec
	sourcePath: string
	outputPath: string
	cachePath?: string
	sourceMovePath?: string
	status: JobStatus
	k8sJobName?: string
	workflowId?: string | null
	progress: FFmpegProgress
	error?: string
	createdAt: string
	startedAt?: string | null
	finishedAt?: string | null
}

export interface WorkflowTrigger {
	type: 'file_created'
	watchDir?: string
	glob?: string
}

export interface WorkflowAction {
	type: 'enqueue_transcode'
	presetId?: string
	outputPath?: string
	sourceMovePath?: string
	cachePath?: string
}

export interface Workflow {
	id: string
	name: string
	enabled: boolean
	trigger: WorkflowTrigger
	actions: WorkflowAction[]
	pollIntervalMs: number
	stableThresholdMs: number
	runsCount: number
	successCount: number
	lastRunAt?: string | null
	createdAt: string
	updatedAt: string
}

export interface WorkflowRun {
	id: number
	workflowId: string
	sourcePath: string
	jobId?: string | null
	outcome: string
	message?: string
	occurredAt: string
}

export interface JobLog {
	jobId: string
	seq: number
	stream: string
	line: string
	emittedAt: string
}

export interface SystemInfo {
	version: { version: string; commit: string; date: string }
	env: string
	dispatcher: string
	mediaRoots: string[]
}

export type SSEEventKind = 'status' | 'progress' | 'log'

export interface SSEEvent<T = unknown> {
	kind: SSEEventKind
	topic: string
	at: string
	data: T
}
