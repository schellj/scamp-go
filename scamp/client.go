package scamp

type ClientChan chan *Client

type Client struct {
  conn *Connection
  incoming chan *Message
}

func Dial(connspec string) (client *Client, err error){
  client = new(Client)

  conn,err := DialConnection(connspec)
  if err != nil {
    return
  }
  client.conn = conn

  return
}

func NewClient(conn *Connection) (client *Client){
  client = new(Client)

  client.conn = conn
  client.incoming = make(chan *Message)
  
  return
}

func (client *Client)Send(msg *Message) (err error){ 
  client.conn.Send(msg)
  return
}

func (client *Client)Close() {
  client.conn.Close()
}