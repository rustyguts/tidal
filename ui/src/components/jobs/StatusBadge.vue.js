const props = defineProps();
const palette = {
    queued: 'badge-ghost',
    dispatched: 'badge-warning',
    running: 'badge-info',
    cancelling: 'badge-warning',
    succeeded: 'badge-success',
    failed: 'badge-error',
    cancelled: 'badge-neutral'
};
const cls = palette[props.status] ?? 'badge-ghost';
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: (['badge badge-sm gap-1', __VLS_ctx.cls]) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span)({
    ...{ class: "inline-block size-1.5 rounded-full bg-current opacity-70" },
});
(__VLS_ctx.status);
/** @type {__VLS_StyleScopedClasses['badge']} */ ;
/** @type {__VLS_StyleScopedClasses['badge-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-1']} */ ;
/** @type {__VLS_StyleScopedClasses['inline-block']} */ ;
/** @type {__VLS_StyleScopedClasses['size-1.5']} */ ;
/** @type {__VLS_StyleScopedClasses['rounded-full']} */ ;
/** @type {__VLS_StyleScopedClasses['bg-current']} */ ;
/** @type {__VLS_StyleScopedClasses['opacity-70']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            cls: cls,
        };
    },
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
