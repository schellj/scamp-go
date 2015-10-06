package main

import "scamp"

func main() {
  scamp.Initialize()

  bestAddr,err := scamp.BestLocalMulticastAddr()
  if err != nil {
    scamp.Trace.Printf("could not find best addr: `%s`", err)
    return
  }

  scamp.Trace.Printf("found %s, the best usable multicast addrs", bestAddr)

  // addr := net.ResolveUDPAddr()
  // _, err := net.ListenMulticastUDP("udp", nil, nil)
  // if err != nil {
  //   scamp.Error.Fatalf("failed to start multicast listener: `%s`", err)
  // }


}