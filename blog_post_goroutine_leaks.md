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

so we spawn a `packetReader` for each `Connection`, a `splitReqsAndReps` for each `Client` and a `Handler` for each `Client` in the `Service`.

The issue which triggered this exercise: my `serv.clients` never seemed to go down. Even when I had many clients coming and going I found that the numbers only went up. It looked like those spawned goroutines for the clients were not exiting or the shutdown accounting had a bug. Good thing I had an opportunity to use the `SIGQUIT` dump I had heard about.

The `SIGQUIT` handler is built in to all go programs, you do not need to opt-in. It will exit the program and dump all the stacks of the goroutines (much like a panic). The great part about goroutine stacks is they track their spawn locations as well so you get a 