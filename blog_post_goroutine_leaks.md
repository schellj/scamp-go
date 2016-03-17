In the last post we detailed a Service/Client/Connection system which spawned a few goroutines with each new struct. This is normal as you need to service all the live connections in your application to handle incoming requests. In this post we'll talk about how to debug goroutines using SIGQUIT and the built in goroutine dump facility.

To quickly recap, we're talking about `scamp-go`, a library for GüdTech's homegrown SOA protocol. Key thing to remember: it multiplexes messages of long-lived TLS connections. Here are the most interesting structs (edited for brevity):

    type Connection struct {
      conn           *tls.Conn
      incomingmsgno  uint64
      outgoingmsgno  uint64

      pktToMsg       map[uint64](*Message)
      msgs           chan *Message

      isClosedM      sync.Mutex
      isClosed       bool
    }

    type Client struct {
      conn *Connection

      requests    MessageChan
      openReplies map[int]MessageChan

      isClosedM   sync.Mutex
      isClosed    bool
    }

    type Service struct {
      listener      net.Listener

      actions       map[string]*ServiceAction
      clients       []*Client
    }

and the the spots where we spawn goroutines to keep data flowing:

    func NewConnection(tlsConn *tls.Conn) (conn *Connection) {
      conn = new(Connection)

      // SNIP

      go conn.packetReader()

      return
    }

    func NewClient(conn *Connection) (client *Client){
      client = new(Client)
      
      go client.splitReqsAndReps()

      return
    }

    func (serv *Service)Run() {
      forLoop:
      for {
        netConn,err := serv.listener.Accept()
        if err != nil {
          break forLoop
        }

        var tlsConn (*tls.Conn) = (netConn).(*tls.Conn)
        if tlsConn == nil {
          break forLoop
        }

        conn := NewConnection(tlsConn)
        client := NewClient(conn)

        serv.clients = append(serv.clients, client)
        go serv.Handle(client)
      }
    }