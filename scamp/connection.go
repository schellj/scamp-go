package scamp

import "crypto/tls"
import "fmt"
import "sync"
import "bufio"

type Connection struct {
	conn        *tls.Conn
	reader      *bufio.Reader
	Fingerprint string
	msgCnt      msgNoType

	sessDemuxMutex *sync.Mutex
	sessDemux    map[msgNoType](*Session)
	newSessions  (chan *Session)
}

// Establish secure connection to remote service.
// You must use the *connection.Fingerprint to verify the
// remote host
func Connect(connspec string) (conn *Connection, err error) {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	config.BuildNameToCertificate()

	tlsConn, err := tls.Dial("tcp", connspec, config)
	if err != nil {
		return
	}

	sessChan := make(chan *Session, 100)

	conn,err = newConnection(tlsConn, sessChan)
	if err != nil {
		return
	}
	go conn.packetRouter(true, false)
	
	return
}

func newConnection(tlsConn *tls.Conn, sessChan (chan *Session)) (conn *Connection, err error) {
	conn = new(Connection)
	conn.conn = tlsConn
	conn.reader = bufio.NewReader(conn.conn)

	conn.sessDemuxMutex = new(sync.Mutex)
	conn.sessDemux = make(map[msgNoType](*Session))
	conn.newSessions = sessChan

	// TODO get the end entity certificate instead
	peerCerts := conn.conn.ConnectionState().PeerCertificates
	if len(peerCerts) == 1 {
		peerCert := peerCerts[0]
		conn.Fingerprint = sha1FingerPrint(peerCert)
	}

	return
}

// Demultiplex packets to their proper buffers.
func (conn *Connection) packetRouter(ignoreUnknownSessions bool, isService bool) (err error) {
	var pkt Packet
	var sess *Session

	for {
		pkt, err = ReadPacket(conn.reader)
		Trace.Printf("received packet: %s", pkt)
		if err != nil {
			// TODO: what are the issues with stopping a packet router here?
			// The socket has probably closed
			return fmt.Errorf("err reading packet: `%s`. (EOF is normal). Returning.", err)
		}

		conn.sessDemuxMutex.Lock()
		sess = conn.sessDemux[pkt.msgNo]
		if sess == nil && !ignoreUnknownSessions {
			sess = newSession(pkt.msgNo, conn)
			conn.sessDemux[pkt.msgNo] = sess
			conn.sessDemuxMutex.Unlock()
			conn.newSessions <- sess // Could block and holding the DemuxMutex would block other tasks (namely: sending requests)
		} else {
			conn.sessDemuxMutex.Unlock()
		}

		if sess == nil && ignoreUnknownSessions {
			return fmt.Errorf("packet (msgNo: %d) has no corresponding session. killing connection.", pkt.msgNo)
		}

		if pkt.packetType == HEADER {
			Trace.Printf("(incoming SESS %d) HEADER packet\n", pkt.msgNo)
			sess.Append(pkt)
		} else if pkt.packetType == DATA {
			Trace.Printf("(incoming SESS %d) DATA packet (%d bytes)\n", pkt.msgNo, len(pkt.body))
			sess.Append(pkt)
		} else if pkt.packetType == EOF {
			Trace.Printf("(incoming SESS %d) EOF packet\n", pkt.msgNo)
			// TODO: need polymorphism on Req/Reply so they can be delivered
			if isService {
				Trace.Printf("session delivering request")
				sess.DeliverRequest()
			} else {
				Trace.Printf("session delivering reply")
				go sess.DeliverReply()
			}
		} else if pkt.packetType == TXERR {
			Trace.Printf("(incoming SESS %d) TXERR\n`%s`", pkt.msgNo, pkt.body)
			// TODO: need polymorphism on Req/Reply so they can be delivered
			if isService {
				sess.DeliverRequest()
			} else {
				go sess.DeliverReply()
			}
		} else {
			Trace.Printf("(incoming SESS %d) unknown packet type %d\n", pkt.packetType)
		}
	}

	return
}

func (conn *Connection) NewSession() (sess *Session, err error) {
	sess = new(Session)

	sess.conn = conn

	sess.msgNo = conn.msgCnt
	conn.msgCnt = conn.msgCnt + 1

	sess.replyChan = make(chan Message, 1)

	conn.sessDemux[sess.msgNo] = sess

	return
}

func (conn *Connection) Send(msg Message) (sess *Session, err error) {
	// The lock must be held until the first packet is sent. 
	// With the current structure it will hold the lock until all
	// packets for req are sent
	conn.sessDemuxMutex.Lock()
	sess,err = conn.NewSession()
	if err != nil {
		return
	}

	err = sess.SendRequest(msg)
	if req,ok := msg.(*Request); ok {
		Trace.Printf("sending request (%d bytes)", len(req.Blob) )
	} else {
		Trace.Printf("sending unknown value on connection")
	}
	if err != nil {
		return
	}
	conn.sessDemuxMutex.Unlock()

	return
}

// Pulls full Requests out of master Request chan
// func (conn *Connection) Recv() Session {
// 	return <-conn.sessionChan
// }

func (conn *Connection) Close() {
	conn.conn.Close()
}

func (conn *Connection) Recv() (sess *Session) {
	sess = <-conn.newSessions
	return
}

func (conn *Connection) Free(sess *Session) {
	conn.sessDemuxMutex.Lock()
	msgNo := sess.packets[0].msgNo
	delete(conn.sessDemux, msgNo)
	conn.sessDemuxMutex.Unlock()
}
