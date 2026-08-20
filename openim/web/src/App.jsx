import { useEffect, useMemo, useRef, useState } from 'react'
import { CbEvents, getSDK } from '@openim/wasm-client-sdk'
import { authApi } from './api'

const API_ADDR = import.meta.env.VITE_OPENIM_API_URL || 'http://127.0.0.1:10002'
const WS_ADDR = import.meta.env.VITE_OPENIM_WS_URL || 'ws://127.0.0.1:10001'
const PLATFORM_ID = Number(import.meta.env.VITE_OPENIM_PLATFORM_ID || 5)

const sdk = getSDK({ coreWasmPath: '/openIM.wasm', sqlWasmPath: '/sql-wasm.wasm', debug: false })

export default function App() {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({ login: '', email: '', password: '', displayName: '' })
  const [session, setSession] = useState(null)
  const [status, setStatus] = useState('Not connected')
  const [error, setError] = useState('')
  const [recipient, setRecipient] = useState('')
  const [userQuery, setUserQuery] = useState('')
  const [userResults, setUserResults] = useState([])
  const [draft, setDraft] = useState('')
  const [messages, setMessages] = useState([])
  const [conversations, setConversations] = useState([])
  const [activeChat, setActiveChat] = useState(null)
  const [isLoadingMessages, setIsLoadingMessages] = useState(false)
  const activeChatRef = useRef(null)
  const currentUserIDRef = useRef('')
  const chatLoadRef = useRef(0)

  const currentUser = useMemo(() => session?.user, [session])

  useEffect(() => {
    const onConnectSuccess = () => setStatus('Connected to OpenIM')
    const onConnecting = () => setStatus('Connecting…')
    const onConnectFailed = (event) => setStatus(`Connection failed: ${event?.errMsg || 'unknown error'}`)
    const onMessages = ({ data = [] }) => {
      const selectedUserID = activeChatRef.current?.userID
      const currentUserID = currentUserIDRef.current
      if (!selectedUserID || !currentUserID) return
      const selectedMessages = data
        .filter((message) => belongsToDirectChat(message, selectedUserID, currentUserID))
        .map(toMessage)
      if (selectedMessages.length > 0) {
        setMessages((current) => [...current, ...selectedMessages])
      }
      refreshConversations()
    }
    sdk.on(CbEvents.OnConnecting, onConnecting)
    sdk.on(CbEvents.OnConnectSuccess, onConnectSuccess)
    sdk.on(CbEvents.OnConnectFailed, onConnectFailed)
    sdk.on(CbEvents.OnRecvNewMessages, onMessages)
    return () => {
      sdk.off(CbEvents.OnConnecting, onConnecting)
      sdk.off(CbEvents.OnConnectSuccess, onConnectSuccess)
      sdk.off(CbEvents.OnConnectFailed, onConnectFailed)
      sdk.off(CbEvents.OnRecvNewMessages, onMessages)
    }
  }, [])

  async function connect(nextSession) {
    setError('')
    setStatus('Connecting…')
    currentUserIDRef.current = nextSession.user.openim_user_id
    await sdk.login({
      userID: nextSession.user.openim_user_id,
      token: nextSession.session.token,
      platformID: PLATFORM_ID,
      apiAddr: nextSession.session.api_addr || API_ADDR,
      wsAddr: nextSession.session.ws_addr || WS_ADDR,
    })
    setSession(nextSession)
    await refreshConversations()
  }

  async function submit(event) {
    event.preventDefault()
    setError('')
    try {
      const result = mode === 'login'
        ? await authApi.login(form.login, form.password)
        : await authApi.register(form.login, form.email, form.password, form.displayName)
      const openimSession = await authApi.session(result.session.access_token)
      await connect({ user: result.user, session: openimSession, appAccessToken: result.session.access_token })
    } catch (cause) {
      setError(cause.message)
      setStatus('Not connected')
    }
  }

  async function refreshConversations() {
    try {
      const result = await sdk.getAllConversationList()
      setConversations(result?.data || [])
    } catch (cause) {
      setError(`Could not load conversations: ${cause.message || cause}`)
    }
  }

  async function searchUsers(event) {
    event.preventDefault()
    if (userQuery.trim().length < 2) return
    try {
      const result = await authApi.searchUsers(session.appAccessToken, userQuery.trim())
      setUserResults(result.users || [])
    } catch (cause) {
      setError(`User search failed: ${cause.message || cause}`)
    }
  }

  async function openDirectChat(user) {
    const userID = user.openim_user_id || user.userID
    if (!userID) return

    const requestID = chatLoadRef.current + 1
    chatLoadRef.current = requestID
    const nextChat = {
      userID,
      username: user.username || '',
      displayName: user.display_name || user.showName || user.username || userID,
      conversationID: user.conversationID || '',
    }
    activeChatRef.current = nextChat
    setActiveChat(nextChat)
    setRecipient(userID)
    setUserQuery(user.username || '')
    setUserResults([])
    setMessages([])
    setError('')
    setIsLoadingMessages(true)

    try {
      const conversationResult = nextChat.conversationID
        ? { data: user }
        : await sdk.getOneConversation({ sourceID: userID, sessionType: 1 })
      const conversation = conversationResult?.data
      if (!conversation?.conversationID) {
        throw new Error('Conversation ID was not returned')
      }

      const historyResult = await sdk.getAdvancedHistoryMessageList({
        conversationID: conversation.conversationID,
        startClientMsgID: '',
        count: 50,
        viewType: 0,
      })
      if (chatLoadRef.current !== requestID || activeChatRef.current?.userID !== userID) return

      const resolvedChat = {
        ...nextChat,
        conversationID: conversation.conversationID,
        displayName: nextChat.displayName || conversation.showName || userID,
      }
      activeChatRef.current = resolvedChat
      setActiveChat(resolvedChat)
      setMessages((historyResult?.data?.messageList || []).map(toMessage).sort(sortMessages))
    } catch (cause) {
      if (chatLoadRef.current === requestID) {
        setMessages([])
        setError(`Could not open conversation: ${cause.message || cause}`)
      }
    } finally {
      if (chatLoadRef.current === requestID) setIsLoadingMessages(false)
    }
  }

  async function sendMessage(event) {
    event.preventDefault()
    const sendTo = activeChat?.userID || recipient.trim()
    if (!sendTo || !draft.trim()) return
    try {
      const created = await sdk.createTextMessage(draft.trim())
      await sdk.sendMessage({ recvID: sendTo, groupID: '', message: created.data })
      setMessages((current) => [...current, { text: draft.trim(), sendID: currentUser.openim_user_id, recvID: sendTo, self: true }])
      setDraft('')
      await refreshConversations()
    } catch (cause) {
      setError(`Message failed to send: ${cause.message || cause}`)
    }
  }

  if (!session) {
    return <main className="shell auth-shell">
      <section className="brand-panel">
        <p className="eyebrow">OPENIM CHAT / M3</p>
        <h1>A chat experience you can make your own.</h1>
        <p>Identity and sessions are managed by app-api; messages run through OpenIM.</p>
      </section>
      <section className="card auth-card">
        <div className="tabs">
          <button className={mode === 'login' ? 'active' : ''} onClick={() => setMode('login')}>Sign in</button>
          <button className={mode === 'register' ? 'active' : ''} onClick={() => setMode('register')}>Create account</button>
        </div>
        <form onSubmit={submit}>
          <label>{mode === 'login' ? 'Username or email' : 'Username'}<input required value={form.login} onChange={(event) => setForm({ ...form, login: event.target.value })} /></label>
          {mode === 'register' && <>
            <label>Email<input required type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} /></label>
            <label>Display name<input required value={form.displayName} onChange={(event) => setForm({ ...form, displayName: event.target.value })} /></label>
          </>}
          <label>Password<input required type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} /></label>
          <button className="primary" type="submit">{mode === 'login' ? 'Sign in to chat' : 'Create account'}</button>
        </form>
        {error && <p className="error">{error}</p>}
      </section>
    </main>
  }

  return <main className="shell app-shell">
    <header className="topbar"><div><p className="eyebrow">OPENIM CHAT / M3</p><h1>Messages</h1></div><span className="status"><i />{status}</span></header>
    <section className="workspace">
      <aside className="card conversations"><div className="section-heading"><h2>Conversations</h2><button type="button" onClick={refreshConversations}>↻</button></div>{conversations.length === 0 ? <p className="muted">No conversations yet.</p> : conversations.map((conversation) => <button className={`conversation ${activeChat?.conversationID === conversation.conversationID ? 'active' : ''}`} type="button" key={conversation.conversationID} onClick={() => openDirectChat(conversation)}><strong>{conversation.showName || conversation.userID || 'Conversation'}</strong><span>{conversationPreview(conversation.latestMsg)}</span></button>)}</aside>
      <section className="card chat"><div className="chat-heading"><div><p className="eyebrow">{activeChat ? 'CHAT WITH' : 'LOGGED IN AS'}</p><h2>{activeChat?.displayName || currentUser.username}</h2></div><span className="pill">{activeChat?.username ? `@${activeChat.username}` : activeChat?.userID || currentUser.openim_user_id}</span></div><form className="user-search" onSubmit={searchUsers}><input placeholder="Search username or name" value={userQuery} onChange={(event) => setUserQuery(event.target.value)} /><button type="submit">Search</button></form>{userResults.length > 0 && <div className="user-results">{userResults.map((user) => <button type="button" key={user.id} onClick={() => openDirectChat(user)}><strong>{user.display_name}</strong><span>@{user.username}</span></button>)}</div>}<div className="messages">{isLoadingMessages ? <p className="muted">Loading messages…</p> : !activeChat && messages.length === 0 ? <p className="muted">Select a user to start chatting.</p> : activeChat && messages.length === 0 ? <p className="muted">Send the first message to get started.</p> : messages.map((message, index) => <div className={`message ${message.self || message.sendID === currentUser.openim_user_id ? 'self' : ''}`} key={`${message.clientMsgID || message.serverMsgID || index}`}><span>{message.text || message.content || message.msg || 'Message'}</span></div>)}</div><form className="composer" onSubmit={sendMessage}><input aria-label="Recipient OpenIM user ID" placeholder="Recipient OpenIM user ID" value={recipient} readOnly={Boolean(activeChat)} onChange={(event) => setRecipient(event.target.value)} /><input placeholder="Write a message…" value={draft} disabled={!activeChat && !recipient.trim()} onChange={(event) => setDraft(event.target.value)} /><button className="primary" type="submit" disabled={!activeChat && !recipient.trim()}>Send</button></form></section>
    </section>
    {error && <p className="error floating">{error}</p>}
  </main>
}

function toMessage(message) {
  return { ...message, text: message.textElem?.content || message.text || message.msg || parseMessageContent(message.content) }
}

function belongsToDirectChat(message, selectedUserID, currentUserID) {
  return (message.sendID === selectedUserID && message.recvID === currentUserID)
    || (message.sendID === currentUserID && message.recvID === selectedUserID)
}

function parseMessageContent(content) {
  if (!content) return ''
  if (typeof content === 'object') {
    return content.content || content.text || content.textElem?.content || ''
  }
  try {
    const parsed = JSON.parse(content)
    return parsed?.content || parsed?.text || parsed?.textElem?.content || content
  } catch {
    return content
  }
}

function sortMessages(left, right) {
  return (left.sendTime || left.createTime || 0) - (right.sendTime || right.createTime || 0)
}

function conversationPreview(latestMsg) {
  if (!latestMsg) return 'No messages yet.'
  if (typeof latestMsg === 'object') return toMessage(latestMsg).text || 'No messages yet.'
  return parseMessageContent(latestMsg) || 'No messages yet.'
}
