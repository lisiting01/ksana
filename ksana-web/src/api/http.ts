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
  const res = await fetch(`${apiBase}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {})
    },
    ...init,
  })

  const text = await res.text()
  const data = text ? JSON.parse(text) : undefined

  if (!res.ok) {
    const errorMessage = (data && (data.message || data.error)) || res.statusText
    throw new ApiError(errorMessage, res.status, res.statusText)
  }

  return data as T
}