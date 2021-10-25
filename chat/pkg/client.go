package pkg

import (
	"bytes"
	"fmt"
	"go-chat/messager"
	"go-chat/persistence"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	StockPattern   = `/stock=(?P<Stock>.*)`
	JoinPattern    = `/join=(?P<Join>.*)`
	QuitPattern    = `/quit=(?P<Quit>.*)`
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// Nickname.
	nick string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Options to be send to the server
	options chan<- Option
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			paramsMap := persistence.GetParams(JoinPattern, string(message))
			joinKey := "Join"
			if _, ok := paramsMap[joinKey]; ok {
				c.hub.unregister <- c
				option := &Option{ID: OPT_JOIN, Client: c, Argument: paramsMap[joinKey]}
				c.options <- *option
				delete(paramsMap, joinKey)
				continue
			}

			paramsMap = persistence.GetParams(QuitPattern, string(message))
			quitKey := "Quit"
			if _, ok := paramsMap[quitKey]; ok {
				c.hub.unregister <- c
				option := &Option{ID: OPT_QUIT, Client: c, Argument: paramsMap[quitKey]}
				c.options <- *option
				delete(paramsMap, quitKey)
				continue
			}

			hours, minutes, _ := time.Now().Clock()
			message = []byte(fmt.Sprintf("%d:%02d - %s", hours, minutes, message))

			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

			paramsMap = persistence.GetParams(StockPattern, string(message))
			stockKey := "Stock"
			if _, ok := paramsMap[stockKey]; ok {
				message := messager.ClientMessage{HubName: c.hub.name, ClientRemoteAddress: c.conn.RemoteAddr().String(), Message: paramsMap[stockKey]}
				messager.SendMessage(&message)
				delete(paramsMap, stockKey)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
