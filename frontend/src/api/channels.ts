import { apiFetch } from './client'

export interface Channel {
  id: string
  name: string
  description: string | null
  created_at: string
}

export function listChannels(): Promise<Channel[]> {
  return apiFetch<Channel[]>('/api/channels')
}

export function createChannel(name: string, description: string): Promise<Channel> {
  return apiFetch<Channel>('/api/channels', {
    method: 'POST',
    body: JSON.stringify({ name, description }),
  })
}
