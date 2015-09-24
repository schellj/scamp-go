package scamp

type Reply struct {
	requestId reqIdType
	Blob []byte
}

func (rep *Reply) setRequestId(reqId reqIdType) {
	rep.requestId = reqId
}

func (rep Reply) toPackets() []Packet {
	headerHeader := PacketHeader{
		MessageType: reply,
		RequestId: rep.requestId,
	}

	headerPacket := Packet{
		packetHeader: headerHeader,
		packetType:   HEADER,
	}

	dataPacket := Packet{
		packetType: DATA,
		body: rep.Blob,
	}

	eofPacket := Packet{
		packetType:  EOF,
	}

	return []Packet{headerPacket, dataPacket, eofPacket}
}

func (rep *Reply) Body() (body []byte) {
	body = rep.Blob
	return
}
