<script setup lang="ts">
import { RouterLink } from 'vue-router'

const links = [
	{ to: '/jobs', label: 'Jobs' },
	{ to: '/presets', label: 'Presets' },
	{ to: '/workflows', label: 'Workflows' },
	{ to: '/asynq', label: 'Queue' },
	{ to: '/settings', label: 'Settings' }
]
</script>

<template>
	<aside class="bg-base-100 min-h-full w-60 border-r border-base-300">
		<div class="flex items-center gap-3 px-5 py-5">
			<div class="avatar avatar-placeholder">
				<div class="bg-primary text-primary-content w-8 rounded-md">
					<span class="text-base font-bold">T</span>
				</div>
			</div>
			<div class="text-base font-semibold">Tidal</div>
		</div>

		<ul class="menu menu-md w-full px-3">
			<li v-for="l in links" :key="l.to">
				<RouterLink :to="l.to" active-class="menu-active">
					{{ l.label }}
				</RouterLink>
			</li>
		</ul>
	</aside>
</template>
