import { useState } from 'react'
import { type Channel, createChannel } from '../api/channels'

interface Props {
  channels: Channel[]
  activeId: string | null
  onSelect: (channel: Channel) => void
  onCreated: (channel: Channel) => void
}

export function ChannelList({ channels, activeId, onSelect, onCreated }: Props) {
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [creating, setCreating] = useState(false)

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setError('')
    setCreating(true)
    try {
      const ch = await createChannel(name.trim(), '')
      setName('')
      onCreated(ch)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create channel')
    } finally {
      setCreating(false)
    }
  }

  return (
    <aside style={{ width: 220, borderRight: '1px solid #ddd', padding: 12, display: 'flex', flexDirection: 'column', gap: 8 }}>
      <strong>Channels</strong>
      <ul style={{ listStyle: 'none', margin: 0, padding: 0, flex: 1, overflowY: 'auto' }}>
        {channels.map((ch) => (
          <li key={ch.id}>
            <button
              onClick={() => onSelect(ch)}
              style={{
                width: '100%',
                textAlign: 'left',
                padding: '6px 8px',
                background: ch.id === activeId ? '#e0e7ff' : 'transparent',
                border: 'none',
                borderRadius: 4,
                cursor: 'pointer',
              }}
            >
              # {ch.name}
            </button>
          </li>
        ))}
      </ul>
      <form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="New channel"
          style={{ padding: '4px 6px', fontSize: 13 }}
        />
        {error && <span style={{ color: 'red', fontSize: 12 }}>{error}</span>}
        <button type="submit" disabled={creating} style={{ padding: '4px 6px' }}>
          {creating ? 'Creating…' : 'Create'}
        </button>
      </form>
    </aside>
  )
}
