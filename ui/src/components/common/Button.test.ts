import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Button from './Button.vue'

describe('Button', () => {
	it('renders slot content', () => {
		const wrapper = mount(Button, { slots: { default: 'Click me' } })
		expect(wrapper.text()).toBe('Click me')
	})

	it('always carries the daisyUI btn class', () => {
		const wrapper = mount(Button, { slots: { default: 'OK' } })
		expect(wrapper.classes()).toContain('btn')
	})

	it('defaults to primary variant', () => {
		const wrapper = mount(Button, { slots: { default: 'OK' } })
		expect(wrapper.classes()).toContain('btn-primary')
	})

	it('applies ghost variant', () => {
		const wrapper = mount(Button, { props: { variant: 'ghost' }, slots: { default: 'Cancel' } })
		expect(wrapper.classes()).toContain('btn-ghost')
	})

	it('applies danger variant (btn-error)', () => {
		const wrapper = mount(Button, { props: { variant: 'danger' }, slots: { default: 'Delete' } })
		expect(wrapper.classes()).toContain('btn-error')
	})

	it('applies size class', () => {
		const wrapper = mount(Button, { props: { size: 'lg' }, slots: { default: 'Big' } })
		expect(wrapper.classes()).toContain('btn-lg')
	})

	it('defaults type to button', () => {
		const wrapper = mount(Button, { slots: { default: 'OK' } })
		expect(wrapper.attributes('type')).toBe('button')
	})

	it('honors submit type', () => {
		const wrapper = mount(Button, { props: { type: 'submit' }, slots: { default: 'Go' } })
		expect(wrapper.attributes('type')).toBe('submit')
	})

	it('shows a spinner when loading', () => {
		const wrapper = mount(Button, { props: { loading: true }, slots: { default: 'Saving' } })
		expect(wrapper.find('.loading-spinner').exists()).toBe(true)
		expect(wrapper.attributes('disabled')).toBeDefined()
	})
})
