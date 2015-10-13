package scamp

type MessageChan chan *Message

type Message struct {
  Action  string
  Envelope envelopeFormat
  // RequestId reqIdType // TODO: how do RequestId's fit in again?
  Version int64
  MessageType messageType
  packets []*Packet
}


func NewMessage() (msg *Message) {
  msg = new(Message)
  return
}

func (msg *Message)SetAction(action string) {
  msg.Action = action
}

func (msg *Message)SetEnvelope(env envelopeFormat) {
  msg.Envelope = env
}

func (msg *Message)SetVersion(version int64) {
  msg.Version = version
}

func (msg *Message)SetMessageType(mtype messageType) {
  msg.MessageType = mtype
}

func (msg *Message)Write(blob []Byte) (n int, err error){
  
  msg.packets = append(msg.packets, pkt)
}

func (msg *Message)toPackets() ([]*Packet) {
  headerHeader := PacketHeader{
    Action:      msg.Action,
    Envelope:    msg.Envelope,
    Version:     msg.Version,
    RequestId:   1, // TODO: nope, can't do this
    MessageType: msg.MessageType,
  }
  
  headerPacket := Packet {
    packetHeader: headerHeader,
    packetType:   HEADER,
  }

  eofPacket := Packet {
    packetType:  EOF,
  }


  if len(msg.packets) > 0 {
    dataPacket := Packet {
      packetType: DATA,
      body: req.Blob,
    }

    return []Packet{headerPacket, dataPacket, eofPacket}
  } else {
    return []Packet{headerPacket, eofPacket}
  }
}