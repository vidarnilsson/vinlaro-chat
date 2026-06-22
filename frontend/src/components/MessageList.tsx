import { useEffect, useRef } from 'react'
import type { Message } from '../api/messages'

interface Props {
  messages: Message[]
  currentUserId: string
}

export function MessageList({ messages, currentUserId }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length])

  return (
    <div style={{ flex: 1, overflowY: 'auto', padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
      {messages.map((msg) => {
        const isOwn = msg.user_id === currentUserId
        return (
          <div key={msg.id} style={{ display: 'flex', flexDirection: 'column', alignItems: isOwn ? 'flex-end' : 'flex-start' }}>
            <span style={{ fontSize: 11, color: '#888', marginBottom: 2 }}>
              {msg.username} · {new Date(msg.created_at).toLocaleTimeString()}
            </span>
            <div
              style={{
                background: isOwn ? '#6366f1' : '#f1f5f9',
                color: isOwn ? '#fff' : '#000',
                borderRadius: 8,
                padding: '6px 12px',
                maxWidth: '70%',
                wordBreak: 'break-word',
              }}
            >
              {msg.content}
            </div>
          </div>
        )
      })}
      <div ref={bottomRef} />
    </div>
  )
}
