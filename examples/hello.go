package main

import "github.com/gudtech/scamp-go/scamp"
import "fmt"
import "bytes"
import "time"

func main() {
	scamp.Initialize("/etc/SCAMP/soa.conf")

	client, err := scamp.Dial("127.0.0.1:30100")
  if err != nil {
    scamp.Error.Fatalf("could not dial to host")
  }
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

    responseChan, err := client.Send(message)
    if err != nil {
      scamp.Error.Printf("error sending msg: `%s`", err)
    }

    select {
    case msg := <- responseChan:
      scamp.Info.Printf("response: %s", msg.Bytes())
    case <- time.After(time.Second * 1):
      scamp.Error.Fatalf("did not receive response in time")
    }
  }
}