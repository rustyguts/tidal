const props = defineProps();
const variantClass = {
    primary: 'btn-primary',
    secondary: 'btn-secondary',
    ghost: 'btn-ghost',
    outline: 'btn-outline',
    danger: 'btn-error',
    neutral: 'btn-neutral'
};
const sizeClass = {
    xs: 'btn-xs',
    sm: 'btn-sm',
    md: '',
    lg: 'btn-lg'
};
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    type: (__VLS_ctx.type ?? 'button'),
    ...{ class: (['btn', __VLS_ctx.variantClass[props.variant ?? 'primary'], __VLS_ctx.sizeClass[props.size ?? 'sm'], props.block ? 'btn-block' : '']) },
    disabled: (__VLS_ctx.loading),
});
if (__VLS_ctx.loading) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span)({
        ...{ class: "loading loading-spinner loading-xs" },
    });
}
var __VLS_0 = {};
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['loading']} */ ;
/** @type {__VLS_StyleScopedClasses['loading-spinner']} */ ;
/** @type {__VLS_StyleScopedClasses['loading-xs']} */ ;
// @ts-ignore
var __VLS_1 = __VLS_0;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            variantClass: variantClass,
            sizeClass: sizeClass,
        };
    },
    __typeProps: {},
});
const __VLS_component = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
export default {};
; /* PartiallyEnd: #4569/main.vue */
