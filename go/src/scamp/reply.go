package scamp

import "errors"
import "io"
import "fmt"
import "bytes"

type Reply struct {
	blob []byte
}

func (rep *Reply) Read(reader io.Reader) (err error) {
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

	Info.Printf("Neat. Read %d packets. Merging.\n", len(packets))
	for i, pkt := range packets {
		Info.Printf( "Packet[%d] (%d bytes): `%s`\n", i, pkt.body, len(pkt.body) )
		mergeBuffer.Write(pkt.body)
	}

	rep.blob = mergeBuffer.Bytes()
	Info.Printf( "Final buffer size: %d\n", len(rep.blob))


	return
}

func (rep *Reply) Body() (body []byte) {
	body = rep.blob
	return
}
