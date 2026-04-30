import type { SSEEvent, SSEEventKind } from './types'

type Listener<T = unknown> = (ev: SSEEvent<T>) => void

export interface SSEHandle {
	close(): void
}

const KINDS: SSEEventKind[] = ['status', 'progress', 'log']

/**
 * Subscribe to a Tidal SSE endpoint. Returns a handle that can be closed to
 * stop the stream.
 *
 * Auto-reconnects with exponential backoff (capped at 10s).
 */
export function streamEvents(url: string, listeners: Partial<Record<SSEEventKind, Listener>>): SSEHandle {
	let es: EventSource | null = null
	let closed = false
	let backoff = 500

	const open = () => {
		es = new EventSource(url, { withCredentials: false })

		for (const kind of KINDS) {
			const fn = listeners[kind]
			if (!fn) continue
			es.addEventListener(kind, (e) => {
				try {
					fn(JSON.parse((e as MessageEvent).data))
					backoff = 500
				} catch (err) {
					console.warn('sse parse error', err)
				}
			})
		}

		es.onerror = () => {
			if (closed) return
			es?.close()
			setTimeout(open, backoff)
			backoff = Math.min(backoff * 2, 10_000)
		}
	}

	open()

	return {
		close() {
			closed = true
			es?.close()
		}
	}
}
