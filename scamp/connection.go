package scamp

import "crypto/tls"
import "fmt"
import "bufio"

type Connection struct {
	conn         *tls.Conn
	Fingerprint  string

	reader         *bufio.Reader
	writer         *bufio.Writer
	incomingmsgno  int
	outgoingmsgno  int

	unackedbytes   int

	pktToMsg       map[int](*Message)
	msgs           MessageChan
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

	conn.pktToMsg      = make(map[int](*Message))
	conn.msgs          = make(MessageChan)

	go conn.packetRouter()

	return
}

func (conn *Connection) packetRouter() (err error) {
	Trace.Printf("starting packetrouter")
	var pkt *Packet
	var msg *Message

	for {
		Trace.Printf("reading packet...")
		pkt,err = ReadPacket(conn.reader)
		Trace.Printf("read packet: %s", pkt)
		if err != nil {
			Error.Printf("err: %s", err)
			return fmt.Errorf("err reading packet: `%s`. (EOF is normal). Returning.", err)
		}

		Trace.Printf("switching...")
		switch {
			case pkt.packetType == HEADER:
				Trace.Printf("HEADER")
				// Allocate new msg
				// First verify it's the expected incoming msgno
				if pkt.msgNo != conn.incomingmsgno {
					err = fmt.Errorf("out of sequence msgno: expected %d but got %d", conn.incomingmsgno, pkt.msgNo)
					Error.Printf("%s", err)
					return err
				}

				msg = conn.pktToMsg[pkt.msgNo]
				if msg != nil {
					err = fmt.Errorf("Bad HEADER; already tracking msgno %d")
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
				// TODO: Do we need the requestId?

				conn.pktToMsg[pkt.msgNo] = msg
				// This is for sending out data
				// conn.incomingNotifiers[pktMsgNo] = &make((chan *Message),1)

				conn.incomingmsgno = conn.incomingmsgno + 1
			case pkt.packetType == 	DATA:
				Trace.Printf("DATA")
				// Append data
				// Verify we are tracking that message
				msg = conn.pktToMsg[pkt.msgNo]
				if msg == nil {
					err = fmt.Errorf("not tracking msgno %d (%s)", pkt.msgNo, pkt)
					Error.Printf("unexpected error: `%s`", err)
					return err
				}

				msg.Write(pkt.body)
			case pkt.packetType == EOF:
				Trace.Printf("EOF")
				// Deliver message
				msg = conn.pktToMsg[pkt.msgNo]
				if msg == nil {
					err = fmt.Errorf("cannot process EOF for unknown msgno %d", pkt.msgNo)
					Error.Printf("err: `%s`", err)
					return
				}

				delete(conn.pktToMsg, pkt.msgNo)
				Trace.Printf("delivering msgno %d up the stack", pkt.msgNo)
				conn.msgs <- msg
			case pkt.packetType == 	TXERR:
				Trace.Printf("TXERR")
				delete(conn.pktToMsg, pkt.msgNo)				
				conn.msgs <- msg
				// TODO: add 'error' path on connection
				// Kill connection
			case pkt.packetType == 	ACK:
				Trace.Printf("ACK `%s` (unackedbytes: %d)", pkt.body, conn.unackedbytes)
				// panic("Xavier needs to support this")
				// Add bytes to message stream tally
		}
	}
}

func (conn *Connection)Send(msg *Message) (err error) {
	outgoingmsgno := conn.outgoingmsgno
	conn.outgoingmsgno = conn.outgoingmsgno + 1

	Trace.Printf("sending msgno %d", outgoingmsgno)

	for i,pkt := range msg.toPackets(outgoingmsgno) {
		Trace.Printf("sending pkt %d (%s)", i, pkt)
		bytesWritten, err := pkt.Write(conn.writer)
		if err != nil {
			Error.Printf("error writing packet: `%s`", err)
			return err
		}

		conn.unackedbytes = conn.unackedbytes+bytesWritten

	}
	conn.writer.Flush()
	Trace.Printf("done sending msg")

	return
}

func (conn *Connection)Close() {
	conn.conn.Close()
}