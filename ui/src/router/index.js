import { createRouter, createWebHistory } from 'vue-router';
const routes = [
    { path: '/', redirect: '/jobs' },
    { path: '/jobs', name: 'jobs', component: () => import('@/views/JobsView.vue'), meta: { title: 'Jobs' } },
    { path: '/jobs/:id', name: 'job-detail', component: () => import('@/views/JobDetailView.vue'), meta: { title: 'Job' } },
    { path: '/presets', name: 'presets', component: () => import('@/views/PresetsView.vue'), meta: { title: 'Presets' } },
    { path: '/presets/:id', name: 'preset-edit', component: () => import('@/views/PresetEditorView.vue'), meta: { title: 'Edit preset' } },
    { path: '/workflows', name: 'workflows', component: () => import('@/views/WorkflowsView.vue'), meta: { title: 'Workflows' } },
    { path: '/workflows/new', name: 'workflow-new', component: () => import('@/views/WorkflowEditorView.vue'), meta: { title: 'New workflow' } },
    { path: '/workflows/:id', name: 'workflow-edit', component: () => import('@/views/WorkflowEditorView.vue'), meta: { title: 'Edit workflow' } },
    { path: '/asynq', name: 'asynq', component: () => import('@/views/AsynqView.vue'), meta: { title: 'Queue' } },
    { path: '/settings', name: 'settings', component: () => import('@/views/SettingsView.vue'), meta: { title: 'Settings' } },
    { path: '/:pathMatch(.*)*', component: () => import('@/views/NotFoundView.vue') }
];
export const router = createRouter({
    history: createWebHistory(),
    routes
});
router.afterEach((to) => {
    const t = to.meta?.title;
    if (typeof t === 'string')
        document.title = `${t} · Tidal`;
    else
        document.title = 'Tidal';
});
