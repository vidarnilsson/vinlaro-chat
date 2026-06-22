import { apiFetch } from './client'

export interface UserResponse {
  user_id: string
  username: string
}

export function register(username: string, email: string, password: string): Promise<UserResponse> {
  return apiFetch<UserResponse>('/api/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, email, password }),
  })
}

export function login(email: string, password: string): Promise<UserResponse> {
  return apiFetch<UserResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function logout(): Promise<void> {
  return apiFetch<void>('/api/auth/logout', { method: 'POST' })
}

export function getMe(): Promise<UserResponse> {
  return apiFetch<UserResponse>('/api/auth/me')
}
