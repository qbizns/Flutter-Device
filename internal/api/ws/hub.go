package ws

import (
	"context"
	"sync"

	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Hub manages WebSocket client connections
type Hub struct {
	// Registered clients by device ID
	clients map[string]map[*Client]bool

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast events to clients
	Broadcast chan BroadcastMessage

	// Event bus for subscribing to device events
	eventBus *events.Bus

	// Event subscriptions
	subscriptions map[string]*events.Subscriber

	logger  *telemetry.Logger
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	DeviceID string
	Event    events.Event
}

// NewHub creates a new WebSocket hub
func NewHub(eventBus *events.Bus, logger *telemetry.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:       make(map[string]map[*Client]bool),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Broadcast:     make(chan BroadcastMessage, 256),
		eventBus:      eventBus,
		subscriptions: make(map[string]*events.Subscriber),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			return

		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastToClients(message)
		}
	}
}

// Stop stops the hub
func (h *Hub) Stop() {
	h.cancel()

	// Close all client connections
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, clients := range h.clients {
		for client := range clients {
			client.Close()
		}
	}

	// Unsubscribe from all events
	for deviceID, sub := range h.subscriptions {
		h.eventBus.Unsubscribe(deviceID, sub.ID)
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize device client map if needed
	if h.clients[client.DeviceID] == nil {
		h.clients[client.DeviceID] = make(map[*Client]bool)

		// Subscribe to events for this device if not already subscribed
		if h.subscriptions[client.DeviceID] == nil {
			sub := h.eventBus.Subscribe(h.ctx, client.DeviceID)
			h.subscriptions[client.DeviceID] = sub

			// Start goroutine to forward events to broadcast channel
			go h.forwardEvents(client.DeviceID, sub)
		}
	}

	h.clients[client.DeviceID][client] = true

	h.logger.Info("websocket client registered",
		telemetry.String("client_id", client.ID),
		telemetry.String("device_id", client.DeviceID),
		telemetry.String("type", client.Type),
	)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.DeviceID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			client.Close()

			// If no more clients for this device, unsubscribe from events
			if len(clients) == 0 {
				delete(h.clients, client.DeviceID)

				if sub, ok := h.subscriptions[client.DeviceID]; ok {
					h.eventBus.Unsubscribe(client.DeviceID, sub.ID)
					delete(h.subscriptions, client.DeviceID)
				}
			}

			h.logger.Info("websocket client unregistered",
				telemetry.String("client_id", client.ID),
				telemetry.String("device_id", client.DeviceID),
			)
		}
	}
}

// broadcastToClients sends a message to all clients for a device
func (h *Hub) broadcastToClients(message BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[message.DeviceID]
	if !ok {
		return
	}

	for client := range clients {
		if err := client.SendEvent(message.Event); err != nil {
			h.logger.Error("failed to send event to client",
				telemetry.String("client_id", client.ID),
				telemetry.Error(err),
			)
		}
	}
}

// forwardEvents forwards events from event bus to broadcast channel
func (h *Hub) forwardEvents(deviceID string, sub *events.Subscriber) {
	for {
		select {
		case <-h.ctx.Done():
			return
		case event := <-sub.Events():
			h.Broadcast <- BroadcastMessage{
				DeviceID: deviceID,
				Event:    event,
			}
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}
