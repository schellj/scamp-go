package main

import "scamp"

func main() {
  scamp.Initialize()

  service,err := scamp.NewService(":30100", "staticdev")
  if err != nil {
    scamp.Error.Fatalf("error creating new service: `%s`", err)
  }

  // rename Register -> RegisterAction
  service.Register("Logger.info", func(req *scamp.Message, client *scamp.Client){
    reply := scamp.NewMessage()
    reply.SetMessageType(scamp.MESSAGE_TYPE_REPLY)
    reply.SetEnvelope(scamp.ENVELOPE_JSON)
    reply.SetRequestId(req.RequestId)
    reply.Write([]byte("{}"))

    _,err = client.Send(reply)
    if err != nil {
      scamp.Error.Printf("could not reply to message: `%s`", err)
      return
    }
  })

  service.Run()
}