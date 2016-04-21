package main

import "github.com/gudtech/scamp-go/scamp"
import "net"
import "time"

var scampGroupName = "239.63.248.106"
var scampAnnounceDest = &net.UDPAddr{IP: net.IPv4(239,63,248,106), Port: 5555}

func main() {
  scamp.Initialize("/etc/SCAMP/soa.conf")
  var err error

  multicastPacketConn,err := scamp.LocalMulticastPacketConn()
  if err != nil {
    scamp.Error.Printf("could not create local multicast packet connection")
    return
  }

  for {
    if _, err := multicastPacketConn.WriteTo([]byte("hello, world\n"), nil, scampAnnounceDest); err != nil {
      scamp.Trace.Printf("failed to write to multicast group: `%s`", err)
      break
    }
    scamp.Trace.Printf("wrote hello world to group `%s`", scampAnnounceDest)
    time.Sleep(5 * time.Second)
  }
}