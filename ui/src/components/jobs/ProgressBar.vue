<script setup lang="ts">
const props = defineProps<{ percent: number; label?: string }>()
</script>

<template>
	<div class="w-full">
		<div v-if="label" class="mb-1 flex items-center justify-between text-xs text-base-content/60">
			<span>{{ label }}</span>
			<span class="font-mono">{{ props.percent.toFixed(1) }}%</span>
		</div>
		<progress
			class="progress progress-primary w-full"
			:value="Math.max(0, Math.min(100, props.percent))"
			max="100"
		/>
	</div>
</template>
