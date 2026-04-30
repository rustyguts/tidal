import { describe, it, expect, vi } from 'vitest'
import { blankSpec, debounce } from './presetDraft'

describe('blankSpec', () => {
	it('returns a v2 schema-versioned spec', () => {
		const s = blankSpec()
		expect(s.schemaVersion).toBe(2)
		expect(s.container.format).toBe('mp4')
		expect(s.container.faststart).toBe(true)
	})

	it('defaults to libx264 + crf 23', () => {
		const s = blankSpec()
		expect(s.video.codec).toBe('libx264')
		expect(s.video.preset).toBe('slow')
		expect(s.video.rate.mode).toBe('crf')
		expect(s.video.rate.crf).toBe(23)
	})

	it('seeds aac audio at 192k', () => {
		const s = blankSpec()
		expect(s.audio.codec).toBe('aac')
		expect(s.audio.bitrate).toBe('192k')
	})

	it('initializes empty filters and rawExtras', () => {
		const s = blankSpec()
		expect(s.filters?.video).toEqual([])
		expect(s.filters?.audio).toEqual([])
		expect(s.rawExtras).toEqual([])
	})

	it('returns independent objects on each call', () => {
		const a = blankSpec()
		const b = blankSpec()
		a.video.codec = 'libx265'
		expect(b.video.codec).toBe('libx264')
	})
})

describe('debounce', () => {
	it('coalesces rapid calls into one invocation', async () => {
		vi.useFakeTimers()
		try {
			const fn = vi.fn()
			const d = debounce(fn, 100)
			d()
			d()
			d()
			expect(fn).not.toHaveBeenCalled()
			vi.advanceTimersByTime(100)
			expect(fn).toHaveBeenCalledTimes(1)
		} finally {
			vi.useRealTimers()
		}
	})

	it('passes the latest arguments through', async () => {
		vi.useFakeTimers()
		try {
			const fn = vi.fn()
			const d = debounce(fn, 50)
			d('a' as never)
			d('b' as never)
			d('c' as never)
			vi.advanceTimersByTime(50)
			expect(fn).toHaveBeenCalledWith('c')
		} finally {
			vi.useRealTimers()
		}
	})

	it('fires multiple times across separate quiet periods', async () => {
		vi.useFakeTimers()
		try {
			const fn = vi.fn()
			const d = debounce(fn, 50)
			d()
			vi.advanceTimersByTime(50)
			d()
			vi.advanceTimersByTime(50)
			expect(fn).toHaveBeenCalledTimes(2)
		} finally {
			vi.useRealTimers()
		}
	})
})
