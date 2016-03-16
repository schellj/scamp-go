package main

import (
	"scamp"
	"time"
)

var timeout time.Duration = time.Duration(300) * time.Second

func main() {
	scamp.Initialize()

	client, err := scamp.Dial("127.0.0.1:30100")
	defer client.Close()
	if err != nil {
		scamp.Error.Fatalf("could not connect! `%s`\n", err)
		return
	}

  message := scamp.NewMessage()
  message.SetRequestId(1234)
  message.SetAction("Logger.log")
  message.SetEnvelope(scamp.ENVELOPE_JSON)
  message.SetVersion(1)
  message.SetMessageType(scamp.MESSAGE_TYPE_REQUEST)
  message.Write([]byte(`hey logger`))

  recvChan,err := client.Send(message)
  if err != nil {
  	scamp.Error.Fatalf("could not send message: `%s`\n", err)
  	return
  }

  select {
  case response,ok := <-recvChan:
  	if !ok {
  		scamp.Error.Printf("recvChan was closed. exiting.")
  	} else {
  		scamp.Info.Printf("got reply: %s", response.Bytes())
  	}
  	return
  case <-time.After(timeout):
  	scamp.Error.Fatalf("failed to get reply before timeout")
  	return
  }
}