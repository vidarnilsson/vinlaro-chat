import { useState, useEffect } from 'react'
import {
  listPendingInvites,
  acceptInvite,
  declineInvite,
  type PendingInvite,
} from '../api/invites'
import type { Channel } from '../api/channels'

interface Props {
  onChannelJoined: (channel: Channel) => void
}

export function InvitePanel({ onChannelJoined }: Props) {
  const [invites, setInvites] = useState<PendingInvite[]>([])

  useEffect(() => {
    listPendingInvites().then(setInvites).catch(() => {})
  }, [])

  if (invites.length === 0) return null

  async function handleAccept(invite: PendingInvite) {
    try {
      await acceptInvite(invite.id)
      setInvites((prev) => prev.filter((i) => i.id !== invite.id))
      // Synthesise a Channel object so the parent can add it to the channel list.
      const joined: Channel = {
        id: invite.channel_id,
        name: invite.channel_name,
        description: null,
        kind: 'private',
        created_at: new Date().toISOString(),
      }
      onChannelJoined(joined)
    } catch {
      // ignore
    }
  }

  async function handleDecline(invite: PendingInvite) {
    try {
      await declineInvite(invite.id)
      setInvites((prev) => prev.filter((i) => i.id !== invite.id))
    } catch {
      // ignore
    }
  }

  return (
    <div style={{ marginTop: 16 }}>
      <strong style={{ fontSize: 12, textTransform: 'uppercase', color: '#888', letterSpacing: '0.05em' }}>
        Invites ({invites.length})
      </strong>
      {invites.map((inv) => (
        <div key={inv.id} style={{ display: 'flex', alignItems: 'center', gap: 4, marginTop: 6, fontSize: 13 }}>
          <span style={{ flex: 1 }}>#{inv.channel_name} from {inv.inviter_username}</span>
          <button onClick={() => handleAccept(inv)} style={{ padding: '2px 6px', fontSize: 12 }}>✓</button>
          <button onClick={() => handleDecline(inv)} style={{ padding: '2px 6px', fontSize: 12 }}>✗</button>
        </div>
      ))}
    </div>
  )
}
