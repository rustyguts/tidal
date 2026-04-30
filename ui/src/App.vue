<script setup lang="ts">
import AppShell from '@/components/layout/AppShell.vue'
import { useSystemStore } from '@/stores/system'
import { onMounted } from 'vue'

const sys = useSystemStore()
onMounted(() => sys.load())
</script>

<template>
	<AppShell>
		<router-view />
	</AppShell>
</template>
