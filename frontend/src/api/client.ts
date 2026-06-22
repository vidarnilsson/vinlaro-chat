let authToken: string | null = null

export function setToken(token: string | null) {
  authToken = token
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string, string>),
  }

  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`
  }

  const res = await fetch(path, { ...init, headers })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error ?? res.statusText)
  }

  // 202 Accepted with a body, 204 No Content without
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
