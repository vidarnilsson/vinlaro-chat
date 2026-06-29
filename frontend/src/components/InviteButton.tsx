import { useState } from 'react'
import { sendInvite } from '../api/invites'
import { searchUsers, type UserResult } from '../api/users'

interface Props {
  channelId: string
}

export function InviteButton({ channelId }: Props) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [results, setResults] = useState<UserResult[]>([])
  const [sent, setSent] = useState<Set<string>>(new Set())
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

  async function handleInvite(user: UserResult) {
    try {
      await sendInvite(channelId, user.id)
      setSent((prev) => new Set([...prev, user.id]))
    } catch {
      // ignore duplicate/conflict
    }
  }

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        style={{ padding: '3px 10px', fontSize: 12, cursor: 'pointer' }}
      >
        Invite
      </button>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <form onSubmit={handleSearch} style={{ display: 'flex', gap: 4 }}>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search users..."
          style={{ padding: '3px 6px', fontSize: 12, flex: 1 }}
          autoFocus
        />
        <button type="submit" disabled={searching} style={{ padding: '3px 8px', fontSize: 12 }}>
          {searching ? '…' : 'Go'}
        </button>
        <button type="button" onClick={() => { setOpen(false); setResults([]); setSearch('') }} style={{ padding: '3px 8px', fontSize: 12 }}>
          ✕
        </button>
      </form>
      {results.length > 0 && (
        <ul style={{ listStyle: 'none', margin: 0, padding: 0, border: '1px solid #ddd', borderRadius: 4 }}>
          {results.map((u) => (
            <li key={u.id} style={{ display: 'flex', alignItems: 'center', padding: '4px 8px' }}>
              <span style={{ flex: 1, fontSize: 12 }}>{u.username}</span>
              {sent.has(u.id) ? (
                <span style={{ fontSize: 11, color: '#888' }}>Invited</span>
              ) : (
                <button onClick={() => handleInvite(u)} style={{ padding: '2px 6px', fontSize: 11 }}>
                  Invite
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
