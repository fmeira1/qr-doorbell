package realtime

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu       sync.Mutex
	clients  map[string]map[*websocket.Conn]bool
	upgrader websocket.Upgrader
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, sessionID string) {
	log.Println("INTENTANDO WS para sesión:", sessionID)

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR upgrade:", err)
		return
	}

	log.Println("WS CONECTADO a sesión:", sessionID)

	h.mu.Lock()
	if h.clients[sessionID] == nil {
		h.clients[sessionID] = make(map[*websocket.Conn]bool)
	}
	h.clients[sessionID][conn] = true
	h.mu.Unlock()

	defer func() {
		log.Println("WS DESCONECTADO de sesión:", sessionID)
		h.mu.Lock()
		delete(h.clients[sessionID], conn)
		h.mu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) Broadcast(sessionID string, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Println(
		"BROADCAST a sesión:",
		sessionID,
		"clientes:",
		len(h.clients[sessionID]),
	)

	for conn := range h.clients[sessionID] {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}
