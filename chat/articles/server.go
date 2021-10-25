package articles

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	hubs    map[string]*Hub
	options chan Option
}

func NewServer() *Server {
	return &Server{
		hubs:    make(map[string]*Hub),
		options: make(chan Option),
	}
}

func (s *Server) GetHubs() *map[string]*Hub {
	return &s.hubs
}

func (s *Server) Run() {
	for cmd := range s.options {
		switch cmd.ID {
		// case OPT_NICK:
		// s.nick(cmd.client, cmd.args)
		case OPT_JOIN:
			s.Join(cmd.Client, cmd.Argument)
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
func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
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
		nick:    "anonymous",
		conn:    conn,
		send:    make(chan []byte, 256),
		options: s.options,
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (s *Server) Join(c *Client, argument string) {
	h, ok := s.hubs[argument]
	if !ok {
		hub := newHub()
		hub.name = argument
		s.hubs[hub.name] = hub
		c.hub = hub
		h = hub

		go s.hubs[hub.name].run()
	}

	h.clients[c] = true
	c.hub = h
	c.hub.register <- c

	c.hub.broadcast <- []byte(fmt.Sprintf("You joined the room %s", argument))
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
