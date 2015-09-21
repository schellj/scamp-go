package scamp

// import "math/rand"
// import "fmt"

// TODO: Requests should come out of a request object pool
// which sets their message_id on retrieval
type Request struct {
	Action         string
	EnvelopeFormat envelopeFormat
	Version        int64
	MessageId      msgNoType
	Blob           []byte
}

// var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// func (req *Request) GenerateMessageId() {
// 	// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
// 	b := make([]rune, 18)
// 	for i := range b {
// 		b[i] = letters[rand.Intn(len(letters))]
// 	}
// 	req.MessageId = string(b)
// }

func (req *Request) toPackets(msgNo msgNoType) []Packet {
	headerHeader := PacketHeader{
		Action:      req.Action,
		Envelope:    req.EnvelopeFormat,
		Version:     req.Version,
		MessageId:   msgNo,
		messageType: request,
	}
	
	headerPacket := Packet {
		packetHeader: headerHeader,
		packetType:   HEADER,
		msgNo:  msgNo,
	}

	eofPacket := Packet {
		packetType:  EOF,
		msgNo: msgNo,
	}


	if len(req.Blob) > 0 {
		dataPacket := Packet {
			packetType: DATA,
			msgNo: msgNo,
			body: req.Blob,
		}

		return []Packet{headerPacket, dataPacket, eofPacket}
	} else {
		return []Packet{headerPacket, eofPacket}
	}

}
