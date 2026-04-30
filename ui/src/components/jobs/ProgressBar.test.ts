import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ProgressBar from './ProgressBar.vue'

describe('ProgressBar', () => {
	it('renders the bar with correct width', () => {
		const wrapper = mount(ProgressBar, { props: { percent: 50 } })
		const bar = wrapper.find('.bg-tidal-500')
		expect(bar.exists()).toBe(true)
		expect(bar.attributes('style')).toContain('width: 50%')
	})

	it('clamps percent to 0-100', () => {
		const wrapper = mount(ProgressBar, { props: { percent: -10 } })
		const bar = wrapper.find('.bg-tidal-500')
		expect(bar.attributes('style')).toContain('width: 0%')
	})

	it('clamps percent above 100', () => {
		const wrapper = mount(ProgressBar, { props: { percent: 150 } })
		const bar = wrapper.find('.bg-tidal-500')
		expect(bar.attributes('style')).toContain('width: 100%')
	})

	it('renders label and percentage when label provided', () => {
		const wrapper = mount(ProgressBar, { props: { percent: 75, label: 'Encoding' } })
		expect(wrapper.text()).toContain('Encoding')
		expect(wrapper.text()).toContain('75.0%')
	})

	it('does not render label when not provided', () => {
		const wrapper = mount(ProgressBar, { props: { percent: 50 } })
		expect(wrapper.text()).not.toContain('%')
	})
})
