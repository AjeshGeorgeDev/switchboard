const base = ''

let refreshPromise: Promise<boolean> | null = null

function isAuthPath(path: string) {
  return (
    path.startsWith('/api/auth/login') ||
    path.startsWith('/api/auth/refresh') ||
    path.startsWith('/api/auth/logout')
  )
}

async function refreshSession(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      try {
        const res = await fetch(`${base}/api/auth/refresh`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: '{}',
        })
        return res.ok
      } catch {
        return false
      } finally {
        refreshPromise = null
      }
    })()
  }
  return refreshPromise
}

async function request<T>(path: string, options: RequestInit = {}, retried = false): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })
  if (!res.ok) {
    if (res.status === 401 && !retried && !isAuthPath(path)) {
      const refreshed = await refreshSession()
      if (refreshed) return request<T>(path, options, true)
    }
    const text = await res.text()
    let message = text || res.statusText || `Request failed (${res.status})`
    if (text.trimStart().startsWith('{')) {
      try {
        const body = JSON.parse(text)
        if (body?.error) message = body.error
      } catch {
        // keep plain-text message
      }
    }
    throw new Error(message)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) => request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  patch: <T>(path: string, body: unknown) => request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) => request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
  download: async (path: string, fallbackName = 'download') => {
    const res = await fetch(`${base}${path}`, { credentials: 'include' })
    if (!res.ok) {
      if (res.status === 401) {
        const refreshed = await refreshSession()
        if (refreshed) return api.download(path, fallbackName)
      }
      const text = await res.text()
      throw new Error(text || res.statusText || `Download failed (${res.status})`)
    }
    const blob = await res.blob()
    let filename = fallbackName
    const cd = res.headers.get('Content-Disposition') || ''
    const match = /filename="?([^";]+)"?/i.exec(cd)
    if (match?.[1]) filename = match[1]
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  },
}
