package main

import "scamp"

var famous_words = []byte("SCAMP SAYS HELLO WORLD")

func main() {
	scamp.Initialize()

	service,err := scamp.NewService(":30100", "helloworld")
	if err != nil {
		scamp.Error.Fatalf("error creating new service: `%s`", err)
	}
	service.Register("helloworld.hello", func(message *scamp.Message, client *scamp.Client){
		scamp.Info.Printf("go message: `%s`", message.Bytes())
		req := <- client.Incoming()
		blob := req.Bytes()
		if len(blob) > 0 {
			scamp.Info.Printf("helloworld had data: %s", blob)
		} else {
			scamp.Trace.Printf("helloworld was called without data")
		}

		reply := &scamp.Message{}
		reply.Write(famous_words)
		err = client.Send(reply)
		if err != nil {
			scamp.Error.Printf("error while sending reply: `%s`. continuing.", err)
			return
		}
		scamp.Trace.Printf("successfully responded to hello world")
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