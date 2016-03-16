package main

import (
  "scamp"
  "time"

  "sync"
)

var timeout time.Duration = time.Duration(300) * time.Second

func main() {
  scamp.Initialize()

  var clients []*scamp.Client = make([]*scamp.Client, 10000)

  var wg sync.WaitGroup

  for i := range clients {
    client, err := scamp.Dial("127.0.0.1:30100")
    if err != nil {
      scamp.Error.Fatalf("could not connect! `%s`\n", err)
      return
    }

    clients[i] = client

    wg.Add(1)
    go hitService(client,wg)
  }

  scamp.Info.Printf("waiting for stress to finish")
  wg.Wait()
  scamp.Info.Printf("done with stress")
}


func hitService(client *scamp.Client, wg sync.WaitGroup) (){
  defer client.Close()
  defer func(){
    scamp.Info.Printf("wg.Done()!")
    wg.Done()
  }()
  defer scamp.Info.Printf("closing %s", client)

  message := scamp.NewMessage()
  message.SetRequestId(1234)
  message.SetAction("Logger.info")
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