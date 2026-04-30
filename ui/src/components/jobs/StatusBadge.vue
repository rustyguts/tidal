<script setup lang="ts">
import type { JobStatus } from '@/api/types'

const props = defineProps<{ status: JobStatus }>()

const palette: Record<JobStatus, string> = {
	queued: 'badge-ghost',
	dispatched: 'badge-warning',
	running: 'badge-info',
	cancelling: 'badge-warning',
	succeeded: 'badge-success',
	failed: 'badge-error',
	cancelled: 'badge-neutral'
}
const cls = palette[props.status] ?? 'badge-ghost'
</script>

<template>
	<span :class="['badge badge-sm gap-1', cls]">
		<span class="inline-block size-1.5 rounded-full bg-current opacity-70" />
		{{ status }}
	</span>
</template>
