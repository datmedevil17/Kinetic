package ws

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait      = 10 * time.Second  // max time to write a message
	pongWait       = 60 * time.Second  // max time between pings
	pingPeriod     = 54 * time.Second  // how often to send a ping (< pongWait)
	maxMessageSize = 4096              // max bytes per incoming message
)

// Client represents one connected WebSocket user.
//
// Each Client has:
//   - A gorilla WebSocket connection (conn)
//   - A buffered send channel (send) — the write pump drains this
//   - UID and SocketID for identity
//
// The key Go pattern here is the "read pump / write pump" separation:
//   Only ONE goroutine should ever write to a websocket.Conn.
//   So we have a dedicated WritePump goroutine that reads from c.send
//   and writes to the connection. All other code sends to c.send (non-blocking).
type Client struct {
	Hub      *Hub
	UID      string
	SocketID string     // UUID assigned at connection time
	Email    string
	conn     *websocket.Conn
	send     chan []byte // buffered outgoing message queue
}

// NewClient creates a Client and starts its goroutines.
func NewClient(hub *Hub, conn *websocket.Conn, uid, socketID, email string) *Client {
	c := &Client{
		Hub:      hub,
		UID:      uid,
		SocketID: socketID,
		Email:    email,
		conn:     conn,
		send:     make(chan []byte, 256), // buffer 256 messages before dropping
	}
	return c
}

// Send queues a message for delivery. Non-blocking — if the send buffer
// is full the client is too slow and will be disconnected.
func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
		// Buffer full — client is not reading fast enough; disconnect
		log.Warn().Str("uid", c.UID).Msg("Client send buffer full, disconnecting")
		close(c.send)
	}
}

// SendEvent builds a message envelope and queues it.
func (c *Client) SendEvent(event string, payload any) {
	msg := mustBuildMessage(event, payload)
	c.Send(msg)
}

// ReadPump pumps messages from the WebSocket connection to the Hub.
// Must run in its own goroutine.
//
// When the client disconnects (network error, browser close, etc.),
// conn.ReadMessage() returns an error, we break the loop, and defer
// calls hub.Unregister to clean up.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.conn.Close()
	}()

	// Limit the size of incoming messages
	c.conn.SetReadLimit(maxMessageSize)

	// Expect a Pong within pongWait after sending a Ping
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			// websocket.IsUnexpectedCloseError filters out normal browser close events
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Error().Err(err).Str("uid", c.UID).Msg("WebSocket read error")
			}
			break
		}

		// Forward raw bytes to the hub for dispatch
		c.Hub.Inbound <- InboundMessage{Client: c, Data: msg}
	}
}

// WritePump pumps messages from c.send to the WebSocket connection.
// Must run in its own goroutine. Only this goroutine writes to conn.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel — send a close frame
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			// Send a ping to keep the connection alive
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
