<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
	<div class="card bg-base-100 border border-base-300">
		<div class="card-body">
			<p class="text-sm text-base-content/60">
				Preset editor route is wired ({{ route.params.id }}). Inline editor will land in a follow-up.
			</p>
		</div>
	</div>
</template>
