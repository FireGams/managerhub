import { getToken } from './App'

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch('/api/v1' + path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(getToken() ? { Authorization: 'Bearer ' + getToken() } : {}),
      ...(init?.headers || {}),
    },
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || res.statusText)
  }
  const text = await res.text()
  return text ? JSON.parse(text) : ({} as T)
}
