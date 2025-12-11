package signaling

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/Aloys-y/chat-go/models"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}

	clients    = make(map[uint]*Client)
	clientsMux sync.RWMutex

	rooms    = make(map[uint]*Room)
	roomsMux sync.RWMutex
)

// Client represents a WebSocket client
type Client struct {
	Conn     *websocket.Conn
	UserID   uint
	User     *models.User
	RoomID   uint
	SendChan chan []byte
}

// Room represents a WebRTC room
type Room struct {
	ID      uint
	Clients map[uint]*Client
	Mux     sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string          `json:"type"`
	UserID    uint            `json:"user_id,omitempty"`
	RoomID    uint            `json:"room_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	TargetID  uint            `json:"target_id,omitempty"`
}

// StartWSServer starts the WebSocket server
func StartWSServer(port int) {
	http.HandleFunc("/ws", handleWebSocket)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("WebSocket server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// handleWebSocket handles new WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Extract user ID from query parameter
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		conn.Close()
		return
	}

	// Create new client
	client := &Client{
		Conn:     conn,
		UserID:   uint(userID),
		SendChan: make(chan []byte, 256),
	}

	// Register client
	clientsMux.Lock()
	clients[client.UserID] = client
	clientsMux.Unlock()

	// Start client goroutines
	go client.readPump()
	go client.writePump()

	log.Printf("Client connected: UserID=%d", client.UserID)
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer func() {
		c.close()
	}()

	for {
		message, ok := <-c.SendChan
		if !ok {
			// Channel closed
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		// Add queued chat messages to the current WebSocket message
		n := len(c.SendChan)
		for i := 0; i < n; i++ {
			w.Write([]byte{'\n'})
			w.Write(<-c.SendChan)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	msg.UserID = c.UserID

	switch msg.Type {
	case "join_room":
		c.handleJoinRoom(msg)
	case "leave_room":
		c.handleLeaveRoom(msg)
	case "sdp_offer":
		fallthrough
	case "sdp_answer":
		fallthrough
	case "ice_candidate":
		c.handleWebRTCMessage(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleJoinRoom handles room join messages
func (c *Client) handleJoinRoom(msg Message) {
	var roomInfo struct {
		RoomID uint `json:"room_id"`
	}

	if err := json.Unmarshal(msg.Payload, &roomInfo); err != nil {
		log.Printf("Failed to unmarshal room info: %v", err)
		return
	}

	// Leave current room if any
	if c.RoomID != 0 {
		c.leaveRoom(c.RoomID)
	}

	// Join new room
	c.RoomID = roomInfo.RoomID
	c.joinRoom(c.RoomID)

	// Notify room members
	msg.Type = "user_joined"
	msg.RoomID = c.RoomID
	msg.Payload, _ = json.Marshal(map[string]interface{}{
		"user_id":       c.UserID,
		"user_name":     c.User.DisplayName,
		"user_username": c.User.Username,
	})

	c.broadcastToRoom(msg)
}

// handleLeaveRoom handles room leave messages
func (c *Client) handleLeaveRoom(msg Message) {
	if c.RoomID == 0 {
		return
	}

	roomID := c.RoomID
	c.leaveRoom(roomID)

	// Notify room members
	msg.Type = "user_left"
	msg.RoomID = roomID
	msg.Payload, _ = json.Marshal(map[string]interface{}{
		"user_id": c.UserID,
	})

	c.broadcastToRoom(msg)
}

// handleWebRTCMessage handles WebRTC signaling messages
func (c *Client) handleWebRTCMessage(msg Message) {
	if c.RoomID == 0 {
		log.Printf("Client %d not in any room", c.UserID)
		return
	}

	// Broadcast to all clients in the room except the sender
	msg.RoomID = c.RoomID
	c.broadcastToRoomExceptSender(msg)
}

// joinRoom adds a client to a room
func (c *Client) joinRoom(roomID uint) {
	roomsMux.Lock()
	room, exists := rooms[roomID]
	if !exists {
		room = &Room{
			ID:      roomID,
			Clients: make(map[uint]*Client),
		}
		rooms[roomID] = room
	}
	roomsMux.Unlock()

	room.Mux.Lock()
	room.Clients[c.UserID] = c
	room.Mux.Unlock()

	log.Printf("Client %d joined room %d", c.UserID, roomID)
}

// leaveRoom removes a client from a room
func (c *Client) leaveRoom(roomID uint) {
	roomsMux.RLock()
	room, exists := rooms[roomID]
	roomsMux.RUnlock()

	if !exists {
		return
	}

	room.Mux.Lock()
	delete(room.Clients, c.UserID)
	room.Mux.Unlock()

	// Clean up empty room
	room.Mux.RLock()
	isEmpty := len(room.Clients) == 0
	room.Mux.RUnlock()

	if isEmpty {
		roomsMux.Lock()
		delete(rooms, roomID)
		roomsMux.Unlock()
	}

	c.RoomID = 0
	log.Printf("Client %d left room %d", c.UserID, roomID)
}

// broadcastToRoom broadcasts a message to all clients in a room
func (c *Client) broadcastToRoom(msg Message) {
	roomsMux.RLock()
	room, exists := rooms[c.RoomID]
	roomsMux.RUnlock()

	if !exists {
		return
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	room.Mux.RLock()
	for _, client := range room.Clients {
		select {
		case client.SendChan <- msgBytes:
		default:
			// Client send buffer full, close connection
			log.Printf("Client %d send buffer full", client.UserID)
			client.close()
		}
	}
	room.Mux.RUnlock()
}

// broadcastToRoomExceptSender broadcasts a message to all clients in a room except the sender
func (c *Client) broadcastToRoomExceptSender(msg Message) {
	roomsMux.RLock()
	room, exists := rooms[c.RoomID]
	roomsMux.RUnlock()

	if !exists {
		return
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	room.Mux.RLock()
	for _, client := range room.Clients {
		if client.UserID == c.UserID {
			continue
		}
		select {
		case client.SendChan <- msgBytes:
		default:
			// Client send buffer full, close connection
			log.Printf("Client %d send buffer full", client.UserID)
			client.close()
		}
	}
	room.Mux.RUnlock()
}

// close closes a client connection
func (c *Client) close() {
	clientsMux.Lock()
	delete(clients, c.UserID)
	clientsMux.Unlock()

	// Leave room if any
	if c.RoomID != 0 {
		c.leaveRoom(c.RoomID)
	}

	c.Conn.Close()
	close(c.SendChan)

	log.Printf("Client %d disconnected", c.UserID)
}