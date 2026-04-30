import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import StatusBadge from './StatusBadge.vue';
describe('StatusBadge', () => {
    it('renders the status text', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'running' } });
        expect(wrapper.text()).toContain('running');
    });
    it('applies queued styling', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'queued' } });
        expect(wrapper.classes()).toContain('bg-slate-500/20');
    });
    it('applies succeeded styling', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'succeeded' } });
        expect(wrapper.classes()).toContain('bg-emerald-500/20');
    });
    it('applies failed styling', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'failed' } });
        expect(wrapper.classes()).toContain('bg-red-500/20');
    });
    it('applies cancelled styling', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'cancelled' } });
        expect(wrapper.classes()).toContain('bg-zinc-500/20');
    });
    it('applies dispatched styling', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'dispatched' } });
        expect(wrapper.classes()).toContain('bg-amber-500/20');
    });
    it('applies cancelling styling', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'cancelling' } });
        expect(wrapper.classes()).toContain('bg-orange-500/20');
    });
    it('applies running styling', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'running' } });
        expect(wrapper.classes()).toContain('bg-tidal-500/25');
    });
    it('falls back to queued styling for unknown status', () => {
        const wrapper = mount(StatusBadge, { props: { status: 'unknown' } });
        expect(wrapper.classes()).toContain('bg-slate-500/20');
    });
});
