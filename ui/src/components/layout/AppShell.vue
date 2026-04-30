<script setup lang="ts">
import { useSystemStore } from '@/stores/system'
import SideNav from './SideNav.vue'
const sys = useSystemStore()
</script>

<template>
	<div class="drawer lg:drawer-open min-h-screen bg-base-200">
		<input id="tidal-drawer" type="checkbox" class="drawer-toggle" />

		<div class="drawer-content flex flex-col">
			<header class="navbar bg-base-100/80 backdrop-blur sticky top-0 z-20 border-b border-base-300">
				<div class="flex-none lg:hidden">
					<label for="tidal-drawer" class="btn btn-square btn-ghost">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>
					</label>
				</div>
				<div class="flex-1 px-2">
					<h1 class="text-lg font-semibold">{{ String($route.meta?.title ?? '') || 'Tidal' }}</h1>
				</div>
				<div class="flex-none">
					<span v-if="sys.info" class="badge badge-ghost font-mono text-xs">{{ sys.info.version.version }}</span>
				</div>
			</header>

			<main class="flex-1 px-4 py-6 sm:px-6 lg:px-8">
				<slot />
			</main>
		</div>

		<div class="drawer-side z-30">
			<label for="tidal-drawer" aria-label="close sidebar" class="drawer-overlay"></label>
			<SideNav />
		</div>
	</div>
</template>
