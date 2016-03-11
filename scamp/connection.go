package scamp

import (
	"crypto/tls"
	"fmt"
	"bufio"

	"time"
	"sync/atomic"
)

type Connection struct {
	conn         *tls.Conn
	Fingerprint  string

	reader         *bufio.Reader
	writer         *bufio.Writer
	incomingmsgno  uint64
	outgoingmsgno  uint64

	unackedbytes   uint64
	ackerShutdown  chan bool

	pktToMsg       map[int](*Message)
	msgs           MessageChan

	client         *Client
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

	conn.unackedbytes  = 0
	conn.ackerShutdown = make(chan bool)

	go conn.packetRouter()
	go conn.packetAcker()

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

			// Trace.Printf("read packet: %s", pkt)
			if err != nil {
				if err.Error() == "readline error: EOF" {
				} else {
					Error.Printf("err: %s", err)
				}
				// return fmt.Errorf("err reading packet: `%s`. (EOF is normal). Returning.", err)
				close(readAttempt)
			}

			readAttempt <- pkt
		}()

		var ok bool
		select {
		case pkt,ok = <-readAttempt:
			if !ok {
				Error.Printf("select statement got a closed channel. exiting packetRouter.")
				return
			}
		}


		Trace.Printf("switching...")
		switch {
			case pkt.packetType == HEADER:
				Trace.Printf("HEADER")
				// Allocate new msg
				// First verify it's the expected incoming msgno
				if uint64(pkt.msgNo) != conn.incomingmsgno {
					err = fmt.Errorf("out of sequence msgno: expected %d but got %d", conn.incomingmsgno, pkt.msgNo)
					Error.Printf("%s", err)
					return err
				}

				msg = conn.pktToMsg[pkt.msgNo]
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
					err = fmt.Errorf("not tracking msgno %d", pkt.msgNo)
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
	if msg.RequestId == 0 {
		err = fmt.Errorf("must specify `ReqestId` on msg before sending")
		return
	}
	
	outgoingmsgno := conn.outgoingmsgno
	atomic.AddUint64(&conn.outgoingmsgno,1)

	Trace.Printf("sending msgno %d", outgoingmsgno)

	for i,pkt := range msg.toPackets(int(outgoingmsgno)) {
		Trace.Printf("sending pkt %d", i)
		bytesWritten, err := pkt.Write(conn.writer)
		if err != nil {
			Error.Printf("error writing packet: `%s`", err)
			return err
		}

		atomic.AddUint64(&conn.unackedbytes, uint64(bytesWritten))
	}
	conn.writer.Flush()
	Trace.Printf("done sending msg")

	return
}

func (conn *Connection)packetAcker() {
	timeout := time.Duration(15) * time.Second

	for {
		select {
		case <-conn.ackerShutdown:
			break
		case <-time.After(timeout):
			err := conn.ackBytes()
			if err != nil {
				Error.Printf("could not ack bytes: %s", err.Error())
			}
		}
	}
}


// type Packet struct {
// 	packetType   PacketType
// 	msgNo        int
// 	packetHeader PacketHeader
// 	body         []byte
// 	ackRequestId int
// }
func (conn *Connection)ackBytes() (err error) {
	if conn.unackedbytes == 0 {
		return
	}

	theseUnackedBytes := conn.unackedbytes

	outgoingmsgno := conn.outgoingmsgno
	atomic.AddUint64(&conn.outgoingmsgno,1)

	ackPacket := Packet{
		packetType: ACK,
		msgNo: int(outgoingmsgno),
		body: []byte(fmt.Sprintf("%d", theseUnackedBytes)),
	}

	_, err = ackPacket.Write(conn.writer)
	if err != nil {
		Error.Printf("error writing ack packet: `%s`", err)
		return err
	}

	atomic.AddUint64(&conn.unackedbytes, -theseUnackedBytes)

	return
}

func (conn *Connection)Close() {
	Trace.Printf("connection is closing")
	conn.ackerShutdown <- true

	conn.conn.Close()
	close(conn.msgs)
}
