package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/srvsurya/system-monitor/internal/models"
)

// upgrades http to ws
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { // remember to change this in prod, true = allows access from all sources
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// goroutine that writes from metrics send to websocket conn
func (cl *Client) writePump() {
	defer cl.conn.Close()

	for msg := range cl.send {
		if err := cl.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[ws] write error: %v", err)
			return
		}
	}
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 10),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run is the central event loop. Call once in a goroutine from main.go.
func (h *Hub) Run() {
	for {
		select {

		case client := <-h.register:
			h.clients[client] = true
			log.Printf("[ws] client connected — total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("[ws] client disconnected — total: %d", len(h.clients))
			}

		case msg := <-h.broadcast:
			// Fan out to every connected client
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:

					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

// StartBroadcasting ticks every 5 seconds, reads the latest metric row,
// and drops it onto the broadcast channel for Hub.Run() to fan out.
func (h *Hub) StartBroadcasting(db *sqlx.DB) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var metric models.SystemMetric

		err := db.Get(&metric, `
			SELECT * FROM system_metrics
			ORDER BY timestamp DESC
			LIMIT 1
		`)
		if err != nil {
			log.Printf("[ws] failed to fetch metric: %v", err)
			continue
		}

		data, err := json.Marshal(metric)
		if err != nil {
			log.Printf("[ws] failed to marshal metric: %v", err)
			continue
		}

		h.broadcast <- data
	}
}

// ServeWS upgrades the HTTP connection to WebSocket and registers the client.
func ServeWS(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil) // conn basically is a websocket conn type that's sent to the client instance
		if err != nil {
			log.Printf("[ws] upgrade error: %v", err)
			return
		}

		client := &Client{
			conn: conn, // type *websocket.Conn
			send: make(chan []byte, 10),
		}

		hub.register <- client

		// writePump runs in its own goroutine — one per client
		go client.writePump()

		// this is only here to unregister and clean up client connection after EOF.
		go func() {
			defer func() {
				hub.unregister <- client
			}()
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()
	}
}
