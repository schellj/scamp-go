package main

import "scamp"
import "fmt"
import "bytes"

func main() {
	scamp.Initialize()

	client, err := scamp.Dial("127.0.0.1:30100")
	defer client.Close()

  for i := 1; i <= 10; i++ {
    buf := new(bytes.Buffer)
    message := scamp.NewMessage()
    message.SetAction("helloworld.hello")
    message.SetEnvelope(scamp.ENVELOPE_JSON)
    message.SetVersion(1)
    message.SetMessageType(scamp.MESSAGE_TYPE_REQUEST)

    fmt.Fprintf(buf, "sup %d", i)
    message.Write(buf.Bytes())

    err = client.Send(message)

    if err != nil {
      scamp.Error.Printf("error sending msg: `%s`", err)
    }

  }

  response := <- client.Incoming()
  scamp.Info.Printf("response: %s", response)
}