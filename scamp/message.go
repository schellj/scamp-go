package scamp

type msgNoType int64

type Message interface {
	toPackets(msgNoType) []Packet
}

type EOFResponse struct {}

func (msg EOFResponse) toPackets(msgNo msgNoType) []Packet {
	eofPacket := Packet{
		packetType:  EOF,
		msgNo: msgNo,
	}

	return []Packet{ eofPacket }
}