package pkg

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Name of the hub
	name string

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Inbound message for specific client.
	sendTo chan *sendTo
}

type sendTo struct {
	message             []byte
	clientRemoteAddress string
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		sendTo:     make(chan *sendTo),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			delete(h.clients, client)
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
				}
			}
		case sendTo := <-h.sendTo:
			for client := range h.clients {
				if client.conn.RemoteAddr().String() == sendTo.clientRemoteAddress {
					client.send <- sendTo.message
				}
			}
		}
	}
}

func (h *Hub) SendTo(message string, clientRemoteAddress string) {
	h.sendTo <- &sendTo{message: []byte(message), clientRemoteAddress: clientRemoteAddress}
}
