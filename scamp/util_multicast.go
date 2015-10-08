package scamp

import "fmt"
import "net"

import "golang.org/x/net/ipv4"

func LoopbackInterface() (lo *net.Interface, err error) {
  lo,err = net.InterfaceByName("lo0")
  if err != nil {
    lo,err = net.InterfaceByName("lo")
    if err != nil {
      Error.Printf("could not find `lo0` or `lo`: `%s`", err)
      return
    }
  }

  return
}

func LocalMulticastPacketConn() (conn *ipv4.PacketConn, err error) {
  lo,err := LoopbackInterface()
  if err != nil {
    return
  }

  maddrs, err := lo.MulticastAddrs()
  if err != nil {
    return
  }

  var bestAddr net.Addr
  for _,maddr := range maddrs {
    Trace.Printf("looking at: `%s`", maddr.String())
    parsedIP := net.ParseIP(maddr.String())
    if parsedIP == nil {
      Error.Printf("could not parsed IP: `%s`", maddr.String())
      continue
    } else if parsedIP.To4() == nil {
      continue
    }
    bestAddr = maddr
    break
  }
  if bestAddr == nil {
    err = fmt.Errorf("could not find a good address to bind to")
    return
  }

  localMulticastSpec := fmt.Sprintf("%s:%d", bestAddr, 5555)
  Trace.Printf("announce binding to port: `%s`", localMulticastSpec)
  udpConn, err := net.ListenPacket("udp", localMulticastSpec)
  if err != nil {
    Error.Printf("could not listen to `%s`", localMulticastSpec)
    return
  }

  conn = ipv4.NewPacketConn(udpConn)
  return
}
