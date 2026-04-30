import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Button from './Button.vue'

describe('Button', () => {
	it('renders slot content', () => {
		const wrapper = mount(Button, { slots: { default: 'Click me' } })
		expect(wrapper.text()).toBe('Click me')
	})

	it('defaults to primary variant', () => {
		const wrapper = mount(Button, { slots: { default: 'OK' } })
		expect(wrapper.classes()).toContain('bg-tidal-600')
	})

	it('applies ghost variant', () => {
		const wrapper = mount(Button, { props: { variant: 'ghost' }, slots: { default: 'Cancel' } })
		expect(wrapper.classes()).toContain('bg-transparent')
	})

	it('applies danger variant', () => {
		const wrapper = mount(Button, { props: { variant: 'danger' }, slots: { default: 'Delete' } })
		expect(wrapper.classes()).toContain('bg-red-600/90')
	})

	it('defaults type to button', () => {
		const wrapper = mount(Button, { slots: { default: 'OK' } })
		expect(wrapper.attributes('type')).toBe('button')
	})

	it('accepts submit type', () => {
		const wrapper = mount(Button, { props: { type: 'submit' }, slots: { default: 'Submit' } })
		expect(wrapper.attributes('type')).toBe('submit')
	})
})
