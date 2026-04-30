import { onMounted, ref, computed } from 'vue';
import { storeToRefs } from 'pinia';
import { useJobsStore } from '@/stores/jobs';
import { usePresetsStore } from '@/stores/presets';
import StatusBadge from '@/components/jobs/StatusBadge.vue';
import ProgressBar from '@/components/jobs/ProgressBar.vue';
import EmptyState from '@/components/common/EmptyState.vue';
import Button from '@/components/common/Button.vue';
import Modal from '@/components/common/Modal.vue';
const jobs = useJobsStore();
const presets = usePresetsStore();
const { items, loading, error } = storeToRefs(jobs);
const DEFAULT_CACHE = '/var/cache/tidal';
const showNew = ref(false);
const form = ref({
    presetId: '',
    sourcePath: '',
    outputPath: '',
    cachePath: DEFAULT_CACHE,
    sourceMovePath: ''
});
const submitting = ref(false);
const submitErr = ref(null);
onMounted(async () => {
    await Promise.all([jobs.load(), presets.load()]);
    jobs.startFirehose();
    if (presets.items.length && !form.value.presetId)
        form.value.presetId = presets.items[0].id;
});
const sortedJobs = computed(() => [...items.value].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt)));
async function submit() {
    submitErr.value = null;
    if (!form.value.presetId || !form.value.sourcePath) {
        submitErr.value = 'preset and source required';
        return;
    }
    submitting.value = true;
    try {
        await jobs.enqueue({
            presetId: form.value.presetId,
            sourcePath: form.value.sourcePath.trim(),
            outputPath: form.value.outputPath.trim() || undefined,
            cachePath: form.value.cachePath.trim() || undefined,
            sourceMovePath: form.value.sourceMovePath.trim() || undefined
        });
        showNew.value = false;
        form.value.sourcePath = '';
        form.value.outputPath = '';
        form.value.sourceMovePath = '';
        form.value.cachePath = DEFAULT_CACHE;
    }
    catch (e) {
        submitErr.value = e instanceof Error ? e.message : String(e);
    }
    finally {
        submitting.value = false;
    }
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
/** @type {[typeof Button, typeof Button, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(Button, new Button({
    ...{ 'onClick': {} },
}));
const __VLS_1 = __VLS_0({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
let __VLS_3;
let __VLS_4;
let __VLS_5;
const __VLS_6 = {
    onClick: (...[$event]) => {
        __VLS_ctx.showNew = true;
    }
};
__VLS_2.slots.default;
var __VLS_2;
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
else if (!__VLS_ctx.sortedJobs.length) {
    /** @type {[typeof EmptyState, typeof EmptyState, ]} */ ;
    // @ts-ignore
    const __VLS_7 = __VLS_asFunctionalComponent(EmptyState, new EmptyState({
        title: "No jobs yet",
        hint: "Enqueue your first transcode to see live progress, logs, and history.",
    }));
    const __VLS_8 = __VLS_7({
        title: "No jobs yet",
        hint: "Enqueue your first transcode to see live progress, logs, and history.",
    }, ...__VLS_functionalComponentArgsRest(__VLS_7));
    __VLS_9.slots.default;
    /** @type {[typeof Button, typeof Button, ]} */ ;
    // @ts-ignore
    const __VLS_10 = __VLS_asFunctionalComponent(Button, new Button({
        ...{ 'onClick': {} },
    }));
    const __VLS_11 = __VLS_10({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_10));
    let __VLS_13;
    let __VLS_14;
    let __VLS_15;
    const __VLS_16 = {
        onClick: (...[$event]) => {
            if (!!(__VLS_ctx.loading))
                return;
            if (!(!__VLS_ctx.sortedJobs.length))
                return;
            __VLS_ctx.showNew = true;
        }
    };
    __VLS_12.slots.default;
    var __VLS_12;
    var __VLS_9;
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "overflow-x-auto rounded-box border border-base-300 bg-base-100" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.table, __VLS_intrinsicElements.table)({
        ...{ class: "table table-zebra" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.thead, __VLS_intrinsicElements.thead)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.tr, __VLS_intrinsicElements.tr)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.th, __VLS_intrinsicElements.th)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.th, __VLS_intrinsicElements.th)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.th, __VLS_intrinsicElements.th)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.th, __VLS_intrinsicElements.th)({
        ...{ class: "w-72" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.th, __VLS_intrinsicElements.th)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.tbody, __VLS_intrinsicElements.tbody)({});
    for (const [j] of __VLS_getVForSourceType((__VLS_ctx.sortedJobs))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.tr, __VLS_intrinsicElements.tr)({
            ...{ onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.loading))
                        return;
                    if (!!(!__VLS_ctx.sortedJobs.length))
                        return;
                    __VLS_ctx.$router.push(`/jobs/${j.id}`);
                } },
            key: (j.id),
            ...{ class: "cursor-pointer hover" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.td, __VLS_intrinsicElements.td)({});
        /** @type {[typeof StatusBadge, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(StatusBadge, new StatusBadge({
            status: (j.status),
        }));
        const __VLS_18 = __VLS_17({
            status: (j.status),
        }, ...__VLS_functionalComponentArgsRest(__VLS_17));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.td, __VLS_intrinsicElements.td)({
            ...{ class: "font-mono text-xs" },
        });
        (j.sourcePath);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.td, __VLS_intrinsicElements.td)({
            ...{ class: "font-mono text-xs" },
        });
        (j.outputPath);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.td, __VLS_intrinsicElements.td)({});
        /** @type {[typeof ProgressBar, ]} */ ;
        // @ts-ignore
        const __VLS_20 = __VLS_asFunctionalComponent(ProgressBar, new ProgressBar({
            percent: (j.progress.percent),
        }));
        const __VLS_21 = __VLS_20({
            percent: (j.progress.percent),
        }, ...__VLS_functionalComponentArgsRest(__VLS_20));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.td, __VLS_intrinsicElements.td)({
            ...{ class: "text-xs text-base-content/60 whitespace-nowrap" },
        });
        (new Date(j.createdAt).toLocaleString());
    }
}
/** @type {[typeof Modal, typeof Modal, ]} */ ;
// @ts-ignore
const __VLS_23 = __VLS_asFunctionalComponent(Modal, new Modal({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.showNew),
    title: "New transcode job",
}));
const __VLS_24 = __VLS_23({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.showNew),
    title: "New transcode job",
}, ...__VLS_functionalComponentArgsRest(__VLS_23));
let __VLS_26;
let __VLS_27;
let __VLS_28;
const __VLS_29 = {
    onClose: (...[$event]) => {
        __VLS_ctx.showNew = false;
    }
};
__VLS_25.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.form, __VLS_intrinsicElements.form)({
    ...{ onSubmit: (__VLS_ctx.submit) },
    ...{ class: "space-y-4" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.fieldset, __VLS_intrinsicElements.fieldset)({
    ...{ class: "fieldset" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.legend, __VLS_intrinsicElements.legend)({
    ...{ class: "fieldset-legend" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.select, __VLS_intrinsicElements.select)({
    value: (__VLS_ctx.form.presetId),
    ...{ class: "select select-bordered w-full" },
});
for (const [p] of __VLS_getVForSourceType((__VLS_ctx.presets.items))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.option, __VLS_intrinsicElements.option)({
        key: (p.id),
        value: (p.id),
    });
    (p.name);
    (p.description);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.fieldset, __VLS_intrinsicElements.fieldset)({
    ...{ class: "fieldset" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.legend, __VLS_intrinsicElements.legend)({
    ...{ class: "fieldset-legend" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    placeholder: "/media/incoming/clip.mkv",
    ...{ class: "input input-bordered w-full font-mono text-sm" },
});
(__VLS_ctx.form.sourcePath);
__VLS_asFunctionalElement(__VLS_intrinsicElements.fieldset, __VLS_intrinsicElements.fieldset)({
    ...{ class: "fieldset" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.legend, __VLS_intrinsicElements.legend)({
    ...{ class: "fieldset-legend" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "text-base-content/50" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    placeholder: "leave empty for auto-derived",
    ...{ class: "input input-bordered w-full font-mono text-sm" },
});
(__VLS_ctx.form.outputPath);
__VLS_asFunctionalElement(__VLS_intrinsicElements.fieldset, __VLS_intrinsicElements.fieldset)({
    ...{ class: "fieldset" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.legend, __VLS_intrinsicElements.legend)({
    ...{ class: "fieldset-legend" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    placeholder: "/var/cache/tidal",
    ...{ class: "input input-bordered w-full font-mono text-sm" },
});
(__VLS_ctx.form.cachePath);
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "text-xs text-base-content/50" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.fieldset, __VLS_intrinsicElements.fieldset)({
    ...{ class: "fieldset" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.legend, __VLS_intrinsicElements.legend)({
    ...{ class: "fieldset-legend" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "text-base-content/50" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    placeholder: "/media/archive/  (or full path)",
    ...{ class: "input input-bordered w-full font-mono text-sm" },
});
(__VLS_ctx.form.sourceMovePath);
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "text-xs text-base-content/50" },
});
if (__VLS_ctx.submitErr) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "alert alert-error" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.submitErr);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "flex justify-end gap-2" },
});
/** @type {[typeof Button, typeof Button, ]} */ ;
// @ts-ignore
const __VLS_30 = __VLS_asFunctionalComponent(Button, new Button({
    ...{ 'onClick': {} },
    variant: "ghost",
    type: "button",
}));
const __VLS_31 = __VLS_30({
    ...{ 'onClick': {} },
    variant: "ghost",
    type: "button",
}, ...__VLS_functionalComponentArgsRest(__VLS_30));
let __VLS_33;
let __VLS_34;
let __VLS_35;
const __VLS_36 = {
    onClick: (...[$event]) => {
        __VLS_ctx.showNew = false;
    }
};
__VLS_32.slots.default;
var __VLS_32;
/** @type {[typeof Button, typeof Button, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(Button, new Button({
    type: "submit",
    loading: (__VLS_ctx.submitting),
}));
const __VLS_38 = __VLS_37({
    type: "submit",
    loading: (__VLS_ctx.submitting),
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
var __VLS_39;
var __VLS_25;
/** @type {__VLS_StyleScopedClasses['space-y-6']} */ ;
/** @type {__VLS_StyleScopedClasses['flex']} */ ;
/** @type {__VLS_StyleScopedClasses['flex-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['items-center']} */ ;
/** @type {__VLS_StyleScopedClasses['justify-between']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-3']} */ ;
/** @type {__VLS_StyleScopedClasses['max-w-prose']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/60']} */ ;
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
/** @type {__VLS_StyleScopedClasses['overflow-x-auto']} */ ;
/** @type {__VLS_StyleScopedClasses['rounded-box']} */ ;
/** @type {__VLS_StyleScopedClasses['border']} */ ;
/** @type {__VLS_StyleScopedClasses['border-base-300']} */ ;
/** @type {__VLS_StyleScopedClasses['bg-base-100']} */ ;
/** @type {__VLS_StyleScopedClasses['table']} */ ;
/** @type {__VLS_StyleScopedClasses['table-zebra']} */ ;
/** @type {__VLS_StyleScopedClasses['w-72']} */ ;
/** @type {__VLS_StyleScopedClasses['cursor-pointer']} */ ;
/** @type {__VLS_StyleScopedClasses['hover']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-xs']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-xs']} */ ;
/** @type {__VLS_StyleScopedClasses['text-xs']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/60']} */ ;
/** @type {__VLS_StyleScopedClasses['whitespace-nowrap']} */ ;
/** @type {__VLS_StyleScopedClasses['space-y-4']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset-legend']} */ ;
/** @type {__VLS_StyleScopedClasses['select']} */ ;
/** @type {__VLS_StyleScopedClasses['select-bordered']} */ ;
/** @type {__VLS_StyleScopedClasses['w-full']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset-legend']} */ ;
/** @type {__VLS_StyleScopedClasses['input']} */ ;
/** @type {__VLS_StyleScopedClasses['input-bordered']} */ ;
/** @type {__VLS_StyleScopedClasses['w-full']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset-legend']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['input']} */ ;
/** @type {__VLS_StyleScopedClasses['input-bordered']} */ ;
/** @type {__VLS_StyleScopedClasses['w-full']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset-legend']} */ ;
/** @type {__VLS_StyleScopedClasses['input']} */ ;
/** @type {__VLS_StyleScopedClasses['input-bordered']} */ ;
/** @type {__VLS_StyleScopedClasses['w-full']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['text-xs']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset']} */ ;
/** @type {__VLS_StyleScopedClasses['fieldset-legend']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['input']} */ ;
/** @type {__VLS_StyleScopedClasses['input-bordered']} */ ;
/** @type {__VLS_StyleScopedClasses['w-full']} */ ;
/** @type {__VLS_StyleScopedClasses['font-mono']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['text-xs']} */ ;
/** @type {__VLS_StyleScopedClasses['text-base-content/50']} */ ;
/** @type {__VLS_StyleScopedClasses['alert']} */ ;
/** @type {__VLS_StyleScopedClasses['alert-error']} */ ;
/** @type {__VLS_StyleScopedClasses['flex']} */ ;
/** @type {__VLS_StyleScopedClasses['justify-end']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-2']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            StatusBadge: StatusBadge,
            ProgressBar: ProgressBar,
            EmptyState: EmptyState,
            Button: Button,
            Modal: Modal,
            presets: presets,
            loading: loading,
            error: error,
            showNew: showNew,
            form: form,
            submitting: submitting,
            submitErr: submitErr,
            sortedJobs: sortedJobs,
            submit: submit,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
