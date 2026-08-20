import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { api, clearSession, readSession, toSession, writeSession } from './api.js';
import { ErrorBoundary } from './ErrorBoundary.jsx';
import { getMessages, nextLanguage } from './i18n.js';
import { connectTinode } from './tinode.js';

const defaultForm = { login: '', username: '', email: '', password: '', displayName: '' };

function App() {
  return <ErrorBoundary><AppShell /></ErrorBoundary>;
}

function AppShell() {
  const [language, setLanguage] = useState(() => localStorage.getItem('bashocode.language') || 'id');
  const t = useMemo(() => getMessages(language), [language]);
  const [session, setSession] = useState(readSession);
  const [tinode, setTinode] = useState(null);
  const [connection, setConnection] = useState('disconnected');
  const [online, setOnline] = useState(() => navigator.onLine);
  const [bootError, setBootError] = useState('');
  const sessionRef = useRef(session);
  const tinodeRef = useRef(null);
  const tinodeAttemptRef = useRef(0);
  const establishingRef = useRef(null);

  useEffect(() => {
    sessionRef.current = session;
  }, [session]);

  useEffect(() => {
    localStorage.setItem('bashocode.language', language);
  }, [language]);

  useEffect(() => {
    const onOnline = () => setOnline(true);
    const onOffline = () => setOnline(false);
    window.addEventListener('online', onOnline);
    window.addEventListener('offline', onOffline);
    return () => {
      window.removeEventListener('online', onOnline);
      window.removeEventListener('offline', onOffline);
    };
  }, []);

  const establishTinode = useCallback((nextSession) => {
    const accessToken = nextSession.session?.access_token;
    if (!accessToken) return Promise.reject(new Error('App session is missing'));
    if (establishingRef.current?.accessToken === accessToken) return establishingRef.current.promise;

    const attempt = tinodeAttemptRef.current + 1;
    tinodeAttemptRef.current = attempt;
    const isCurrent = () => tinodeAttemptRef.current === attempt
      && sessionRef.current?.session?.access_token === accessToken;
    setConnection('connecting');
    setBootError('');
    const promise = (async () => {
      try {
        const result = await connectTinode({
          username: nextSession.tinodeLogin.username,
          password: nextSession.__password,
          token: nextSession.tinodeAuth?.token,
          expires: nextSession.tinodeAuth?.expires,
        }, {
          onConnect: () => { if (isCurrent()) setConnection('connected'); },
          onDisconnect: () => { if (isCurrent()) setConnection('disconnected'); },
        });
        if (!isCurrent()) {
          result.client.disconnect();
          return null;
        }
        const authToken = result.authToken ? { token: result.authToken.token, expires: result.authToken.expires.toISOString() } : null;
        const persisted = { user: nextSession.user, session: nextSession.session, tinodeLogin: nextSession.tinodeLogin, tinodeAuth: authToken };
        writeSession(persisted);
        tinodeRef.current = result.client;
        setSession(persisted);
        setTinode(result.client);
        setConnection('connected');
        return result.client;
      } catch (error) {
        if (!isCurrent()) return null;
        setConnection('disconnected');
        setBootError(error.message || 'Tinode connection failed');
        if (!nextSession.__password) {
          clearSession();
          sessionRef.current = null;
          setSession(null);
        }
        throw error;
      }
    })();
    establishingRef.current = { accessToken, promise };
    promise.then(() => {
      if (establishingRef.current?.promise === promise) establishingRef.current = null;
    }, () => {
      if (establishingRef.current?.promise === promise) establishingRef.current = null;
    });
    return promise;
  }, []);

  const disconnectTinode = useCallback(() => {
    tinodeAttemptRef.current += 1;
    establishingRef.current = null;
    tinodeRef.current?.disconnect();
    tinodeRef.current = null;
    setTinode(null);
    setConnection('disconnected');
  }, []);

  useEffect(() => {
    if (!session?.session?.access_token) return undefined;
    let active = true;
    const accessToken = session.session.access_token;
    let refreshTimer;
    const scheduleRefresh = (expiresAt) => {
      const delay = Math.max(1000, new Date(expiresAt).getTime() - Date.now() - 30_000);
      refreshTimer = window.setTimeout(refresh, delay);
    };
    const refresh = async () => {
      try {
        const refreshed = await api.refresh(session.session.refresh_token);
        const persisted = { ...session, ...toSession(refreshed), tinodeAuth: session.tinodeAuth };
        delete persisted.__password;
        if (active && sessionRef.current?.session?.access_token === accessToken) {
          writeSession(persisted);
          sessionRef.current = persisted;
          setSession(persisted);
        }
      } catch {
        if (active && sessionRef.current?.session?.access_token === accessToken) {
          clearSession();
          sessionRef.current = null;
          disconnectTinode();
          setSession(null);
        }
      }
    };
    (async () => {
      try {
        await api.me(session.session.access_token);
        if (active && sessionRef.current?.session?.access_token === accessToken) scheduleRefresh(session.session.access_expires_at);
      } catch (error) {
        if (error.status !== 401 || !session.session.refresh_token) {
          if (active && sessionRef.current?.session?.access_token === accessToken) {
            clearSession();
            sessionRef.current = null;
            disconnectTinode();
            setSession(null);
          }
          return;
        }
        await refresh();
      }
    })();
    return () => { active = false; window.clearTimeout(refreshTimer); };
  }, [session?.session?.access_token, disconnectTinode]);

  useEffect(() => {
    if (!session || tinode || (!session.__password && !session.tinodeAuth)) return undefined;
    establishTinode(session).catch(() => {});
    return undefined;
  }, [session, tinode, establishTinode]);

  const finishAuth = async (data, password) => {
    const next = toSession(data);
    next.__password = password;
    sessionRef.current = next;
    await establishTinode(next);
  };

  const handleAuth = async (mode, form) => {
    const data = mode === 'register'
      ? await api.register({ username: form.username, email: form.email, password: form.password, display_name: form.displayName })
      : await api.login({ login: form.login, password: form.password });
    await finishAuth(data, form.password);
  };

  const handleLogout = async () => {
    const currentSession = sessionRef.current;
    const accessToken = currentSession?.session?.access_token;
    clearSession();
    sessionRef.current = null;
    disconnectTinode();
    setSession(null);
    if (accessToken) {
      try {
        await api.logout(accessToken);
      } catch (error) {
        if (error.status === 401 && currentSession.session.refresh_token) {
          try {
            const refreshed = await api.refresh(currentSession.session.refresh_token);
            await api.logout(refreshed.session.access_token);
          } catch {
            // Local state is already cleared; the server session may expire naturally.
          }
        }
      }
    }
  };

  if (!session) {
    return <AuthPage t={t} language={language} onLanguageChange={() => setLanguage(nextLanguage(language))} onSubmit={handleAuth} />;
  }

  return (
    <Workspace
      t={t}
      language={language}
      online={online}
      connection={connection}
      user={session.user}
      bootError={bootError}
      onLanguageChange={() => setLanguage(nextLanguage(language))}
      onLogout={handleLogout}
    />
  );
}

