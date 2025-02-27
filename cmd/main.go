package main

import (
	"fmt"

	"github.com/Drinnn/consume-it/pb"
	"google.golang.org/protobuf/proto"
)

func main() {
	packet := pb.NewChatPacket(1, "Hello, World!")

	data, err := proto.Marshal(packet)
	if err != nil {
		fmt.Println("Error marshalling packet:", err)
		return
	}

	fmt.Println("Marshalled packet:", data)
}
