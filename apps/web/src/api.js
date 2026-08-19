const API_BASE = (import.meta.env.VITE_APP_API_BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const SESSION_KEY = 'bashocode.app.session';

export function readSession() {
  try {
    const raw = localStorage.getItem(SESSION_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function writeSession(session) {
  localStorage.setItem(SESSION_KEY, JSON.stringify(session));
}

export function clearSession() {
  localStorage.removeItem(SESSION_KEY);
}

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const error = new Error(body?.error?.message || 'Request failed');
    error.status = response.status;
    error.code = body?.error?.code;
    error.fields = body?.error?.fields;
    throw error;
  }
  return body.data;
}

export const api = {
  register(input) {
    return request('/v1/auth/register', { method: 'POST', body: JSON.stringify(input) });
  },
  login(input) {
    return request('/v1/auth/login', { method: 'POST', body: JSON.stringify(input) });
  },
  refresh(refreshToken) {
    return request('/v1/auth/refresh', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) });
  },
  me(accessToken) {
    return request('/v1/me', { headers: { Authorization: `Bearer ${accessToken}` } });
  },
  logout(accessToken) {
    return request('/v1/auth/logout', { method: 'POST', headers: { Authorization: `Bearer ${accessToken}` } });
  },
};

export function toSession(data) {
  return {
    user: data.user,
    session: data.session,
    tinodeLogin: data.tinode_login || { scheme: 'basic', username: data.user.username },
  };
}
