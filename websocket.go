package hago

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient represents a WebSocket client for Home Assistant
type WSClient struct {
	URL         string
	AccessToken string
	conn        *websocket.Conn
	mu          sync.Mutex
	connected   bool
	msgID       int64
	handlers    map[string][]func(msg map[string]interface{})
	done        chan struct{}
}

// NewWSClient creates a new WebSocket client
func NewWSClient(url, accessToken string) *WSClient {
	return &WSClient{
		URL:         url,
		AccessToken: accessToken,
		msgID:       0,
		handlers:    make(map[string][]func(msg map[string]interface{})),
		done:        make(chan struct{}),
	}
}

// Connect establishes a WebSocket connection to Home Assistant
func (c *WSClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	conn, _, err := websocket.DefaultDialer.Dial(c.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %v", err)
	}

	c.conn = conn
	c.connected = true

	// Start the message reader
	go c.readMessages()

	// Authenticate with Home Assistant
	err = c.authenticate()
	if err != nil {
		c.conn.Close()
		c.connected = false
		return fmt.Errorf("authentication failed: %v", err)
	}

	return nil
}

// authenticate sends an authentication message to Home Assistant
func (c *WSClient) authenticate() error {
	authMsg := map[string]interface{}{
		"type":         "auth",
		"access_token": c.AccessToken,
	}

	err := c.conn.WriteJSON(authMsg)
	if err != nil {
		return err
	}

	// Wait for auth response
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(message, &response); err != nil {
		return err
	}

	if response["type"] != "auth_ok" {
		return fmt.Errorf("authentication failed: %v", response["message"])
	}

	return nil
}

// Close closes the WebSocket connection
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	close(c.done)
	err := c.conn.Close()
	c.connected = false
	return err
}

// nextID generates a unique message ID
func (c *WSClient) nextID() int64 {
	return atomic.AddInt64(&c.msgID, 1)
}

// readMessages reads messages from the WebSocket connection
func (c *WSClient) readMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading WebSocket message: %v", err)
				c.reconnect()
				continue
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			// Handle message based on its type
			if msgType, ok := msg["type"].(string); ok {
				c.handleMessage(msgType, msg)
			}
		}
	}
}

// handleMessage dispatches the message to registered handlers
func (c *WSClient) handleMessage(msgType string, msg map[string]interface{}) {
	c.mu.Lock()
	handlers, ok := c.handlers[msgType]
	c.mu.Unlock()

	if ok {
		for _, handler := range handlers {
			go handler(msg)
		}
	}
}

// reconnect attempts to reconnect to the WebSocket server
func (c *WSClient) reconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return
	}

	c.conn.Close()
	c.connected = false

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to reconnect to WebSocket (attempt %d/%d)", i+1, maxRetries)
		conn, _, err := websocket.DefaultDialer.Dial(c.URL, nil)
		if err == nil {
			c.conn = conn
			c.connected = true
			if err := c.authenticate(); err != nil {
				conn.Close()
				c.connected = false
				log.Printf("Authentication failed during reconnect: %v", err)
			} else {
				log.Printf("Successfully reconnected to WebSocket")
				return
			}
		}
		time.Sleep(time.Second * 3 * time.Duration(i+1))
	}
	log.Printf("Failed to reconnect to WebSocket after %d attempts", maxRetries)
}

// Send sends a message to the WebSocket server
func (c *WSClient) Send(message map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to WebSocket")
	}

	// Add message ID if not present
	if _, ok := message["id"]; !ok {
		message["id"] = c.nextID()
	}

	return c.conn.WriteJSON(message)
}

// SubscribeEvents subscribes to Home Assistant events
func (c *WSClient) SubscribeEvents(eventType string) error {
	message := map[string]interface{}{
		"type": "subscribe_events",
		"id":   c.nextID(),
	}

	if eventType != "" {
		message["event_type"] = eventType
	}

	return c.Send(message)
}

// UnsubscribeEvents unsubscribes from Home Assistant events
func (c *WSClient) UnsubscribeEvents(subscription int64) error {
	message := map[string]interface{}{
		"type":         "unsubscribe_events",
		"id":           c.nextID(),
		"subscription": subscription,
	}

	return c.Send(message)
}

// AddEventHandler registers a handler function for event messages
func (c *WSClient) AddEventHandler(eventType string, handler func(msg map[string]interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.handlers[eventType]; !ok {
		c.handlers[eventType] = []func(msg map[string]interface{}){}
	}
	c.handlers[eventType] = append(c.handlers[eventType], handler)
}
