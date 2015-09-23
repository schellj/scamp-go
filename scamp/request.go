package scamp

// import "math/rand"
// import "fmt"

// TODO: Requests should come out of a request object pool
// which sets their message_id on retrieval
type Request struct {
	Action         string
	EnvelopeFormat envelopeFormat
	Version        int64
	RequestId      reqIdType
	Blob           []byte
}

func (req Request) setRequestId(reqId reqIdType) {
	req.RequestId = reqId
}

func (req Request) toPackets() []Packet {
	headerHeader := PacketHeader{
		Action:      req.Action,
		Envelope:    req.EnvelopeFormat,
		Version:     req.Version,
		RequestId:   req.RequestId,
		MessageType: request,
	}
	
	headerPacket := Packet {
		packetHeader: headerHeader,
		packetType:   HEADER,
	}

	eofPacket := Packet {
		packetType:  EOF,
	}


	if len(req.Blob) > 0 {
		dataPacket := Packet {
			packetType: DATA,
			body: req.Blob,
		}

		return []Packet{headerPacket, dataPacket, eofPacket}
	} else {
		return []Packet{headerPacket, eofPacket}
	}

}
