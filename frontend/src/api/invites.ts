import { apiFetch } from './client'

export interface PendingInvite {
  id: string
  channel_id: string
  channel_name: string
  inviter_username: string
  created_at: string
}

export function listPendingInvites(): Promise<PendingInvite[]> {
  return apiFetch<PendingInvite[]>('/api/invites')
}

export function sendInvite(channelID: string, userID: string): Promise<unknown> {
  return apiFetch(`/api/channels/${channelID}/invite/${userID}`, { method: 'POST' })
}

export function acceptInvite(inviteID: string): Promise<unknown> {
  return apiFetch(`/api/invites/${inviteID}/accept`, { method: 'POST' })
}

export function declineInvite(inviteID: string): Promise<unknown> {
  return apiFetch(`/api/invites/${inviteID}/decline`, { method: 'POST' })
}
