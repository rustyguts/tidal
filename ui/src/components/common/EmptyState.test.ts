import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import EmptyState from './EmptyState.vue'

describe('EmptyState', () => {
	it('renders title', () => {
		const wrapper = mount(EmptyState, { props: { title: 'No items' } })
		expect(wrapper.text()).toContain('No items')
	})

	it('renders hint when provided', () => {
		const wrapper = mount(EmptyState, { props: { title: 'Empty', hint: 'Create something' } })
		expect(wrapper.text()).toContain('Create something')
	})

	it('does not render hint when not provided', () => {
		const wrapper = mount(EmptyState, { props: { title: 'Empty' } })
		expect(wrapper.find('p.mt-1').exists()).toBe(false)
	})

	it('renders default slot', () => {
		const wrapper = mount(EmptyState, {
			props: { title: 'Empty' },
			slots: { default: '<button>Action</button>' }
		})
		expect(wrapper.find('button').exists()).toBe(true)
	})
})
