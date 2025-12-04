// Package realtime provides WebSocket-based real-time emissions streaming.
//
// This enables live dashboard updates as activities are ingested and emissions
// calculated, giving organizations sub-second visibility into their carbon footprint.
package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Message Types
// =============================================================================

// MessageType identifies the type of real-time message.
type MessageType string

const (
	// MessageTypeEmission is sent when new emissions are calculated.
	MessageTypeEmission MessageType = "emission"

	// MessageTypeActivity is sent when new activities are ingested.
	MessageTypeActivity MessageType = "activity"

	// MessageTypeAlert is sent for anomaly detections or threshold breaches.
	MessageTypeAlert MessageType = "alert"

	// MessageTypeCompliance is sent for compliance status changes.
	MessageTypeCompliance MessageType = "compliance"

	// MessageTypeHeartbeat keeps connections alive.
	MessageTypeHeartbeat MessageType = "heartbeat"
)

// Message represents a real-time update sent to clients.
type Message struct {
	ID        string          `json:"id"`
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	TenantID  string          `json:"tenantId"`
	Payload   json.RawMessage `json:"payload"`
}

// EmissionPayload contains emission update data.
type EmissionPayload struct {
	Scope           int     `json:"scope"`
	EmissionsKgCO2e float64 `json:"emissionsKgCo2e"`
	Source          string  `json:"source"`
	Category        string  `json:"category,omitempty"`
	Region          string  `json:"region,omitempty"`
	Delta           float64 `json:"delta,omitempty"` // Change from previous
}

// ActivityPayload contains activity ingestion data.
type ActivityPayload struct {
	Source     string  `json:"source"`
	Count      int     `json:"count"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
	Processing bool    `json:"processing"`
}

// AlertPayload contains anomaly or threshold alert data.
type AlertPayload struct {
	Severity    string  `json:"severity"` // critical, high, medium, low
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Metric      string  `json:"metric,omitempty"`
	Value       float64 `json:"value,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

// =============================================================================
// Client Management
// =============================================================================

// Client represents a connected WebSocket client.
type Client struct {
	ID         string
	TenantID   string
	UserID     string
	Send       chan []byte
	Hub        *Hub
	done       chan struct{}
	subscribed map[MessageType]bool
	mu         sync.RWMutex
}

// NewClient creates a new WebSocket client.
func NewClient(tenantID, userID string, hub *Hub) *Client {
	return &Client{
		ID:       uuid.NewString(),
		TenantID: tenantID,
		UserID:   userID,
		Send:     make(chan []byte, 256),
		Hub:      hub,
		done:     make(chan struct{}),
		subscribed: map[MessageType]bool{
			MessageTypeEmission:   true,
			MessageTypeActivity:   true,
			MessageTypeAlert:      true,
			MessageTypeCompliance: true,
		},
	}
}

// Subscribe adds message type subscriptions.
func (c *Client) Subscribe(types ...MessageType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, t := range types {
		c.subscribed[t] = true
	}
}

// Unsubscribe removes message type subscriptions.
func (c *Client) Unsubscribe(types ...MessageType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, t := range types {
		delete(c.subscribed, t)
	}
}

// IsSubscribed checks if client is subscribed to a message type.
func (c *Client) IsSubscribed(t MessageType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscribed[t]
}

// Close signals the client to disconnect.
func (c *Client) Close() {
	close(c.done)
}

// =============================================================================
// Hub - Central Message Broker
// =============================================================================

// Hub manages all WebSocket connections and message broadcasting.
type Hub struct {
	// Registered clients by tenant ID
	clients map[string]map[*Client]bool

	// Channel for broadcasting messages
	broadcast chan *Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Metrics
	metrics *HubMetrics

	// Logger
	logger *slog.Logger

	mu sync.RWMutex
}

// HubMetrics tracks real-time streaming metrics.
type HubMetrics struct {
	ConnectedClients int64        `json:"connectedClients"`
	MessagesSent     int64        `json:"messagesSent"`
	MessagesDropped  int64        `json:"messagesDropped"`
	BytesSent        int64        `json:"bytesSent"`
	ConnectionsTotal int64        `json:"connectionsTotal"`
	DisconnectsTotal int64        `json:"disconnectsTotal"`
	mu               sync.RWMutex `json:"-"`
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 1024),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		metrics:    &HubMetrics{},
		logger:     logger.With("component", "realtime-hub"),
	}
}

