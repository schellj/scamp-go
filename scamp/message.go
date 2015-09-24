package scamp

type msgNoType int64

type Message interface {
	toPackets() []Packet
}

type EOFResponse struct {}

func (msg EOFResponse) toPackets() []Packet {
	eofPacket := Packet{
		packetType:  EOF,
	}

	return []Packet{ eofPacket }
}

type ACKResponse struct {
  requestId reqIdType
}

func (msg ACKResponse) toPackets() []Packet {
  ackPacket := Packet {
    packetType: ACK,
    ackRequestId: msg.requestId,
  }

  return []Packet{ ackPacket }
}