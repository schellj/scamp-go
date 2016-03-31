package main

import "github.com/gudtech/scamp-go/scamp"

var famous_words = []byte("SCAMP SHOUTS `HELLO WORLD`")

func main() {
	scamp.Initialize()

	service,err := scamp.NewService(":30100", "helloworld")
	if err != nil {
		scamp.Error.Fatalf("error creating new service: `%s`", err)
	}
	service.Register("helloworld.hello", func(message *scamp.Message, client *scamp.Client){
		scamp.Info.Printf("go message: `%s`", message.Bytes())
		blob := message.Bytes()

		if len(blob) > 0 {
			scamp.Info.Printf("helloworld had data: %s", blob)
		} else {
			scamp.Trace.Printf("helloworld was called without data")
		}

		reply := &scamp.Message{MessageType: scamp.MESSAGE_TYPE_REPLY}
		reply.Write(famous_words)
		reply.SetRequestId(message.RequestId)
		_, err := client.Send(reply)
		if err != nil {
			scamp.Error.Printf("error while sending reply: `%s`. continuing.", err)
			return
		}
	})

	announcer,err := scamp.NewDiscoveryAnnouncer()
	if err != nil {
		scamp.Error.Printf("failed to create announcer: `%s`", err)
		return
	}
	announcer.Track(service)

	go announcer.AnnounceLoop()

	service.Run()
}