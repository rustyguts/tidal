<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'

const props = defineProps<{ open: boolean; title?: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const dialog = ref<HTMLDialogElement | null>(null)

watch(
	() => props.open,
	(o) => {
		const d = dialog.value
		if (!d) return
		if (o && !d.open) d.showModal()
		if (!o && d.open) d.close()
	}
)

function onClose() {
	emit('close')
}

onBeforeUnmount(() => {
	if (dialog.value?.open) dialog.value.close()
})
</script>

<template>
	<dialog ref="dialog" class="modal" @close="onClose">
		<div class="modal-box w-11/12 max-w-2xl">
			<form method="dialog">
				<button class="btn btn-sm btn-circle btn-ghost absolute end-2 top-2" aria-label="close">✕</button>
			</form>
			<h3 v-if="title" class="text-lg font-semibold">{{ title }}</h3>
			<div class="mt-4">
				<slot />
			</div>
		</div>
		<form method="dialog" class="modal-backdrop">
			<button>close</button>
		</form>
	</dialog>
</template>
