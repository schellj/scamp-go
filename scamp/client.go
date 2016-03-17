package scamp

import (
  "sync"
)

type ClientChan chan *Client

type Client struct {
  conn *Connection
  serv *Service

  requests MessageChan
  openReplies map[int]MessageChan

  closeSplitReqsAndReps chan bool

  isClosed bool
  closedMutex sync.Mutex
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

  client.closeSplitReqsAndReps = make(chan bool)

  conn.SetClient(client)
  
  go client.splitReqsAndReps()

  return
}

func (client *Client)SetService(serv *Service) {
  client.serv = serv
}

func (client *Client)Send(msg *Message) (responseChan MessageChan, err error){ 
  // Info.Printf("sending message `%d`", msg.RequestId)
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
  client.closedMutex.Lock()
  if client.isClosed {
    Trace.Printf("client already closed. skipping shutdown.")
    client.closedMutex.Unlock()
    return
  }

  client.closeSplitReqsAndReps <- true

  // // Notify wrapper service that we're dead
  if client.serv != nil {
    client.serv.RemoveClient(client)
  }

  client.isClosed = true
  client.closedMutex.Unlock()
}

func (client *Client)splitReqsAndReps() (err error) {
  var replyChan MessageChan

  forLoop:
  for {
    select {
    case message := <-client.conn.msgs:
      Trace.Printf("splitting incoming message to reqs and reps")

      if message.MessageType == MESSAGE_TYPE_REQUEST {
        // interesting things happen if there are outstanding messages
        // and the client closes
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
    case <- client.closeSplitReqsAndReps:
      // Info.Printf("closing down SplitReqsAndReps")
      break forLoop
    }
  }

  Trace.Printf("done with SplitReqsAndReps")

  close(client.requests)
  for _,openReplyChan := range client.openReplies {
    close(openReplyChan)
  }

  return
}

func (client *Client)Incoming() (MessageChan) {
  return client.requests
}