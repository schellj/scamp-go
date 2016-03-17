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

	// reader         *bufio.Reader
	// writer         *bufio.Writer
	readWriter     *bufio.ReadWriter

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

	// conn.reader        = bufio.NewReader(conn.conn)
	// conn.writer        = bufio.NewWriter(conn.conn)
	conn.readWriter       = bufio.NewReadWriter(bufio.NewReader(conn.conn), bufio.NewWriter(conn.conn))
	conn.incomingmsgno = 0
	conn.outgoingmsgno = 0

	conn.pktToMsg      = make(map[IncomingMsgNo](*Message))
	conn.msgs          = make(MessageChan)

	conn.isClosed      = false

	go conn.packetReader()

	return
}

func (conn *Connection) SetClient(client *Client) {
	conn.client = client
}

func (conn *Connection) packetReader() (err error) {
	// Trace.Printf("starting packetrouter")
	var pkt *Packet

	PacketReaderLoop:
	for {
		Trace.Printf("reading packet...")

		pkt,err = ReadPacket(conn.readWriter)
		if err != nil {
			if strings.Contains(err.Error(), "readline error: EOF") {
			} else if strings.Contains(err.Error(), "use of closed network connection") {
			} else if strings.Contains(err.Error(), "connection reset by peer") {
			} else {
				Error.Printf("err: %s", err)
			}
			break PacketReaderLoop
		}
		err = conn.routePacket(pkt)
		if err != nil {
			break PacketReaderLoop
		}
	}

	// we close after routePacket is no longer possible
	// to avoid any send after close panics
	close(conn.msgs)

	return
}

func (conn *Connection) routePacket(pkt *Packet) (err error) {
		var msg *Message

		Trace.Printf("routing packet...")
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
				conn.ackBytes(IncomingMsgNo(pkt.msgNo), uint64(len(pkt.body)))
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
				Trace.Printf("ACK `%d` for msgno %d", pkt.msgNo, len(pkt.body))
				// panic("Xavier needs to support this")
				// Add bytes to message stream tally
		}

		return
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
		_, err := pkt.Write(conn.readWriter)
		if err != nil {
			Error.Printf("error writing packet: `%s`", err)
			return err
		}
	}
	conn.readWriter.Flush()
	Trace.Printf("done sending msg")

	return
}

func (conn *Connection)ackBytes(msgno IncomingMsgNo, unackedByteCount uint64) (err error) {
	ackPacket := Packet{
		packetType: ACK,
		msgNo: uint64(msgno),
		body: []byte(fmt.Sprintf("%d", unackedByteCount)),
	}

	_, err = ackPacket.Write(conn.readWriter)
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

	conn.isClosed = true
	conn.closedMutex.Unlock()
}
