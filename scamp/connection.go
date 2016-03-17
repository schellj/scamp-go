package scamp

import (
	"crypto/tls"
	"fmt"
	"bufio"

	"sync/atomic"
	"sync"
	"strings"
)

type IncomingMsgNo uint64
type OutgoingMsgNo uint64

type Connection struct {
	conn         *tls.Conn
	Fingerprint  string

	reader         *bufio.Reader
	writer         *bufio.Writer
	incomingmsgno  IncomingMsgNo
	outgoingmsgno  OutgoingMsgNo

	pktToMsg       map[IncomingMsgNo](*Message)
	msgs           MessageChan

	client         *Client

	isClosed       bool
	closedMutex    sync.Mutex
}

// Used by Client to establish a secure connection to the remote service.
// TODO: You must use the *connection.Fingerprint to verify the
// remote host
func DialConnection(connspec string) (conn *Connection, err error) {
	Trace.Printf("dialing connection to `%s`", connspec)
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	config.BuildNameToCertificate()

	tlsConn, err := tls.Dial("tcp", connspec, config)
	if err != nil {
		return
	}

	conn = NewConnection(tlsConn)
	
	return
}

// Used by Service
func NewConnection(tlsConn *tls.Conn) (conn *Connection) {
	conn = new(Connection)
	conn.conn = tlsConn

	// TODO get the end entity certificate instead
	peerCerts := conn.conn.ConnectionState().PeerCertificates
	if len(peerCerts) == 1 {
		peerCert := peerCerts[0]
		conn.Fingerprint = sha1FingerPrint(peerCert)
	}

	conn.reader        = bufio.NewReader(conn.conn)
	conn.writer        = bufio.NewWriter(conn.conn)
	conn.incomingmsgno = 0
	conn.outgoingmsgno = 0

	conn.pktToMsg      = make(map[IncomingMsgNo](*Message))
	conn.msgs          = make(MessageChan)

	conn.isClosed      = false

	go conn.packetRouter()
	// go conn.packetAcker()

	return
}

func (conn *Connection) SetClient(client *Client) {
	conn.client = client
}

