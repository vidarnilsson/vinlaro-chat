import { useState, useEffect } from 'react'
import {
  listFriends,
  listFriendRequests,
  sendFriendRequest,
  acceptFriendRequest,
  declineFriendRequest,
  type Friend,
  type FriendRequest,
} from '../api/friends'
import { searchUsers, type UserResult } from '../api/users'
import { getOrCreateDM } from '../api/dm'
import type { DMConversation } from '../api/dm'

interface Props {
  onOpenDM: (channelId: string, otherUsername: string) => void
  onDMCreated: (dm: DMConversation) => void
}

export function FriendsPanel({ onOpenDM, onDMCreated }: Props) {
  const [friends, setFriends] = useState<Friend[]>([])
  const [requests, setRequests] = useState<FriendRequest[]>([])
  const [search, setSearch] = useState('')
  const [results, setResults] = useState<UserResult[]>([])
  const [searching, setSearching] = useState(false)
  const [sentRequests, setSentRequests] = useState<Set<string>>(new Set())

  useEffect(() => {
    listFriends().then(setFriends).catch(() => {})
    listFriendRequests().then(setRequests).catch(() => {})
  }, [])

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

  async function handleSendRequest(user: UserResult) {
    try {
      await sendFriendRequest(user.id)
      setSentRequests((prev) => new Set([...prev, user.id]))
    } catch {
      // ignore duplicate / conflict
    }
  }

  async function handleAccept(req: FriendRequest) {
    try {
      await acceptFriendRequest(req.id)
      setRequests((prev) => prev.filter((r) => r.id !== req.id))
      const newFriend: Friend = {
        friendship_id: req.id,
        friend_id: req.requester_id,
        friend_username: req.requester_username,
        status: 'accepted',
      }
      setFriends((prev) => [...prev, newFriend])
    } catch {
      // ignore
    }
  }

  async function handleDecline(req: FriendRequest) {
    try {
      await declineFriendRequest(req.id)
      setRequests((prev) => prev.filter((r) => r.id !== req.id))
    } catch {
      // ignore
    }
  }

  async function handleMessageFriend(friend: Friend) {
    try {
      const { channel_id } = await getOrCreateDM(friend.friend_id)
      const dm: DMConversation = {
        channel_id,
        other_user_id: friend.friend_id,
        other_username: friend.friend_username,
      }
      onDMCreated(dm)
      onOpenDM(channel_id, friend.friend_username)
    } catch {
      // ignore
    }
  }

  return (
    <div style={{ marginTop: 16 }}>
      <strong style={{ fontSize: 12, textTransform: 'uppercase', color: '#888', letterSpacing: '0.05em' }}>
        Friends
      </strong>

      {requests.length > 0 && (
        <div style={{ margin: '6px 0' }}>
          <div style={{ fontSize: 12, color: '#888', marginBottom: 4 }}>Pending requests</div>
          {requests.map((req) => (
            <div key={req.id} style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 4, fontSize: 13 }}>
              <span style={{ flex: 1 }}>{req.requester_username}</span>
              <button onClick={() => handleAccept(req)} style={{ padding: '2px 6px', fontSize: 12 }}>✓</button>
              <button onClick={() => handleDecline(req)} style={{ padding: '2px 6px', fontSize: 12 }}>✗</button>
            </div>
          ))}
        </div>
      )}

      <ul style={{ listStyle: 'none', margin: '6px 0', padding: 0 }}>
        {friends.map((f) => (
          <li key={f.friendship_id} style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 2 }}>
            <span style={{ flex: 1, fontSize: 13 }}>{f.friend_username}</span>
            <button
              onClick={() => handleMessageFriend(f)}
              style={{ padding: '2px 6px', fontSize: 12 }}
              title="Send message"
            >
              DM
            </button>
          </li>
        ))}
      </ul>

      <form onSubmit={handleSearch} style={{ display: 'flex', flexDirection: 'column', gap: 4, marginTop: 8 }}>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Add friend..."
          style={{ padding: '4px 6px', fontSize: 13 }}
        />
        <button type="submit" disabled={searching} style={{ padding: '4px 6px' }}>
          {searching ? 'Searching…' : 'Search'}
        </button>
      </form>
      {results.length > 0 && (
        <ul style={{ listStyle: 'none', margin: '4px 0', padding: 0, border: '1px solid #ddd', borderRadius: 4 }}>
          {results.map((u) => (
            <li key={u.id} style={{ display: 'flex', alignItems: 'center', padding: '4px 8px' }}>
              <span style={{ flex: 1, fontSize: 13 }}>{u.username}</span>
              {sentRequests.has(u.id) ? (
                <span style={{ fontSize: 12, color: '#888' }}>Request sent</span>
              ) : (
                <button onClick={() => handleSendRequest(u)} style={{ padding: '2px 6px', fontSize: 12 }}>
                  Add
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
