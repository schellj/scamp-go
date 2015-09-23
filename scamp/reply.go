package scamp

import "errors"
import "fmt"
import "bytes"
import "bufio"

type Reply struct {
	requestId reqIdType
	Blob []byte
}

func (rep *Reply) Read(reader *bufio.Reader) (err error) {
	var packet Packet;
	var packets []Packet;

	for {
		packet, err = ReadPacket(reader)
		if err != nil {
			err = errors.New(fmt.Sprintf("err reading packet: `%s`", err))
			return
		}
		if packet.packetType == EOF || packet.packetType == TXERR {
			break
		} else if packet.packetType != DATA {
			continue
		}
		packets = append(packets, packet)
	}

	var mergeBuffer bytes.Buffer

	for _, pkt := range packets {
		mergeBuffer.Write(pkt.body)
	}

	rep.Blob = mergeBuffer.Bytes()

	return
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