// Run starts the hub's main event loop.
func (h *Hub) Run(ctx context.Context) {
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("hub shutting down")
			h.closeAllClients()
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-heartbeatTicker.C:
			h.sendHeartbeat()
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all subscribed clients for a tenant.
func (h *Hub) Broadcast(msg *Message) {
	select {
	case h.broadcast <- msg:
	default:
		h.metrics.mu.Lock()
		h.metrics.MessagesDropped++
		h.metrics.mu.Unlock()
		h.logger.Warn("broadcast channel full, message dropped",
			"messageType", msg.Type,
			"tenantId", msg.TenantID)
	}
}

// BroadcastEmission is a convenience method for emission updates.
func (h *Hub) BroadcastEmission(tenantID string, payload EmissionPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	h.Broadcast(&Message{
		ID:        uuid.NewString(),
		Type:      MessageTypeEmission,
		Timestamp: time.Now(),
		TenantID:  tenantID,
		Payload:   data,
	})
	return nil
}

// BroadcastActivity is a convenience method for activity updates.
func (h *Hub) BroadcastActivity(tenantID string, payload ActivityPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	h.Broadcast(&Message{
		ID:        uuid.NewString(),
		Type:      MessageTypeActivity,
		Timestamp: time.Now(),
		TenantID:  tenantID,
		Payload:   data,
	})
	return nil
}

// BroadcastAlert is a convenience method for alert notifications.
func (h *Hub) BroadcastAlert(tenantID string, payload AlertPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	h.Broadcast(&Message{
		ID:        uuid.NewString(),
		Type:      MessageTypeAlert,
		Timestamp: time.Now(),
		TenantID:  tenantID,
		Payload:   data,
	})
	return nil
}

// GetMetrics returns current hub metrics.
func (h *Hub) GetMetrics() HubMetrics {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()
	// Return a copy without the mutex to avoid copying the lock
	return HubMetrics{
		ConnectedClients: h.metrics.ConnectedClients,
		MessagesSent:     h.metrics.MessagesSent,
		MessagesDropped:  h.metrics.MessagesDropped,
		BytesSent:        h.metrics.BytesSent,
		ConnectionsTotal: h.metrics.ConnectionsTotal,
		DisconnectsTotal: h.metrics.DisconnectsTotal,
	}
}

// GetClientCount returns the number of connected clients for a tenant.
func (h *Hub) GetClientCount(tenantID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[tenantID]; ok {
		return len(clients)
	}
	return 0
}

// GetTotalClientCount returns the total number of connected clients.
func (h *Hub) GetTotalClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// =============================================================================
// Internal Methods
// =============================================================================

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.TenantID]; !ok {
		h.clients[client.TenantID] = make(map[*Client]bool)
	}
	h.clients[client.TenantID][client] = true

	h.metrics.mu.Lock()
	h.metrics.ConnectedClients++
	h.metrics.ConnectionsTotal++
	h.metrics.mu.Unlock()

	h.logger.Info("client registered",
		"clientId", client.ID,
		"tenantId", client.TenantID,
		"userId", client.UserID)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.TenantID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)

			if len(clients) == 0 {
				delete(h.clients, client.TenantID)
			}

			h.metrics.mu.Lock()
			h.metrics.ConnectedClients--
			h.metrics.DisconnectsTotal++
			h.metrics.mu.Unlock()

			h.logger.Info("client unregistered",
				"clientId", client.ID,
				"tenantId", client.TenantID)
		}
	}
}

func (h *Hub) broadcastMessage(msg *Message) {
	h.mu.RLock()
	clients := h.clients[msg.TenantID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal message", "error", err)
		return
	}

	for client := range clients {
		if !client.IsSubscribed(msg.Type) {
			continue
		}

		select {
		case client.Send <- data:
			h.metrics.mu.Lock()
			h.metrics.MessagesSent++
			h.metrics.BytesSent += int64(len(data))
			h.metrics.mu.Unlock()
		default:
			h.logger.Warn("client send buffer full",
				"clientId", client.ID)
			h.metrics.mu.Lock()
			h.metrics.MessagesDropped++
			h.metrics.mu.Unlock()
		}
	}
}

func (h *Hub) sendHeartbeat() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	heartbeat := &Message{
		ID:        uuid.NewString(),
		Type:      MessageTypeHeartbeat,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(heartbeat)

	for _, clients := range h.clients {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				// Skip if buffer full
			}
		}
	}
}

func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, clients := range h.clients {
		for client := range clients {
			close(client.Send)
		}
	}
	h.clients = make(map[string]map[*Client]bool)
}

// =============================================================================
// HTTP Handler
// =============================================================================

// Handler provides HTTP endpoints for real-time connections.
type Handler struct {
	hub    *Hub
	logger *slog.Logger
}

// NewHandler creates a new real-time HTTP handler.
func NewHandler(hub *Hub, logger *slog.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// ServeHTTP handles WebSocket upgrade requests.
// Note: Actual WebSocket upgrade requires gorilla/websocket or nhooyr.io/websocket.
// This is a placeholder showing the integration pattern.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// In production, use:
	// conn, err := websocket.Accept(w, r, nil)
	//
	// For now, return SSE (Server-Sent Events) as fallback
	h.serveSSE(w, r)
}

// serveSSE provides Server-Sent Events as WebSocket fallback.
func (h *Handler) serveSSE(w http.ResponseWriter, r *http.Request) {
	// Get tenant from context (set by auth middleware)
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" {
		http.Error(w, "tenant ID required", http.StatusUnauthorized)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create and register client
	client := NewClient(tenantID, userID, h.hub)
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// Send initial connection event
	_, _ = w.Write([]byte("event: connected\ndata: {\"clientId\":\"" + client.ID + "\"}\n\n"))
	flusher.Flush()

	// Stream messages
	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			_, _ = w.Write([]byte("event: message\ndata: "))
			_, _ = w.Write(msg)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

// MetricsHandler returns hub metrics.
func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	m := h.hub.GetMetrics()
	// Create anonymous struct without mutex for JSON encoding
	response := struct {
		ConnectedClients int64 `json:"connectedClients"`
		MessagesSent     int64 `json:"messagesSent"`
		MessagesDropped  int64 `json:"messagesDropped"`
		BytesSent        int64 `json:"bytesSent"`
		ConnectionsTotal int64 `json:"connectionsTotal"`
		DisconnectsTotal int64 `json:"disconnectsTotal"`
	}{
		ConnectedClients: m.ConnectedClients,
		MessagesSent:     m.MessagesSent,
		MessagesDropped:  m.MessagesDropped,
		BytesSent:        m.BytesSent,
		ConnectionsTotal: m.ConnectionsTotal,
		DisconnectsTotal: m.DisconnectsTotal,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
