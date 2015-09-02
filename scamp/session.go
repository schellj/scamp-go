package scamp

import "bytes"
import "errors"

type Session struct {
	msgNo msgNoType
	conn *Connection
	packets []Packet
	replyChan (chan Message)
	requestChan (chan Request)
}

func newSession(newMsgNo msgNoType, conn *Connection) (sess *Session) {
	sess = new(Session)
	sess.msgNo = newMsgNo
	sess.conn = conn
	sess.replyChan = make(chan Message)
	sess.requestChan = make(chan Request)
	return
}

func (sess *Session) SendRequest(req Message) (err error) {
	pkts := req.toPackets(sess.msgNo)
	for _, pkt := range pkts {
		// Trace.Printf("sending msgNo %d", pkt.msgNo)
		err = pkt.Write(sess.conn.conn)
		if err != nil {
			return
		}
	}

	return
}

func (sess *Session) SendReply(rep Message) (err error) {
	pkts := rep.toPackets(sess.msgNo)
	for _, pkt := range pkts {
		Trace.Printf("SENDING (%d, %s)", pkt.msgNo, pkt.packetType)
		err = pkt.Write(sess.conn.conn)
		if err != nil {
			return
		}
	}

	return
}

func (sess *Session) RecvReply() (rep Reply, err error) {
	var ok bool
	msg := <-sess.replyChan
	if rep,ok = msg.(Reply); !ok {
		err = errors.New("did not receive Reply in RecvReply")
		return
	}
	return
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

	hdrPkt := sess.packets[0].packetHeader

	for _,pkt := range sess.packets {
		// TODO: should this be converted to use a buffer?
		bodyBlob = append(bodyBlob[:], pkt.body[:]...)
	}

	req := Request {
		Action: hdrPkt.Action,
		Blob: bodyBlob,
	}
	sess.requestChan <- req
}

func (sess *Session) CloseReply() {
	sess.SendReply(&EOFResponse{})
}

func (sess *Session) Free(){
	sess.conn.Free(sess)
}