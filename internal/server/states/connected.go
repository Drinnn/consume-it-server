package states

import (
	"fmt"
	"log"

	"github.com/Drinnn/consume-it/internal/server"
	"github.com/Drinnn/consume-it/pb"
)

type Connected struct {
	client server.ClientInterfacer
	logger *log.Logger
}

func (c *Connected) Name() string {
	return "Connected"
}

func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	loggingPrefix := fmt.Sprintf("Client %d [%s]: ", c.client.Id(), c.Name())
	c.logger = log.New(log.Writer(), loggingPrefix, log.LstdFlags)
}

func (c *Connected) OnEnter() {
	c.client.SocketSend(pb.NewIdPacket(c.client.Id()))
}

func (c *Connected) OnExit() {}

func (c *Connected) HandleMessage(senderId uint64, message pb.Msg) {
	if senderId == c.client.Id() {
		c.client.Broadcast(message)
	} else {
		c.client.SocketSendAs(message, senderId)
	}
}
