package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	DeviceID string
	Type     string // scanner, payment, devices
	Logger   *telemetry.Logger

	mu     sync.Mutex
	closed bool
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, deviceID, clientType string, logger *telemetry.Logger) *Client {
	return &Client{
		ID:       uuid.New().String(),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		DeviceID: deviceID,
		Type:     clientType,
		Logger:   logger,
		closed:   false,
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump(ctx context.Context, hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.Logger.Warn("websocket read error",
						telemetry.String("client_id", c.ID),
						telemetry.Error(err),
					)
				}
				return
			}

			c.Logger.Debug("received websocket message",
				telemetry.String("client_id", c.ID),
				telemetry.String("message", string(message)),
			)
		}
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client
func (c *Client) SendMessage(message []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	select {
	case c.Send <- message:
	default:
		c.Logger.Warn("client send buffer full, dropping message",
			telemetry.String("client_id", c.ID),
		)
	}
}

// SendEvent sends an event to the client
func (c *Client) SendEvent(event events.Event) error {
	// Convert event to JSON
	msg := map[string]interface{}{
		"type":      event.Type,
		"device_id": event.DeviceID,
		"timestamp": time.Now().Format(time.RFC3339),
		"data":      event.Data,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.SendMessage(data)
	return nil
}

// Close closes the client connection
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.Send)
	}
}
