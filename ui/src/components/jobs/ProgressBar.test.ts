import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ProgressBar from './ProgressBar.vue'

describe('ProgressBar', () => {
	it('renders a daisyUI progress element', () => {
		const wrapper = mount(ProgressBar, { props: { percent: 50 } })
		const bar = wrapper.find('progress.progress')
		expect(bar.exists()).toBe(true)
		expect(bar.attributes('value')).toBe('50')
		expect(bar.attributes('max')).toBe('100')
	})

	it('clamps percent to 0', () => {
		const wrapper = mount(ProgressBar, { props: { percent: -10 } })
		expect(wrapper.find('progress').attributes('value')).toBe('0')
	})

	it('clamps percent above 100', () => {
		const wrapper = mount(ProgressBar, { props: { percent: 150 } })
		expect(wrapper.find('progress').attributes('value')).toBe('100')
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
