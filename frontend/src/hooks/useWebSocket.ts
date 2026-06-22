import { useEffect, useRef, useCallback } from 'react'
import type { Message } from '../api/messages'

const MAX_RETRIES = 3
const BASE_DELAY_MS = 1000

interface Options {
  channelId: string | null
  onMessage: (msg: Message) => void
}

export function useWebSocket({ channelId, onMessage }: Options) {
  const wsRef = useRef<WebSocket | null>(null)
  const retriesRef = useRef(0)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  const connect = useCallback(() => {
    if (!channelId) return

    // The session cookie is sent automatically by the browser.
    const ws = new WebSocket(`/ws/channels/${channelId}`)
    wsRef.current = ws

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as Message
        onMessageRef.current(msg)
      } catch {
        // ignore malformed frames
      }
    }

    ws.onopen = () => {
      retriesRef.current = 0
    }

    ws.onclose = () => {
      if (retriesRef.current < MAX_RETRIES) {
        const delay = BASE_DELAY_MS * 2 ** retriesRef.current
        retriesRef.current++
        setTimeout(connect, delay)
      }
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [channelId])

  useEffect(() => {
    retriesRef.current = 0
    connect()

    return () => {
      retriesRef.current = MAX_RETRIES
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [connect])
}
