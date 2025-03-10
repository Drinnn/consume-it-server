package pb

func NewChatPacket(message string) Msg {
	return &Packet_Chat{
		Chat: &ChatMessage{
			Msg: message,
		},
	}
}

func NewIdPacket(id uint64) Msg {
	return &Packet_Id{
		Id: &IdMessage{
			Id: id,
		},
	}
}
