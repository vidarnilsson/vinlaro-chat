import { apiFetch } from './client'

export interface Friend {
  friendship_id: string
  friend_id: string
  friend_username: string
  status: string
}

export interface FriendRequest {
  id: string
  requester_id: string
  requester_username: string
  created_at: string
}

export function listFriends(): Promise<Friend[]> {
  return apiFetch<Friend[]>('/api/friends')
}

export function listFriendRequests(): Promise<FriendRequest[]> {
  return apiFetch<FriendRequest[]>('/api/friends/requests')
}

export function sendFriendRequest(userID: string): Promise<unknown> {
  return apiFetch(`/api/friends/request/${userID}`, { method: 'POST' })
}

export function acceptFriendRequest(friendshipID: string): Promise<unknown> {
  return apiFetch(`/api/friends/accept/${friendshipID}`, { method: 'POST' })
}

export function declineFriendRequest(friendshipID: string): Promise<unknown> {
  return apiFetch(`/api/friends/decline/${friendshipID}`, { method: 'POST' })
}
