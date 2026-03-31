package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages WebSocket clients connected to specific lobbies
type Hub struct {
	// clients maps Lobby ID -> User ID -> *Client
	clients map[string]map[string]*Client
	mu      sync.RWMutex
}

// GlobalHub is the singleton hub instance
var GlobalHub *Hub

func init() {
	GlobalHub = &Hub{
		clients: make(map[string]map[string]*Client),
	}
}

// Client represents a single connected user
type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	LobbyID string
	UserID  string
}

// AddClient adds a new client to the hub
func (h *Hub) AddClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[c.LobbyID]; !ok {
		h.clients[c.LobbyID] = make(map[string]*Client)
	}

	// If the user already has a connection for this lobby, close the old one
	if oldClient, exists := h.clients[c.LobbyID][c.UserID]; exists {
		oldClient.Conn.Close()
	}

	h.clients[c.LobbyID][c.UserID] = c
	log.Printf("User %s joined WebSocket for lobby %s", c.UserID, c.LobbyID)
}

// RemoveClient removes a client from the hub
func (h *Hub) RemoveClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if lobbyClients, ok := h.clients[c.LobbyID]; ok {
		if existing, exists := lobbyClients[c.UserID]; exists && existing == c {
			delete(lobbyClients, c.UserID)
			// Clean up map if empty
			if len(lobbyClients) == 0 {
				delete(h.clients, c.LobbyID)
			}
			log.Printf("User %s left WebSocket for lobby %s", c.UserID, c.LobbyID)
		}
	}
}

// BroadcastToLobby sends a message to all connected clients in a specific lobby
func (h *Hub) BroadcastToLobby(lobbyID string, eventType string, payload interface{}) {
	msg := map[string]interface{}{
		"type":    eventType,
		"payload": payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling WS message:", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	lobbyClients, ok := h.clients[lobbyID]
	if !ok {
		return
	}

	for _, client := range lobbyClients {
		// Use a separate goroutine or non-blocking write to avoid blocking if one client is slow
		go func(c *websocket.Conn) {
			err := c.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error writing to WS:", err)
				c.Close()
			}
		}(client.Conn)
	}
}
