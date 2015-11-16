package main

import "scamp"

func main() {
	scamp.Initialize()

	client, err := scamp.Dial("127.0.0.1:30100")
	defer client.Close()

	message := scamp.NewMessage()
	message.SetAction("helloworld.hello")
  message.SetEnvelope(scamp.ENVELOPE_JSON)
  message.SetVersion(1)
  message.SetMessageType(scamp.MESSAGE_TYPE_REQUEST)
  message.Write([]byte("sup"))

	err = client.Send(message)
	if err != nil {
		scamp.Error.Printf("error sending msg: `%s`", err)
	}

  response := <- client.Incoming()
  scamp.Info.Printf("response: %s", response)
}