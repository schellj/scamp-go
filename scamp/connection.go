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

	conn.msgCnt = 0

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
		if err != nil {
			// TODO: what are the issues with stopping a packet router here?
			// The socket has probably closed
			return fmt.Errorf("err reading packet: `%s`. (EOF is normal). Returning.", err)
		}

		conn.sessDemuxMutex.Lock()
		sess = conn.sessDemux[pkt.msgNo]
		if sess == nil && !ignoreUnknownSessions {
			// TODO only have useful packetHeader on HEADER packets... should check that huh?
			sess = newSession(pkt.packetHeader.RequestId, conn)
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
			sess.Append(pkt)
		} else if pkt.packetType == DATA {
			sess.Append(pkt)
		} else if pkt.packetType == EOF {
			// TODO: need polymorphism on Req/Reply so they can be delivered
			if isService {
				if len(sess.packets) > 0 {
					sess.DeliverRequest()
				} else {
					Error.Printf("sess msgno: %d had no packets on EOF. Freeing and moving on.", pkt.msgNo)
				}
				conn.Free(pkt.msgNo)
			} else {
				go func(){
					sess.DeliverReply()
					conn.Free(pkt.msgNo)
				}()
			}
		} else if pkt.packetType == TXERR {
			Trace.Printf("(incoming SESS %d) TXERR\n`%s`", pkt.msgNo, pkt.body)
			// TODO: need polymorphism on Req/Reply so they can be delivered
			if isService {
				sess.DeliverRequest()
				conn.Free(pkt.msgNo)
			} else {
				go func(){
					sess.DeliverReply()
					conn.Free(pkt.msgNo)
				}()
			}
		} else if (pkt.packetType == ACK) {
			// TODO we should use this to cancel a timer on the Message
		} else {
			Error.Printf("(incoming SESS %d) unknown packet type %d\n", pkt.packetType)
		}
	}

	return
}

func (conn *Connection) Send(msg Message) (err error) {
	// The lock must be held until the first packet is sent. 
	// With the current structure it will hold the lock until all
	// packets for req are sent
	conn.sessDemuxMutex.Lock()
	for _,pkt := range msg.toPackets() {
		pkt.msgNo = conn.msgCnt

		// TODO: RequestId should be allocated on Reply allocation, not Reply send
		if pkt.packetType == HEADER {
			// Trace.Printf("HEADER %d, reqId: %d", pkt.msgNo, pkt.packetHeader.RequestId)
		} else if pkt.packetType == DATA {
			// Trace.Printf("DATA: %d", pkt.msgNo)
		} else if pkt.packetType == EOF {
			conn.msgCnt = conn.msgCnt + 1
		}

		err = pkt.Write(conn.conn)
		if err != nil {
			return
		}
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

func (conn *Connection) Free(msgNo msgNoType) {
	conn.sessDemuxMutex.Lock()
	delete(conn.sessDemux, msgNo)
	conn.sessDemuxMutex.Unlock()
}
