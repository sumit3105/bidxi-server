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
	writeMu sync.Mutex
}

// SafeWrite sends a message to this client with write-mutex protection.
func (c *Client) SafeWrite(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.Conn.WriteMessage(websocket.TextMessage, data)
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
	go h.BroadcastOnlineUsers(c.LobbyID)
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
			go h.BroadcastOnlineUsers(c.LobbyID)
		}
	}
}

// BroadcastOnlineUsers broadcasts the list of online user IDs to the lobby
func (h *Hub) BroadcastOnlineUsers(lobbyID string) {
	h.mu.RLock()
	lobbyClients, ok := h.clients[lobbyID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	var userIDs []string
	for uid := range lobbyClients {
		userIDs = append(userIDs, uid)
	}
	h.mu.RUnlock()
	h.BroadcastToLobby(lobbyID, "ONLINE_USERS", userIDs)
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
	lobbyClients, ok := h.clients[lobbyID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	// Copy client list to avoid holding lock during writes
	clients := make([]*Client, 0, len(lobbyClients))
	for _, c := range lobbyClients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		go func(c *Client) {
			if err := c.SafeWrite(data); err != nil {
				log.Printf("Error writing to WS user %s: %v", c.UserID, err)
				c.Conn.Close()
			}
		}(client)
	}
}

// SendToUser sends a message to a specific user in a lobby
func (h *Hub) SendToUser(lobbyID string, userID string, eventType string, payload interface{}) {
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
	lobbyClients, ok := h.clients[lobbyID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	client, exists := lobbyClients[userID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	go func(c *Client) {
		if err := c.SafeWrite(data); err != nil {
			log.Printf("Error writing to WS user %s: %v", c.UserID, err)
			c.Conn.Close()
		}
	}(client)
}
