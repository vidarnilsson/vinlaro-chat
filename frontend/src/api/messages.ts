import { apiFetch } from './client'

export interface Message {
  id: string
  channel_id: string
  user_id: string
  username: string
  content: string
  created_at: string
}

export function getMessages(channelId: string, limit = 50): Promise<Message[]> {
  return apiFetch<Message[]>(`/api/channels/${channelId}/messages?limit=${limit}`)
}

export function sendMessage(channelId: string, content: string): Promise<Message> {
  return apiFetch<Message>(`/api/channels/${channelId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  })
}
