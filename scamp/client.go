package scamp

type ClientChan chan *Client

type Client struct {
  conn *Connection
  serv *Service

  requests MessageChan
  openReplies map[int]MessageChan
  doneChan chan bool
}

func Dial(connspec string) (client *Client, err error){
  Trace.Printf("connecting to `%s`", connspec)

  conn,err := DialConnection(connspec)
  client = NewClient(conn)
  if err != nil {
    return
  }
  client.conn = conn

  return
}

func NewClient(conn *Connection) (client *Client){
  Trace.Printf("client allocated")
  client = new(Client)

  client.conn = conn
  client.requests = make(MessageChan)
  client.openReplies = make(map[int]MessageChan)
  client.doneChan = make(chan bool)

  conn.SetClient(client)
  
  go client.SplitReqsAndReps()

  return
}

func (client *Client)SetService(serv *Service) {
  client.serv = serv
}

func (client *Client)Send(msg *Message) (responseChan MessageChan, err error){ 
  Trace.Printf("sending message `%d`", msg.RequestId)
  err = client.conn.Send(msg)
  if err != nil {
    return
  }

  if msg.MessageType == MESSAGE_TYPE_REQUEST {
    Trace.Printf("sending request so waiting for reply")
    responseChan = make(MessageChan)
    client.openReplies[msg.RequestId] = responseChan
  } else {
    Trace.Printf("sending reply so done with this message")
  }

  return
}

func (client *Client)Close() {
  client.conn.Close()

  close(client.requests)
  for _,openReplyChan := range client.openReplies {
    close(openReplyChan)
  }

  // Notify wrapper service that we're dead
  if client.serv != nil {
    client.serv.RemoveClient(client)
  }

  client.doneChan <- true
}

func (client *Client)Done() (chan bool) {
  return client.doneChan
}

func (client *Client)SplitReqsAndReps() (err error) {
  var replyChan MessageChan

  for message := range client.conn.msgs {
    Trace.Printf("splitting incoming message to reqs and reps")

    if message.MessageType == MESSAGE_TYPE_REQUEST {
      client.requests <- message
    } else if message.MessageType == MESSAGE_TYPE_REPLY {
      replyChan = client.openReplies[message.RequestId]

      if replyChan == nil {
        Error.Printf("got an unexpected reply for requestId: %d. Skipping.", message.RequestId)
        continue
      }

      delete(client.openReplies, message.RequestId)
      replyChan <- message

    } else {
      Error.Printf("Could not handle msg, it's neither req or reply. Skipping.")
      continue
    }

  }

  Trace.Printf("done with SplitReqsAndReps")

  return
}

func (client *Client)Incoming() (MessageChan) {
  return client.requests
}