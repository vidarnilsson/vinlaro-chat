import { useState, useEffect, useCallback } from 'react'
import { type Channel, listChannels } from '../api/channels'
import { type Message, getMessages, sendMessage } from '../api/messages'
import { logout } from '../api/auth'
import { listDMs, type DMConversation } from '../api/dm'
import { useAuth } from '../hooks/useAuth'
import { useWebSocket } from '../hooks/useWebSocket'
import { ChannelList } from '../components/ChannelList'
import { DMList } from '../components/DMList'
import { FriendsPanel } from '../components/FriendsPanel'
import { InvitePanel } from '../components/InvitePanel'
import { InviteButton } from '../components/InviteButton'
import { MessageList } from '../components/MessageList'
import { MessageInput } from '../components/MessageInput'

interface ActiveChannel {
  id: string
  name: string
  kind: string
}

export function ChatPage() {
  const { auth, signOut } = useAuth()
  const [channels, setChannels] = useState<Channel[]>([])
  const [dms, setDMs] = useState<DMConversation[]>([])
  const [activeChannel, setActiveChannel] = useState<ActiveChannel | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [loadingMessages, setLoadingMessages] = useState(false)

  useEffect(() => {
    listChannels().then(setChannels).catch(() => {})
    listDMs().then(setDMs).catch(() => {})
  }, [])

  async function openChannel(id: string, name: string, kind: string) {
    if (id === activeChannel?.id) return
    setActiveChannel({ id, name, kind })
    setMessages([])
    setLoadingMessages(true)
    try {
      const history = await getMessages(id)
      setMessages(history)
    } finally {
      setLoadingMessages(false)
    }
  }

  async function handleSelectChannel(channel: Channel) {
    await openChannel(channel.id, channel.name, channel.kind)
  }

  async function handleOpenDM(channelId: string, otherUsername: string) {
    await openChannel(channelId, otherUsername, 'dm')
  }

  const handleWsMessage = useCallback((incoming: Message) => {
    setMessages((prev) => {
      if (prev.some((m) => m.id === incoming.id)) return prev
      return [...prev, incoming]
    })
  }, [])

  useWebSocket({
    channelId: activeChannel?.id ?? null,
    onMessage: handleWsMessage,
  })

  async function handleLogout() {
    await logout().catch(() => {})
    signOut()
  }

  async function handleSend(content: string) {
    if (!activeChannel || !auth) return
    const optimistic: Message = {
      id: crypto.randomUUID(),
      channel_id: activeChannel.id,
      user_id: auth.user_id,
      username: auth.username,
      content,
      created_at: new Date().toISOString(),
    }
    setMessages((prev) => [...prev, optimistic])
    try {
      const confirmed = await sendMessage(activeChannel.id, content)
      setMessages((prev) => prev.map((m) => m.id === optimistic.id ? confirmed : m))
    } catch {
      setMessages((prev) => prev.filter((m) => m.id !== optimistic.id))
    }
  }

  function handleDMCreated(dm: DMConversation) {
    setDMs((prev) => {
      if (prev.some((d) => d.channel_id === dm.channel_id)) return prev
      return [dm, ...prev]
    })
  }

  function handleChannelJoined(channel: Channel) {
    setChannels((prev) => {
      if (prev.some((c) => c.id === channel.id)) return prev
      return [...prev, channel]
    })
  }

  const channelLabel = activeChannel
    ? activeChannel.kind === 'dm'
      ? `@ ${activeChannel.name}`
      : activeChannel.kind === 'private'
        ? `🔒 ${activeChannel.name}`
        : `# ${activeChannel.name}`
    : null

  return (
    <div style={{ display: 'flex', height: '100vh', flexDirection: 'column' }}>
      <header style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '10px 16px',
        borderBottom: '1px solid #ddd',
        flexShrink: 0,
      }}>
        <strong>Vinlaro Chat</strong>
        <span style={{ fontSize: 13, color: '#555' }}>
          {auth?.username}{' '}
          <button onClick={handleLogout} style={{ marginLeft: 8, fontSize: 12, cursor: 'pointer' }}>
            Log out
          </button>
        </span>
      </header>

      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        <aside style={{
          width: 230,
          borderRight: '1px solid #ddd',
          padding: 12,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: 0,
        }}>
          <ChannelList
            channels={channels}
            activeId={activeChannel?.id ?? null}
            onSelect={handleSelectChannel}
            onCreated={(ch) => {
              setChannels((prev) => [...prev, ch])
              handleSelectChannel(ch)
            }}
          />
          <DMList
            dms={dms}
            activeChannelId={activeChannel?.id ?? null}
            onOpenDM={handleOpenDM}
            onDMCreated={handleDMCreated}
          />
          <FriendsPanel onOpenDM={handleOpenDM} onDMCreated={handleDMCreated} />
          <InvitePanel onChannelJoined={handleChannelJoined} />
        </aside>

        <main style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          {activeChannel ? (
            <>
              <div style={{
                padding: '8px 16px',
                borderBottom: '1px solid #ddd',
                fontWeight: 600,
                display: 'flex',
                alignItems: 'center',
                gap: 12,
              }}>
                <span>{channelLabel}</span>
                {activeChannel.kind === 'private' && (
                  <InviteButton channelId={activeChannel.id} />
                )}
              </div>
              {loadingMessages ? (
                <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#888' }}>
                  Loading messages…
                </div>
              ) : (
                <MessageList messages={messages} currentUserId={auth?.user_id ?? ''} />
              )}
              <MessageInput onSend={handleSend} disabled={loadingMessages} />
            </>
          ) : (
            <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#888' }}>
              Select a channel or conversation to start chatting
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
