package pb

func NewChatPacket(senderID uint64, message string) *Packet {
	return &Packet{
		SenderId: senderID,
		Msg: &Packet_Chat{
			Chat: &ChatMessage{
				Msg: message,
			},
		},
	}
}

func NewIdPacket(senderID uint64, id uint64) *Packet {
	return &Packet{
		SenderId: senderID,
		Msg: &Packet_Id{
			Id: &IdMessage{
				Id: id,
			},
		},
	}
}
