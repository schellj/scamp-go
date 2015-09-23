package scamp

type msgNoType int64

type Message interface {
	toPackets() []Packet
  // setRequestId(reqId reqIdType)
}

type EOFResponse struct {}

func (msg *EOFResponse) setRequestId(reqId reqIdType) {}

func (msg EOFResponse) toPackets() []Packet {
	eofPacket := Packet{
		packetType:  EOF,
	}

	return []Packet{ eofPacket }
}