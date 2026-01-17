package ws

import (
	"sync"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	mu sync.RWMutex

	// Registered clients by draft ID
	drafts map[uuid.UUID]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to a draft room
	broadcast chan *BroadcastMessage
}

// BroadcastMessage is a message to broadcast to a draft room
type BroadcastMessage struct {
	DraftID uuid.UUID
	Message []byte
	Exclude *Client // Optional client to exclude from broadcast
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		drafts:     make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.drafts[client.DraftID]; !ok {
				h.drafts[client.DraftID] = make(map[*Client]bool)
			}
			h.drafts[client.DraftID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.drafts[client.DraftID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.drafts, client.DraftID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.drafts[message.DraftID]; ok {
				for client := range clients {
					if message.Exclude != nil && client == message.Exclude {
						continue
					}
					select {
					case client.send <- message.Message:
					default:
						// Client's buffer is full, will be cleaned up
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all clients in a draft room
func (h *Hub) Broadcast(draftID uuid.UUID, message []byte, exclude *Client) {
	h.broadcast <- &BroadcastMessage{
		DraftID: draftID,
		Message: message,
		Exclude: exclude,
	}
}

// GetClientCount returns the number of clients in a draft room
func (h *Hub) GetClientCount(draftID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.drafts[draftID]; ok {
		return len(clients)
	}
	return 0
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(client *Client, message []byte) {
	select {
	case client.send <- message:
	default:
		// Client's buffer is full
	}
}
