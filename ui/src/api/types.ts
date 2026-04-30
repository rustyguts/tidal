// Hand-mirrored from Go domain types in internal/domain.
// Phase 8 will replace this with openapi-typescript codegen from api/openapi.yaml.

export interface Resolution {
	width: number
	height: number
}

export interface PresetSpec {
	container: 'mp4' | 'mkv' | 'webm' | 'mov' | string
	videoCodec: string
	videoPreset?: string
	crf: number
	pixelFormat?: string
	audioCodec?: string
	audioBitrate?: string
	extraArgs?: string[]
	outputSuffix?: string
	resolution?: Resolution | null
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
	sourcePath: string
	outputPath: string
	cachePath?: string
	sourceMovePath?: string
	status: JobStatus
	k8sJobName?: string
	automationId?: string | null
	progress: FFmpegProgress
	error?: string
	createdAt: string
	startedAt?: string | null
	finishedAt?: string | null
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
