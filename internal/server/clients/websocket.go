package clients

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Drinnn/consume-it/internal/server"
	"github.com/Drinnn/consume-it/internal/server/states"
	"github.com/Drinnn/consume-it/pb"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type WebSocketClient struct {
	id          uint64
	conn        *websocket.Conn
	hub         *server.Hub
	sendChannel chan *pb.Packet
	state       server.ClientStateMachineHandler
	logger      *log.Logger
}

func NewWebSocketClient(hub *server.Hub, writer http.ResponseWriter, request *http.Request) (server.ClientInterfacer, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		return nil, err
	}

	client := &WebSocketClient{
		hub:         hub,
		conn:        conn,
		sendChannel: make(chan *pb.Packet, 256),
		logger:      log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
	}

	return client, nil
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}

func (c *WebSocketClient) ProcessMessage(senderId uint64, message pb.Msg) {
	c.state.HandleMessage(senderId, message)
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id
	c.logger.SetPrefix(fmt.Sprintf("Client %d: ", id))
	c.SetState(&states.Connected{})
}

func (c *WebSocketClient) SetState(state server.ClientStateMachineHandler) {
	prevStateName := "None"
	if c.state != nil {
		prevStateName = c.state.Name()
		c.state.OnExit()
	}

	newStateName := "None"
	if state != nil {
		newStateName = state.Name()
	}

	c.logger.Printf("Switching from state %s to %s", prevStateName, newStateName)

	c.state = state

	if c.state != nil {
		c.state.SetClient(c)
		c.state.OnEnter()
	}
}

func (c *WebSocketClient) SocketSend(message pb.Msg) {
	c.SocketSendAs(message, c.id)
}

func (c *WebSocketClient) SocketSendAs(message pb.Msg, senderId uint64) {
	select {
	case c.sendChannel <- &pb.Packet{SenderId: senderId, Msg: message}:
	default:
		c.logger.Printf("Send channel full, dropping message: %T", message)
	}
}

func (c *WebSocketClient) PassToPeer(message pb.Msg, peerId uint64) {
	if peer, exists := c.hub.Clients.Get(peerId); exists {
		peer.ProcessMessage(c.id, message)
	}
}

func (c *WebSocketClient) Broadcast(message pb.Msg) {
	c.hub.BroadcastChannel <- &pb.Packet{SenderId: c.id, Msg: message}
}

func (c *WebSocketClient) WritePump() {
	defer func() {
		c.logger.Println("Closing write pump")
		c.conn.Close()
	}()

	for packet := range c.sendChannel {
		writer, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			c.logger.Printf("Error getting writer for %T packet, closing client: %v", packet.Msg, err)
			return
		}

		data, err := proto.Marshal(packet)
		if err != nil {
			c.logger.Printf("Error parsing %T packet, closing client: %v", packet.Msg, err)
			continue
		}

		_, err = writer.Write(data)
		if err != nil {
			c.logger.Printf("Error writing %T packet, closing client: %v", packet.Msg, err)
			return
		}

		writer.Write([]byte{'\n'})

		if err := writer.Close(); err != nil {
			c.logger.Printf("Error closing writer for %T packet, closing client: %v", packet.Msg, err)
			return
		}
	}

}

func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.logger.Println("Closing read pump")
		c.Close("Read pump closed")
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("Error: %v", err)
			}
			break
		}

		packet := &pb.Packet{}
		err = proto.Unmarshal(data, packet)
		if err != nil {
			c.logger.Printf("Error parsing packet: %v", err)
			continue
		}

		if packet.SenderId == 0 {
			packet.SenderId = c.id
		}

		c.ProcessMessage(packet.SenderId, packet.Msg)
	}
}

func (c *WebSocketClient) Close(reason string) {
	c.logger.Printf("Closing client: %s", reason)

	c.SetState(nil)

	c.hub.UnregisterChan <- c
	c.conn.Close()
	if _, closed := <-c.sendChannel; !closed {
		close(c.sendChannel)
	}
}
