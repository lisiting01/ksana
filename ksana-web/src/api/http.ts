import { useSettingsStore } from '@/stores/settings'

export const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:7100'

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public statusText: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const settingsStore = useSettingsStore()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init?.headers || {})
  }

  if (settingsStore.apiKey) {
    headers['Authorization'] = `ApiKey ${settingsStore.apiKey}`
    headers['X-API-Key'] = settingsStore.apiKey
  }

  const res = await fetch(`${settingsStore.apiBase}${path}`, {
    headers,
    ...init,
  })

  const text = await res.text()
  const data = text ? JSON.parse(text) : undefined

  if (!res.ok) {
    let errorMessage = (data && (data.message || data.error)) || res.statusText

    if (res.status === 401 || res.status === 403) {
      errorMessage = `未授权访问：${errorMessage}。请检查API密钥配置。`
    }

    throw new ApiError(errorMessage, res.status, res.statusText)
  }

  return data as T
}