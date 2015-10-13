package scamp

type Client struct {
  conn *Connection
  msgs MessageChan
}

func Dial(connspec string) (client *Client, err error){
  client = new(Client)
  client.msgs = make(MessageChan)

  conn,err := DialConnection(connspec, client.msgs)
  if err != nil {
    return
  }
  client.conn = conn

  return
}

func (client *Client)Send(msg *Message) (err error){ 
  client.conn.Send(msg)
  return
}

func (client *Client)Close() {
  client.conn.Close()
}