<script setup lang="ts">
import { useSystemStore } from '@/stores/system'
import { storeToRefs } from 'pinia'

const sys = useSystemStore()
const { info } = storeToRefs(sys)
</script>

<template>
	<div class="card bg-base-100 border border-base-300">
		<div class="card-body">
			<h2 class="card-title">System</h2>
			<table v-if="info" class="table table-sm w-full">
				<tbody>
					<tr><th class="w-40 text-base-content/60">Version</th><td class="font-mono">{{ info.version.version }}</td></tr>
					<tr><th class="text-base-content/60">Commit</th><td class="font-mono">{{ info.version.commit }}</td></tr>
					<tr><th class="text-base-content/60">Built</th><td class="font-mono">{{ info.version.date }}</td></tr>
					<tr><th class="text-base-content/60">Environment</th><td>{{ info.env }}</td></tr>
					<tr><th class="text-base-content/60">Dispatcher</th><td>{{ info.dispatcher }}</td></tr>
					<tr><th class="text-base-content/60">Media roots</th><td class="font-mono">{{ info.mediaRoots.join(', ') }}</td></tr>
				</tbody>
			</table>
			<p v-else class="flex items-center gap-2 text-sm text-base-content/60">
				<span class="loading loading-spinner loading-sm" /> Loading…
			</p>
		</div>
	</div>
</template>
