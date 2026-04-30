import { onMounted, ref } from 'vue';
import { storeToRefs } from 'pinia';
import { usePresetsStore } from '@/stores/presets';
import EmptyState from '@/components/common/EmptyState.vue';
import Button from '@/components/common/Button.vue';
const store = usePresetsStore();
const { items, loading, error } = storeToRefs(store);
onMounted(() => store.load());
const restoring = ref(false);
async function restoreDefaults() {
    restoring.value = true;
    try {
        await store.restoreDefaults();
    }
    finally {
        restoring.value = false;
    }
}
async function remove(id) {
    if (!confirm('Delete this preset?'))
        return;
    await store.remove(id);
}
async function duplicate(id, currentName) {
    const next = prompt('New preset name', `${currentName}-copy`);
    if (!next)
        return;
    await store.duplicate(id, next);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "space-y-6" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "flex flex-wrap items-center justify-between gap-3" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "max-w-prose text-sm text-base-content/60" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "flex gap-2" },
});
/** @type {[typeof Button, typeof Button, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(Button, new Button({
    ...{ 'onClick': {} },
    variant: "ghost",
    loading: (__VLS_ctx.restoring),
}));
const __VLS_1 = __VLS_0({
    ...{ 'onClick': {} },
    variant: "ghost",
    loading: (__VLS_ctx.restoring),
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
let __VLS_3;
let __VLS_4;
let __VLS_5;
const __VLS_6 = {
    onClick: (__VLS_ctx.restoreDefaults)
};
__VLS_2.slots.default;
var __VLS_2;
/** @type {[typeof Button, typeof Button, ]} */ ;
// @ts-ignore
const __VLS_7 = __VLS_asFunctionalComponent(Button, new Button({
    disabled: (true),
}));
const __VLS_8 = __VLS_7({
    disabled: (true),
}, ...__VLS_functionalComponentArgsRest(__VLS_7));
__VLS_9.slots.default;
var __VLS_9;
if (__VLS_ctx.error) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "alert alert-error" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.error);
}
if (__VLS_ctx.loading) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "flex items-center gap-2 text-sm text-base-content/60" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span)({
        ...{ class: "loading loading-spinner loading-sm" },
    });
}
else if (!__VLS_ctx.items.length) {
    /** @type {[typeof EmptyState, ]} */ ;
    // @ts-ignore
    const __VLS_10 = __VLS_asFunctionalComponent(EmptyState, new EmptyState({
        title: "No presets",
        hint: "Click Restore defaults to seed the built-in presets.",
    }));
    const __VLS_11 = __VLS_10({
        title: "No presets",
        hint: "Click Restore defaults to seed the built-in presets.",
    }, ...__VLS_functionalComponentArgsRest(__VLS_10));
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "grid gap-4 md:grid-cols-2 xl:grid-cols-3" },
    });
    for (const [p] of __VLS_getVForSourceType((__VLS_ctx.items))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (p.id),
            ...{ class: "card bg-base-100 border border-base-300 hover:border-primary/40 transition-colors" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-body" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "flex items-start justify-between gap-2" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
            ...{ class: "card-title" },
        });
        (p.name);
        if (p.builtin) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "badge badge-primary badge-outline" },
            });
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "text-sm text-base-content/60" },
        });
        (p.description || '—');
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dl, __VLS_intrinsicElements.dl)({
            ...{ class: "grid grid-cols-2 gap-x-3 gap-y-1 text-xs mt-2" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dt, __VLS_intrinsicElements.dt)({
            ...{ class: "text-base-content/50" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dd, __VLS_intrinsicElements.dd)({
            ...{ class: "font-mono" },
        });
        (p.spec.videoCodec);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dt, __VLS_intrinsicElements.dt)({
            ...{ class: "text-base-content/50" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dd, __VLS_intrinsicElements.dd)({
            ...{ class: "font-mono" },
        });
        (p.spec.crf);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dt, __VLS_intrinsicElements.dt)({
            ...{ class: "text-base-content/50" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dd, __VLS_intrinsicElements.dd)({
            ...{ class: "font-mono" },
        });
        (p.spec.container);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dt, __VLS_intrinsicElements.dt)({
            ...{ class: "text-base-content/50" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.dd, __VLS_intrinsicElements.dd)({
            ...{ class: "font-mono" },
        });
        (p.spec.resolution ? `${p.spec.resolution.width}×${p.spec.resolution.height}` : '—');
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-actions justify-end mt-2" },
        });
        /** @type {[typeof Button, typeof Button, ]} */ ;
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(Button, new Button({
            ...{ 'onClick': {} },
            variant: "ghost",
            size: "xs",
        }));
        const __VLS_14 = __VLS_13({
            ...{ 'onClick': {} },
            variant: "ghost",
            size: "xs",
        }, ...__VLS_functionalComponentArgsRest(__VLS_13));
        let __VLS_16;
        let __VLS_17;
        let __VLS_18;
        const __VLS_19 = {
            onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loading))
                    return;
                if (!!(!__VLS_ctx.items.length))
                    return;
                __VLS_ctx.duplicate(p.id, p.name);
            }
        };
        __VLS_15.slots.default;
        var __VLS_15;
        /** @type {[typeof Button, typeof Button, ]} */ ;
        // @ts-ignore
        const __VLS_20 = __VLS_asFunctionalComponent(Button, new Button({
            ...{ 'onClick': {} },
            variant: "danger",
            size: "xs",
        }));
        const __VLS_21 = __VLS_20({
            ...{ 'onClick': {} },
            variant: "danger",
            size: "xs",
        }, ...__VLS_functionalComponentArgsRest(__VLS_20));
        let __VLS_23;
        let __VLS_24;
        let __VLS_25;
        const __VLS_26 = {
            onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loading))
                    return;
                if (!!(!__VLS_ctx.items.length))
                    return;
                __VLS_ctx.remove(p.id);
            }
        };
        __VLS_22.slots.default;
        var __VLS_22;
    }
}
/** @type {__VLS_StyleScopedClasses['space-y-6']} */ ;
/** @type {__VLS_StyleScopedClasses['flex']} */ ;
/** @type {__VLS_StyleScopedClasses['flex-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['items-center']} */ ;
/** @type {__VLS_StyleScopedClasses['justify-between']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-3']} */ ;
/** @type {__VLS_StyleScopedClasses['max-w-prose']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/60']} */ ;
/** @type {__VLS_StyleScopedClasses['flex']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-2']} */ ;
/** @type {__VLS_StyleScopedClasses['alert']} */ ;
/** @type {__VLS_StyleScopedClasses['alert-error']} */ ;
/** @type {__VLS_StyleScopedClasses['flex']} */ ;
/** @type {__VLS_StyleScopedClasses['items-center']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-2']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/60']} */ ;
/** @type {__VLS_StyleScopedClasses['loading']} */ ;
/** @type {__VLS_StyleScopedClasses['loading-spinner']} */ ;
/** @type {__VLS_StyleScopedClasses['loading-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['grid']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-4']} */ ;
/** @type {__VLS_StyleScopedClasses['md:grid-cols-2']} */ ;
/** @type {__VLS_StyleScopedClasses['xl:grid-cols-3']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['bg-base-100']} */ ;
/** @type {__VLS_StyleScopedClasses['border']} */ ;
/** @type {__VLS_StyleScopedClasses['border-base-300']} */ ;
/** @type {__VLS_StyleScopedClasses['hover:border-primary/40']} */ ;
/** @type {__VLS_StyleScopedClasses['transition-colors']} */ ;
/** @type {__VLS_StyleScopedClasses['card-body']} */ ;
/** @type {__VLS_StyleScopedClasses['flex']} */ ;
/** @type {__VLS_StyleScopedClasses['items-start']} */ ;
/** @type {__VLS_StyleScopedClasses['justify-between']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-2']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['badge']} */ ;
/** @type {__VLS_StyleScopedClasses['badge-primary']} */ ;
/** @type {__VLS_StyleScopedClasses['badge-outline']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/60']} */ ;
/** @type {__VLS_StyleScopedClasses['grid']} */ ;
/** @type {__VLS_StyleScopedClasses['grid-cols-2']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-x-3']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-y-1']} */ ;
/** @type {__VLS_StyleScopedClasses['text-xs']} */ ;
/** @type {__VLS_StyleScopedClasses['mt-2']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['card-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['justify-end']} */ ;
/** @type {__VLS_StyleScopedClasses['mt-2']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            EmptyState: EmptyState,
            Button: Button,
            items: items,
            loading: loading,
            error: error,
            restoring: restoring,
            restoreDefaults: restoreDefaults,
            remove: remove,
            duplicate: duplicate,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
