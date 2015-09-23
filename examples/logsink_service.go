package main

import "scamp"

func main() {
  scamp.Initialize()

  service,err := scamp.NewService(":30100", "staticdev")
  if err != nil {
    scamp.Error.Fatalf("error creating new service: `%s`", err)
  }

  service.Register("Logger.info", func(req scamp.Request, sess *scamp.Session){
    if len(req.Blob) > 0 {
      scamp.Info.Printf("Logging.info had data: %s", req.Blob)
    } else {
      scamp.Trace.Printf("Logging.info was called without data")
    }

    err = sess.Send(scamp.Reply{Blob: []byte("{}")})
    if err != nil {
      scamp.Error.Printf("error while sending reply: `%s`. continuing.", err)
      return
    }
    scamp.Trace.Printf("successfully responded to Logging.info")
  })

  service.Run()
}