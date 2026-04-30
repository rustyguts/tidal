<script setup lang="ts">
defineProps<{ title: string; hint?: string }>()
</script>

<template>
	<div class="card bg-base-100 border border-dashed border-base-300">
		<div class="card-body items-center text-center py-16">
			<h3 class="card-title">{{ title }}</h3>
			<p v-if="hint" class="max-w-md text-base-content/60">{{ hint }}</p>
			<div class="card-actions mt-2">
				<slot />
			</div>
		</div>
	</div>
</template>
