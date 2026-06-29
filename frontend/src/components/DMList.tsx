import { useState } from 'react'
import { type DMConversation, getOrCreateDM } from '../api/dm'
import { searchUsers, type UserResult } from '../api/users'

interface Props {
  dms: DMConversation[]
  activeChannelId: string | null
  onOpenDM: (channelId: string, otherUsername: string) => void
  onDMCreated: (dm: DMConversation) => void
}

export function DMList({ dms, activeChannelId, onOpenDM, onDMCreated }: Props) {
  const [search, setSearch] = useState('')
  const [results, setResults] = useState<UserResult[]>([])
  const [searching, setSearching] = useState(false)

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    if (!search.trim()) return
    setSearching(true)
    try {
      const users = await searchUsers(search.trim())
      setResults(users)
    } catch {
      setResults([])
    } finally {
      setSearching(false)
    }
  }

  async function handleStartDM(user: UserResult) {
    try {
      const { channel_id } = await getOrCreateDM(user.id)
      const existing = dms.find((d) => d.channel_id === channel_id)
      if (existing) {
        onOpenDM(channel_id, existing.other_username)
      } else {
        const newDM: DMConversation = {
          channel_id,
          other_user_id: user.id,
          other_username: user.username,
        }
        onDMCreated(newDM)
        onOpenDM(channel_id, user.username)
      }
      setSearch('')
      setResults([])
    } catch {
      // ignore
    }
  }

  return (
    <div style={{ marginTop: 16 }}>
      <strong style={{ fontSize: 12, textTransform: 'uppercase', color: '#888', letterSpacing: '0.05em' }}>
        Direct Messages
      </strong>
      <ul style={{ listStyle: 'none', margin: '6px 0', padding: 0 }}>
        {dms.map((dm) => (
          <li key={dm.channel_id}>
            <button
              onClick={() => onOpenDM(dm.channel_id, dm.other_username)}
              style={{
                width: '100%',
                textAlign: 'left',
                padding: '5px 8px',
                background: dm.channel_id === activeChannelId ? '#e0e7ff' : 'transparent',
                border: 'none',
                borderRadius: 4,
                cursor: 'pointer',
                fontSize: 14,
              }}
            >
              @ {dm.other_username}
            </button>
          </li>
        ))}
      </ul>
      <form onSubmit={handleSearch} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Find user..."
          style={{ padding: '4px 6px', fontSize: 13 }}
        />
        <button type="submit" disabled={searching} style={{ padding: '4px 6px' }}>
          {searching ? 'Searching…' : 'Search'}
        </button>
      </form>
      {results.length > 0 && (
        <ul style={{ listStyle: 'none', margin: '4px 0', padding: 0, border: '1px solid #ddd', borderRadius: 4 }}>
          {results.map((u) => (
            <li key={u.id}>
              <button
                onClick={() => handleStartDM(u)}
                style={{
                  width: '100%',
                  textAlign: 'left',
                  padding: '5px 8px',
                  background: 'transparent',
                  border: 'none',
                  cursor: 'pointer',
                  fontSize: 13,
                }}
              >
                {u.username} — Message
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
