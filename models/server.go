package models

import (
	"fmt"
	"log"
	"net/http"
)

type server struct {
	hubs    map[string]*Hub
	options chan option
}

func NewServer() *server {
	return &server{
		hubs:    make(map[string]*Hub),
		options: make(chan option),
	}
}

func (s *server) Run() {
	for cmd := range s.options {
		switch cmd.id {
		// case OPT_NICK:
		// s.nick(cmd.client, cmd.args)
		case OPT_JOIN:
			s.Join(cmd.client, cmd.argument)
			// case OPT_ROOMS:
			// 	s.listRooms(cmd.client, cmd.args)
			// case OPT_MSG:
			// 	s.msg(cmd.client, cmd.args)
			// case OPT_QUIT:
			// 	s.quit(cmd.client, cmd.args)
		}
	}
}

// serveWs handles websocket requests from the peer.
func (s *server) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	if len(s.hubs) == 0 {
		hub := newHub()
		hub.name = "#general"
		s.hubs[hub.name] = hub

		go s.hubs[hub.name].run()
	}

	client := &Client{hub: s.hubs["#general"],
		nick: "anonymous",
		conn: conn,
		send: make(chan []byte, 256)}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (s *server) Join(c *Client, argument string) {
	h, ok := s.hubs[argument]
	if !ok {
		h = &Hub{
			name:    argument,
			clients: make(map[*Client]bool),
		}
		s.hubs[argument] = h
	}

	h.clients[c] = true

	// s.quitCurrentRoom(c)

	c.hub = h

	c.hub.broadcast <- []byte(fmt.Sprintf("%s has joined the room", c.nick))
}

// func (s *server) listRooms(c *Client, args []string) {
// 	var hubs []string
// 	for name := range s.hubs {
// 		hubs = append(hubs, name)
// 	}

// 	c.msg(fmt.Sprintf("available rooms are: %s", strings.Join(hubs, ", ")))
// }

// func (s *server) quitCurrentRoom(c *Client) {
// 	if c.hub != nil {
// 		delete(c.hub.clients, c)
// 		c.hub.broadcast <- []byte(fmt.Sprintf("%s has left the room", c.nick))
// 	}
// }
