package main

import "scamp"
import "net"
import "time"

func main() {
  scamp.Initialize()

  config := scamp.DefaultConfig()
  multicastSpec,err := config.BusSpec()

  udpAddr, err := net.ResolveUDPAddr("udp", multicastSpec)
  if err != nil {
    scamp.Trace.Printf("error resolving UDP address: `%s`", udpAddr)
  }

  multicastConn, err := net.DialUDP("udp", nil, udpAddr)
  if err != nil {
    scamp.Trace.Printf("could not dial multicast address: `%s`", err)
  }

  scamp.Trace.Printf("starting announce loop...")

  for {
    multicastConn.Write([]byte("hello, world\n"))
    time.Sleep(1 * time.Second)
  }
}