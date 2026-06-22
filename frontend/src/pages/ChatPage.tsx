import { useState, useEffect, useCallback } from 'react'
import { type Channel, listChannels } from '../api/channels'
import { type Message, getMessages, sendMessage } from '../api/messages'
import { useAuth } from '../hooks/useAuth'
import { useWebSocket } from '../hooks/useWebSocket'
import { ChannelList } from '../components/ChannelList'
import { MessageList } from '../components/MessageList'
import { MessageInput } from '../components/MessageInput'

export function ChatPage() {
  const { auth, signOut } = useAuth()
  const [channels, setChannels] = useState<Channel[]>([])
  const [activeChannel, setActiveChannel] = useState<Channel | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [loadingMessages, setLoadingMessages] = useState(false)

  useEffect(() => {
    listChannels()
      .then(setChannels)
      .catch(() => {}) // non-fatal on load
  }, [])

  async function handleSelectChannel(channel: Channel) {
    if (channel.id === activeChannel?.id) return
    setActiveChannel(channel)
    setMessages([])
    setLoadingMessages(true)
    try {
      const history = await getMessages(channel.id)
      setMessages(history)
    } finally {
      setLoadingMessages(false)
    }
  }

  const handleWsMessage = useCallback((incoming: Message) => {
    setMessages((prev) => {
      if (prev.some((m) => m.id === incoming.id)) return prev
      return [...prev, incoming]
    })
  }, [])

  useWebSocket({
    channelId: activeChannel?.id ?? null,
    token: auth?.token ?? null,
    onMessage: handleWsMessage,
  })

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
      // Replace the optimistic message with the server-confirmed one (has real ID).
      setMessages((prev) => prev.map((m) => m.id === optimistic.id ? confirmed : m))
    } catch {
      // Remove optimistic message on failure.
      setMessages((prev) => prev.filter((m) => m.id !== optimistic.id))
    }
  }

  return (
    <div style={{ display: 'flex', height: '100vh', flexDirection: 'column' }}>
      <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '10px 16px', borderBottom: '1px solid #ddd' }}>
        <strong>Vinlaro Chat</strong>
        <span style={{ fontSize: 13, color: '#555' }}>
          {auth?.username}{' '}
          <button onClick={signOut} style={{ marginLeft: 8, fontSize: 12, cursor: 'pointer' }}>
            Log out
          </button>
        </span>
      </header>

      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        <ChannelList
          channels={channels}
          activeId={activeChannel?.id ?? null}
          onSelect={handleSelectChannel}
          onCreated={(ch) => {
            setChannels((prev) => [...prev, ch])
            handleSelectChannel(ch)
          }}
        />

        <main style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          {activeChannel ? (
            <>
              <div style={{ padding: '10px 16px', borderBottom: '1px solid #ddd', fontWeight: 600 }}>
                # {activeChannel.name}
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
              Select a channel to start chatting
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
