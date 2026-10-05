package ws

import (
	"log"
	"net/http"
	"sync"
	"time"

	"chatapp/model"
	"chatapp/pkg/redisrepo"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CORS is handled by the http server, accept any origin here
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	Username string
	Conn     *websocket.Conn
	send     chan model.Chat
}

// clients maps a username to its active connection
var (
	clients   = make(map[string]*Client)
	clientsMu sync.RWMutex
)

// ServeWs upgrades the request to a websocket connection.
// The user is identified by the "username" query param: /ws?username=john
func ServeWs(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" || !redisrepo.IsUserExist(username) {
		http.Error(w, "unknown user", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error while upgrading connection", err)
		return
	}

	client := &Client{Username: username, Conn: conn, send: make(chan model.Chat, 16)}
	register(client)

	go client.writePump()
	client.readPump()
}

func register(c *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// a user reconnecting replaces its previous connection
	if old, ok := clients[c.Username]; ok {
		close(old.send)
	}
	clients[c.Username] = c
	log.Println("client connected:", c.Username)
}

func unregister(c *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if clients[c.Username] == c {
		delete(clients, c.Username)
		close(c.send)
	}
	log.Println("client disconnected:", c.Username)
}

// readPump reads chats sent by the client, persists them and forwards
// them to the recipient (if online) and back to the sender.
func (c *Client) readPump() {
	defer func() {
		unregister(c)
		c.Conn.Close()
	}()

	for {
		var chat model.Chat
		if err := c.Conn.ReadJSON(&chat); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Println("error while reading message", err)
			}
			return
		}

		// never trust the sender field coming from the client
		chat.From = c.Username
		chat.Timestamp = time.Now().Unix()

		if chat.To == "" || chat.Msg == "" || !redisrepo.IsUserExist(chat.To) {
			log.Println("invalid chat from", c.Username, "to", chat.To)
			continue
		}

		id, err := redisrepo.CreateChat(&chat)
		if err != nil {
			continue
		}
		chat.ID = id

		deliver(chat.To, chat)
		if chat.To != chat.From {
			deliver(chat.From, chat)
		}
	}
}

func deliver(username string, chat model.Chat) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	if c, ok := clients[username]; ok {
		select {
		case c.send <- chat:
		default:
			log.Println("send buffer full, dropping message for", username)
		}
	}
}

func (c *Client) writePump() {
	for chat := range c.send {
		if err := c.Conn.WriteJSON(chat); err != nil {
			log.Println("error while writing message", err)
			c.Conn.Close()
			return
		}
	}
	c.Conn.Close()
}
