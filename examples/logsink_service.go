package main

import (
  "scamp"

  "os"
  "os/signal"
  "syscall"
  "time"
)

func main() {
  scamp.Initialize("/etc/SCAMP/soa.conf")

  service,err := scamp.NewService(":30100", "staticdev")
  if err != nil {
    scamp.Error.Fatalf("error creating new service: `%s`", err)
  }

  // rename Register -> RegisterAction
  service.Register("Logger.log", func(req *scamp.Message, client *scamp.Client){
    scamp.Info.Printf("received msg len: %d", len(req.Bytes()))

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

  serviceDone := make(chan bool)
  go func(){
    service.Run()
    serviceDone <- true
  }()

  sigUsr1 := make(chan os.Signal)
  signal.Notify(sigUsr1, syscall.SIGUSR1)
  select {
  case <-sigUsr1:
    scamp.Info.Printf("shutdown requested")
    service.Stop()
  }

  select {
  case <-serviceDone:
    scamp.Info.Printf("service exiting gracefully")
  }

  scamp.Info.Printf("going to timeout so you can send a SIGQUIT and dump the goroutines...")
  select {
  case <- time.After(time.Duration(1) * time.Minute):
    scamp.Info.Printf("1 minute timeout achieved")
  }
}