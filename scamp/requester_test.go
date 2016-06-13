package scamp

import (
  "os"
  "flag"

  "testing"
)

func TestRequester(t *testing.T) {
  var err error

  msg := NewRequestMessage()
  msg.SetEnvelope(ENVELOPE_JSON)

  responseChan,err := MakeJsonRequest("main", "Logger.info", 1, msg)
  if err != nil {
    t.Fatalf(err.Error())
  }

  select {
  case resp := <-responseChan:
    panicjson(resp)
  default:
    panicjson("no way")
  }
}

func TestMain(m *testing.M) {
  flag.Parse()
  Initialize("/etc/SCAMP/soa.conf")
  os.Exit(m.Run())
}