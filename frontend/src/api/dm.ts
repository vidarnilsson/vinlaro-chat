import { apiFetch } from './client'

export interface DMConversation {
  channel_id: string
  other_user_id: string
  other_username: string
}

export function listDMs(): Promise<DMConversation[]> {
  return apiFetch<DMConversation[]>('/api/dm')
}

export function getOrCreateDM(targetUserID: string): Promise<{ channel_id: string }> {
  return apiFetch<{ channel_id: string }>(`/api/dm/${targetUserID}`, { method: 'POST' })
}
