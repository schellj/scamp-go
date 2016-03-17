GüdTech maintains a service oriented architecture powered by an in-house protocol called SCAMP (often refered to simply as gtsoa). The SCAMP protocol offers:

  * Discovery
  * Security
  * Quality of service

The spec has quite a few moving parts but the long and short of it: scamp uses persistent TLS connections between hosts and interleaves chunked messages to avoid head-of-line blocking issues. To those paying attention, that's very similar to how HTTP/2 works. Just many years earlier. A few key things to note with scamp: a) it's homegrown at GüdTech, b) it takes a stance on discovery, c) GüdTech maintains implementations for JavaScript, Perl, C# and now Golang.

This post will dig in to bugs related to scamp-go's client design and concurrency. We had some memory usage issues which quickly turned in to concurrency bugs and it seemed to useful to document in a casual blog post.

To understand how a single, persistent TLS connection handled interleaved message chunks we should consider this `Connection` definition:

    type Connection struct {
      conn           *tls.Conn
      incomingmsgno  uint64
      outgoingmsgno  uint64

      pktToMsg       map[uint64](*Message)
      msgs           chan *Message
    }

you can see here the Connection holds the tls connection and has a incoming and outgoing msgno. The msgno is used to keep both ends of the connection in sync; the msgno's both start at zero. In this implementation we wait until all parts of the message are received before delivering the message "up the stack" so we use `pktToMsg` as the nursery. When the message is fully formed we place it in the MessageChan.

This primitive message reconstruction flow is then wrapped by a `Client` whose job it is to differentiate incoming requests from incoming replies. Consider this `Client` definition:

    type Client struct {
      conn *Connection

      requests    MessageChan
      openReplies map[int]MessageChan
    }

A `Client` can be instantiated one of two ways:

  * Directly using a `Dial` function. You provide a `net.Conn`-style connection spec (.e.g., `"127.0.0.1:30100"`)
  * Spawned from a tls listening socket (e.g., a scamp service accepting incoming connections)

The service use case is the most interesting because concurrency becomes a big consideration. The `Service` is responsible for accepting new `Client`s, receiving their incoming `Messages`, and routing requests to registered functions, and also timing out long-lived, idle connections. Consider this `Service` definition:

    type ServiceActionFunc func(*Message, *Client)
    type ServiceAction struct {
      callback ServiceActionFunc
      // action metadata elided for simplicity's sake
    }

    type Service struct {
      listener      net.Listener

      actions       map[string]*ServiceAction
      clients       []*Client
    }

Over time this service will see many `Client`s which will have produced many different request/reply message pairs. Now that we have `Connection`/`Client`/`Service` all wired up we have to provide a rigorous system for cleaning up when:

  * a connection times out due to inactivity
  * a service shuts down
  * a client closes its connection (generally indistinguishable from a crash or clean shutdown on the opposite end)

This entire implementation is powered by go's builtin channels but along the way we have hit some bumps in our own understanding. Channels provide a simple way to signal between channels, you use the `aChan<-data` operator to send data and you use the `incomingData := <- aChan` to receive. Both send an receive are atomic with the ability to signal either side is done by calling `close(aChan)`. When your data flow is top-down channels are easier than a mutex because you won't forget to unlock them. The big problem? We were not closing channels resulting in zombie goroutines and inflated memory use!

So how do you shutdown a `Client` when there is contention from the `Service` (reading data) and the `Connection` (writing data)? The answer: you have to carefully sequence your `close(aChan)` calls and watch for data races. Consider these 3 shutdown procedures. Let's go over a naïve first pass:

    func (conn *Connection)Close() {
      conn.conn.Close()
    }

    func (client *Client)Close() {
      client.conn.Close()
      close(client.requests)
      for _,openReplyChan := range client.openReplies {
        close(openReplyChan)
      }
    }

    func (serv *Service)Stop(){
      serv.listener.Close()
      for _,client := range serv.clients {
        client.Close()
      }
    }

This seems straight-forward enough but it will cause a `panic` if we call `Client.close()` and `Connection.close()` -- this would result in calling `Connection.close()` twice and the go runtime will panic on a double close. There is also no way to interrogate a channel for its current state. You will also receive a panic if you try to send data on a closed channel. After working through these issues it's clear the golang creators intended channel closes to stay simple and be handled in a one-time manner.

Why would a `Client` and a `Connection` try to close at the same time? Well, they're used by different parts of the system: the `Client` is managed by the `Service` and the `Connection` can close at anytime if its underlying TLS connection goes away. Our live service will experience all manner of ordering on shutdowns so we will eventually trigger this scenario.

Enter the `sync.Mutex`. We'll protect state variable since we cannot interrogate channels.

    import "sync"
    type Connection struct {
      conn           *tls.Conn
      incomingmsgno  uint64
      outgoingmsgno  uint64

      pktToMsg       map[uint64](*Message)
      msgs           chan *Message

      isClosedM      sync.Mutex
      isClosed       bool
    }

    func (conn *Connection)Close() {
      conn.isClosedM.Lock()
      defer conn.isClosedM.Lock()
      if conn.isClosed {
        return
      }

      conn.conn.Close()

      conn.isClosed = true
    }

    type Client struct {
      conn *Connection

      requests    MessageChan
      openReplies map[int]MessageChan

      isClosedM   sync.Mutex
      isClosed    bool
    }
    func (client *Client)Close() {
      conn.isClosedM.Lock()
      defer conn.isClosedM.Lock()
      if conn.isClosed {
        client.conn.Close()  
      }
      
      close(client.requests)
      for _,openReplyChan := range client.openReplies {
        close(openReplyChan)
      }

      conn.isClosed = true
    }

    func (serv *Service)Stop(){
      serv.listener.Close()
      for _,client := range serv.clients {
        client.Close()
      }
    }

Yes, this ads more lines of code but now we can rest assured that we will not double-close a channel.