func (conn *Connection) packetRouter() (err error) {
	// Trace.Printf("starting packetrouter")
	var pkt *Packet
	var msg *Message

	// defer func(){
	// 	if conn.client != nil {
	// 		// Notify wrapper client we're dead
	// 		conn.client.Close()
	// 	}
	// }()

	for {
		Trace.Printf("reading packet...")
		readAttempt := make(chan *Packet)

		go func(){
			pkt,err := ReadPacket(conn.reader)

			if err != nil {
				if strings.Contains(err.Error(), "readline error: EOF") {
				} else if strings.Contains(err.Error(), "use of closed network connection") {
				} else if strings.Contains(err.Error(), "connection reset by peer") {
				} else {
					Error.Printf("err: %s", err)
				}
				close(readAttempt)
				return
			}

			conn.ackBytes(IncomingMsgNo(pkt.msgNo), uint64(len(pkt.body)))

			readAttempt <- pkt
		}()

		var ok bool
		select {
		case pkt,ok = <-readAttempt:
			if !ok {
				Trace.Printf("select statement got a closed channel. exiting packetRouter.")
				if conn.client != nil {
					conn.client.Close()
				}
				return
			}
		}


		Trace.Printf("switching...")
		switch {
			case pkt.packetType == HEADER:
				Trace.Printf("HEADER")
				// Allocate new msg
				// First verify it's the expected incoming msgno
				incomingmsgno := atomic.LoadUint64((*uint64)(&conn.incomingmsgno))
				if pkt.msgNo != incomingmsgno {
					err = fmt.Errorf("out of sequence msgno: expected %d but got %d", incomingmsgno, pkt.msgNo)
					Error.Printf("%s", err)
					return err
				}

				msg = conn.pktToMsg[IncomingMsgNo(pkt.msgNo)]
				if msg != nil {
					err = fmt.Errorf("Bad HEADER; already tracking msgno %d", pkt.msgNo)
					Error.Printf("%s", err)
					return err
				}

				// Allocate message and copy over header values so we don't have to track them
				// We copy out the packetHeader values and then we can discard it
				msg = NewMessage()
				msg.SetAction(pkt.packetHeader.Action)
				msg.SetEnvelope(pkt.packetHeader.Envelope)
				msg.SetVersion(pkt.packetHeader.Version)
				msg.SetMessageType(pkt.packetHeader.MessageType)
				msg.SetRequestId(pkt.packetHeader.RequestId)
				// TODO: Do we need the requestId?

				conn.pktToMsg[IncomingMsgNo(pkt.msgNo)] = msg
				// This is for sending out data
				// conn.incomingNotifiers[pktMsgNo] = &make((chan *Message),1)

				atomic.AddUint64((*uint64)(&conn.incomingmsgno), 1)
			case pkt.packetType == 	DATA:
				Trace.Printf("DATA")
				// Append data
				// Verify we are tracking that message
				msg = conn.pktToMsg[IncomingMsgNo(pkt.msgNo)]
				if msg == nil {
					err = fmt.Errorf("not tracking msgno %d", pkt.msgNo)
					Error.Printf("unexpected error: `%s`", err)
					return err
				}

				msg.Write(pkt.body)
			case pkt.packetType == EOF:
				Trace.Printf("EOF")
				// Deliver message
				msg = conn.pktToMsg[IncomingMsgNo(pkt.msgNo)]
				if msg == nil {
					err = fmt.Errorf("cannot process EOF for unknown msgno %d", pkt.msgNo)
					Error.Printf("err: `%s`", err)
					return
				}

				delete(conn.pktToMsg, IncomingMsgNo(pkt.msgNo))
				Trace.Printf("delivering msgno %d up the stack", pkt.msgNo)
				conn.msgs <- msg
			case pkt.packetType == 	TXERR:
				Trace.Printf("TXERR")
				delete(conn.pktToMsg, IncomingMsgNo(pkt.msgNo))
				conn.msgs <- msg
				// TODO: add 'error' path on connection
				// Kill connection
			case pkt.packetType == 	ACK:
				Trace.Printf("ACK `%s` (unackedbytes: %d)", pkt.body)
				// panic("Xavier needs to support this")
				// Add bytes to message stream tally
		}
	}
}

func (conn *Connection)Send(msg *Message) (err error) {
	if msg.RequestId == 0 {
		err = fmt.Errorf("must specify `ReqestId` on msg before sending")
		return
	}
	
	outgoingmsgno := atomic.LoadUint64((*uint64)(&conn.outgoingmsgno))
	atomic.AddUint64((*uint64)(&conn.outgoingmsgno),1)

	Trace.Printf("sending msgno %d", outgoingmsgno)

	for i,pkt := range msg.toPackets(outgoingmsgno) {
		Trace.Printf("sending pkt %d", i)
		_, err := pkt.Write(conn.writer)
		if err != nil {
			Error.Printf("error writing packet: `%s`", err)
			return err
		}
	}
	conn.writer.Flush()
	Trace.Printf("done sending msg")

	return
}

// func (conn *Connection)packetAcker() {
// 	timeout := time.Duration(15) * time.Second

// 	for {
// 		select {
// 		case <-conn.ackerShutdown:
// 			break
// 		case <-time.After(timeout):
// 			err := conn.ackBytes()
// 			if err != nil {
// 				Error.Printf("could not ack bytes: %s", err.Error())
// 			}
// 		}
// 	}
// }


func (conn *Connection)ackBytes(msgno IncomingMsgNo, unackedByteCount uint64) (err error) {
	ackPacket := Packet{
		packetType: ACK,
		msgNo: uint64(msgno),
		body: []byte(fmt.Sprintf("%d", unackedByteCount)),
	}

	_, err = ackPacket.Write(conn.writer)
	if err != nil {
		return err
	}

	return
}

func (conn *Connection)Close() {
	conn.closedMutex.Lock()
	if conn.isClosed {
		Trace.Printf("connection already closed. skipping shutdown.")
		conn.closedMutex.Unlock()
		return
	}


	Trace.Printf("connection is closing")

	conn.conn.Close()
	// close(conn.msgs) // hit a very rare bug where this was closed on insert

	conn.isClosed = true
	conn.closedMutex.Unlock()
}
