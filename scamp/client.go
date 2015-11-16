package scamp

type ClientChan chan *Client

type Client struct {
  conn *Connection
  incoming MessageChan
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
  client.incoming = make(chan *Message)
  
  return
}

func (client *Client)Send(msg *Message) (err error){ 
  Trace.Printf("sending message `%s`", msg)
  client.conn.Send(msg)
  return
}

func (client *Client)Close() {
  client.conn.Close()
}

func (client *Client)Incoming() (MessageChan) {
  return client.incoming
}