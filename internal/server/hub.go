package server

import (
	"log"
	"net/http"

	"github.com/Drinnn/consume-it/internal"
	"github.com/Drinnn/consume-it/pb"
)

type ClientInterfacer interface {
	Id() uint64
	ProcessMessage(senderId uint64, message pb.Msg)

	// Sets the client's ID and anything else that needs to be initialized
	Initialize(id uint64)

	// Puts data from this client in the write pump
	SocketSend(message pb.Msg)

	// Puts data from another client in the write pump
	SocketSendAs(message pb.Msg, senderId uint64)

	// Forward message to another client for processing
	PassToPeer(message pb.Msg, peerId uint64)

	// Forward message to all other clients for processing
	Broadcast(message pb.Msg)

	// Pump data from the client directly to the connected socket
	WritePump()

	// Pump data from the connected socket directly to the client
	ReadPump()

	// Close the client's connections and cleanup
	Close(reason string)
}

// The hub is the central point of communication between all connected clients
type Hub struct {
	Clients *internal.SharedCollection[ClientInterfacer]

	// Packets in this channel will be processed by all connected clients except the sender
	BroadcastChannel chan *pb.Packet

	// Clients in this channel will be registered to the hub
	RegisterChannel chan ClientInterfacer

	// Clients in this channel will be unregistered with the hub
	UnregisterChan chan ClientInterfacer
}

func NewHub() *Hub {
	return &Hub{
		Clients:          internal.NewSharedCollection[ClientInterfacer](),
		BroadcastChannel: make(chan *pb.Packet),
		RegisterChannel:  make(chan ClientInterfacer),
		UnregisterChan:   make(chan ClientInterfacer),
	}
}

func (h *Hub) Run() {
	log.Println("Awaiting client registrations...")
	for {
		select {
		case client := <-h.RegisterChannel:
			client.Initialize(h.Clients.Add(client))
		case client := <-h.UnregisterChan:
			h.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChannel:
			h.Clients.ForEach(func(clientId uint64, client ClientInterfacer) {
				if clientId != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

func (h *Hub) Serve(getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error), writer http.ResponseWriter, request *http.Request) {
	log.Println("New client connected from", request.RemoteAddr)
	client, err := getNewClient(h, writer, request)

	if err != nil {
		log.Printf("Error obtaining client for new connection: %v", err)
		return
	}

	h.RegisterChannel <- client

	go client.WritePump()
	go client.ReadPump()
}
