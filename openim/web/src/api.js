const API_URL = import.meta.env.VITE_APP_API_URL || 'http://127.0.0.1:8080'

async function request(path, options = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(body?.error?.message || `Request failed (${response.status})`)
  }
  return body.data
}

export const authApi = {
  login: (login, password) => request('/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ login, password }),
  }),
  register: (username, email, password, displayName) => request('/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, email, password, display_name: displayName }),
  }),
  session: (accessToken) => request('/v1/openim/session', {
    headers: { Authorization: `Bearer ${accessToken}` },
  }),
  searchUsers: (accessToken, query) => request(`/v1/users/search?q=${encodeURIComponent(query)}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  }),
}
