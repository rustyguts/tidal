import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import StatusBadge from './StatusBadge.vue'

// daisyUI badge classes — see palette in StatusBadge.vue
describe('StatusBadge', () => {
	it('renders the status text', () => {
		const wrapper = mount(StatusBadge, { props: { status: 'running' } })
		expect(wrapper.text()).toContain('running')
	})

	const cases: Array<[string, string]> = [
		['queued', 'badge-ghost'],
		['dispatched', 'badge-warning'],
		['running', 'badge-info'],
		['cancelling', 'badge-warning'],
		['succeeded', 'badge-success'],
		['failed', 'badge-error'],
		['cancelled', 'badge-neutral']
	]

	for (const [status, expected] of cases) {
		it(`applies ${expected} for ${status}`, () => {
			const wrapper = mount(StatusBadge, { props: { status: status as any } })
			expect(wrapper.classes()).toContain(expected)
		})
	}

	it('falls back to badge-ghost for unknown status', () => {
		const wrapper = mount(StatusBadge, { props: { status: 'unknown' as any } })
		expect(wrapper.classes()).toContain('badge-ghost')
	})
})
