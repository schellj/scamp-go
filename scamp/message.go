package scamp

import "bytes"

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

func (msg *Message)Write(blob []byte) (n int, err error){
  msg.packets = append(msg.packets, &Packet{packetType: DATA, body: blob})
  return len(blob), nil
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

  packets := make([]*Packet, 2)
  packets[0] = &headerPacket
  packets[1] = &eofPacket

  for _,dataPacket := range msg.packets {
    packets = append(packets, dataPacket)
  }

  return packets
}

func (msg *Message)Bytes() ([]byte) {
  buf := new(bytes.Buffer)
  for _,pkt := range msg.packets {
    buf.Write(pkt.body)
  }

  return buf.Bytes()
}