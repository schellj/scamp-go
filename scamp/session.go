package scamp

import "bytes"
import "fmt"

type Session struct {
	conn *Connection
	requestId reqIdType
	packets []Packet
	replyChan (chan Message)
	requestChan (chan Request)
}

func newSession(requestId reqIdType, conn *Connection) (sess *Session) {
	sess = new(Session)
	sess.conn = conn
	sess.replyChan = make(chan Message)
	sess.requestChan = make(chan Request)
	sess.requestId = requestId
	return
}

func (sess *Session) Send(msg Message) (err error) {
	switch t := (msg).(type) {
	case Reply:
		(&t).setRequestId(sess.requestId)
		sess.requestId = sess.requestId + 1
		err = sess.conn.Send(&t)
	default:
		panic("should only be sending Replies...")
	}


	return
}

func (sess *Session) RecvReply() (rep Reply, err error) {
	msg := <-sess.replyChan

	switch t := (msg).(type) {
	case Reply:
		return t, nil
	default:
		return
	}
}

func (sess *Session) RecvChan() (chan Message) {
	return sess.replyChan
}

func (sess *Session) RecvRequest() (req Request, err error){
	req = <- sess.requestChan
	return
}

func (sess *Session) Append(pkt Packet) {
	sess.packets = append(sess.packets, pkt)
}

func (sess *Session) DeliverReply() {
	dataBuf := new(bytes.Buffer)
	if len(sess.packets) > 0 {
		for _,pkt := range sess.packets[1:] {
			dataBuf.Write(pkt.body)
		}
	}

	rep := Reply {
		Blob: dataBuf.Bytes(),
	}

	sess.replyChan <- rep
}

func (sess *Session) DeliverRequest() {
	var bodyBlob []byte

	if len(sess.packets) == 0 {
		panic( fmt.Sprintf( "trying to deliver without packets. sess: %s", sess) )
	}

	hdrPkt := sess.packets[0].packetHeader

	for _,pkt := range sess.packets {
		// TODO: should this be converted to use a buffer?
		bodyBlob = append(bodyBlob[:], pkt.body[:]...)
	}

	req := Request {
		Action: hdrPkt.Action,
		Blob: bodyBlob,
		RequestId: hdrPkt.RequestId,
	}
	sess.requestChan <- req

	sess.conn.Send(ACKResponse{ requestId: hdrPkt.RequestId, })
}

func (sess *Session) CloseReply() {
	sess.Send(&EOFResponse{})
}

func (sess *Session) Free(){
	sess.conn.Free(sess)
}