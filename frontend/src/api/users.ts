import { apiFetch } from './client'

export interface UserResult {
  id: string
  username: string
}

export function searchUsers(query: string): Promise<UserResult[]> {
  return apiFetch<UserResult[]>(`/api/users?search=${encodeURIComponent(query)}`)
}
