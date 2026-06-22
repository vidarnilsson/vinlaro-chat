package ws

import (
	"sync"

	"github.com/google/uuid"
)

type broadcastMsg struct {
	channelID uuid.UUID
	payload   []byte
}

// Hub maintains all active WebSocket clients and routes broadcasts to the
// correct channel room.
type Hub struct {
	mu       sync.RWMutex
	rooms    map[uuid.UUID]map[*Client]bool
	register chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan broadcastMsg, 256),
	}
}

// Run drives the hub. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.channelID] == nil {
				h.rooms[client.channelID] = make(map[*Client]bool)
			}
			h.rooms[client.channelID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.channelID]; ok {
				if room[client] {
					delete(room, client)
					close(client.send)
				}
				if len(room) == 0 {
					delete(h.rooms, client.channelID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			room := h.rooms[msg.channelID]
			h.mu.RUnlock()
			for client := range room {
				select {
				case client.send <- msg.payload:
				default:
					// Slow client — drop and disconnect.
					h.unregister <- client
				}
			}
		}
	}
}

// Register enqueues a client for registration with the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// BroadcastToChannel fans out payload to all clients in the given channel.
func (h *Hub) BroadcastToChannel(channelID uuid.UUID, payload []byte) {
	h.broadcast <- broadcastMsg{channelID: channelID, payload: payload}
}