function AuthPage({ t, language, onLanguageChange, onSubmit }) {
  const [mode, setMode] = useState('login');
  const [form, setForm] = useState(defaultForm);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const update = (key) => (event) => setForm((current) => ({ ...current, [key]: event.target.value }));
  const submit = async (event) => {
    event.preventDefault();
    setBusy(true);
    setError('');
    try {
      await onSubmit(mode, form);
    } catch (submitError) {
      setError(submitError.message || 'Request failed');
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="auth-page">
      <section className="auth-visual">
        <div className="orb orb-one" />
        <div className="orb orb-two" />
        <div className="visual-copy">
          <span className="eyebrow">TINODE · SELF-HOSTED</span>
          <h1>{t.brand}</h1>
          <p>{t.tagline}</p>
          <div className="feature-pill"><span className="status-dot" /> {t.privateByDefault}</div>
        </div>
      </section>
      <section className="auth-panel">
        <div className="language-row"><span>{t.language}</span><button type="button" className="text-button" onClick={onLanguageChange}>{language.toUpperCase()}</button></div>
        <div className="auth-card">
          <div className="brand-mark">B</div>
          <span className="eyebrow">{mode === 'login' ? t.signIn : t.signUp}</span>
          <h2>{mode === 'login' ? t.submitSignIn : t.submitSignUp}</h2>
          <p className="muted">{mode === 'login' ? t.haveAccount : t.noAccount} <button type="button" className="text-button" onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setError(''); }}>{mode === 'login' ? t.switchToSignUp : t.switchToSignIn}</button></p>
          <form onSubmit={submit}>
            {mode === 'register' ? <>
              <label>{t.username}<input required minLength="4" autoComplete="username" value={form.username} onChange={update('username')} /></label>
              <label>{t.displayName}<input required maxLength="80" value={form.displayName} onChange={update('displayName')} /></label>
              <label>{t.email}<input required type="email" autoComplete="email" value={form.email} onChange={update('email')} /></label>
            </> : <label>{t.usernameOrEmail}<input required autoComplete="username" value={form.login} onChange={update('login')} /> </label>}
            <label>{t.password}<input required minLength="12" type="password" autoComplete={mode === 'login' ? 'current-password' : 'new-password'} value={form.password} onChange={update('password')} /></label>
            {error && <div className="form-error" role="alert">{error}</div>}
            <button className="primary-button" disabled={busy} type="submit">{busy ? t.connecting : (mode === 'login' ? t.signIn : t.signUp)} <span>→</span></button>
          </form>
        </div>
        <p className="legal-note">{t.legal}</p>
      </section>
    </main>
  );
}

function Workspace({ t, language, online, connection, user, bootError, onLanguageChange, onLogout }) {
  const connectionText = connection === 'connected' ? t.connected : connection === 'connecting' ? t.connecting : t.disconnected;
  return (
    <main className="workspace-page">
      <header className="topbar">
        <div className="brand-lockup"><div className="brand-mark small">B</div><div><strong>{t.brand}</strong><span>{t.workspace}</span></div></div>
        <div className="topbar-actions"><span className={`connection-badge ${connection}`}><span className="status-dot" />{online ? connectionText : t.offline}</span><button type="button" className="text-button" onClick={onLanguageChange}>{language.toUpperCase()}</button><button type="button" className="outline-button" onClick={onLogout}>{t.logout}</button></div>
      </header>
      <section className="workspace-grid">
        <aside className="sidebar-card"><div className="profile-block"><div className="avatar">{(user.display_name || user.username).slice(0, 1).toUpperCase()}</div><div><strong>{user.display_name || user.username}</strong><span>@{user.username}</span></div></div><div className="sidebar-empty"><span className="empty-icon">✦</span><strong>{t.readyForChat}</strong><p>{t.readyForChatText}</p></div></aside>
        <section className="welcome-card"><div className="welcome-glow" /><span className="eyebrow">M5 · WEB FOUNDATION</span><h1>{t.workspace}</h1><p>{t.privateByDefaultText}</p>{bootError && <div className="form-error" role="alert">{bootError}</div>}<div className="status-panel"><div><span className="muted">Tinode</span><strong>{connectionText}</strong></div><div><span className="muted">App session</span><strong>{t.online}</strong></div></div><p className="muted small-copy">{t.nextMilestone}</p></section>
      </section>
    </main>
  );
}

export default App;
