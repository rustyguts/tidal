<script setup lang="ts">
type Variant = 'primary' | 'secondary' | 'ghost' | 'outline' | 'danger' | 'neutral'
type Size = 'xs' | 'sm' | 'md' | 'lg'

const props = defineProps<{
	variant?: Variant
	size?: Size
	type?: 'button' | 'submit'
	loading?: boolean
	block?: boolean
}>()

const variantClass: Record<Variant, string> = {
	primary: 'btn-primary',
	secondary: 'btn-secondary',
	ghost: 'btn-ghost',
	outline: 'btn-outline',
	danger: 'btn-error',
	neutral: 'btn-neutral'
}
const sizeClass: Record<Size, string> = {
	xs: 'btn-xs',
	sm: 'btn-sm',
	md: '',
	lg: 'btn-lg'
}
</script>

<template>
	<button
		:type="type ?? 'button'"
		:class="['btn', variantClass[props.variant ?? 'primary'], sizeClass[props.size ?? 'sm'], props.block ? 'btn-block' : '']"
		:disabled="loading"
	>
		<span v-if="loading" class="loading loading-spinner loading-xs" />
		<slot />
	</button>
</template>
