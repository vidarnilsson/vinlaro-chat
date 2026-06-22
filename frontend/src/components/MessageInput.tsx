import { useState } from 'react'

interface Props {
  onSend: (content: string) => Promise<void>
  disabled: boolean
}

export function MessageInput({ onSend, disabled }: Props) {
  const [text, setText] = useState('')
  const [sending, setSending] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const content = text.trim()
    if (!content || sending) return
    setSending(true)
    try {
      await onSend(content)
      setText('')
    } finally {
      setSending(false)
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e as unknown as React.FormEvent)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      style={{ display: 'flex', gap: 8, padding: '12px 16px', borderTop: '1px solid #ddd' }}
    >
      <textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Message… (Enter to send, Shift+Enter for newline)"
        disabled={disabled || sending}
        rows={1}
        style={{ flex: 1, resize: 'none', padding: '8px 10px', borderRadius: 6, border: '1px solid #ccc', fontSize: 14 }}
      />
      <button type="submit" disabled={disabled || sending || !text.trim()} style={{ padding: '0 16px' }}>
        Send
      </button>
    </form>
  )
}
