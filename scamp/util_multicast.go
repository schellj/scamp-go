package scamp

import "fmt"
import "net"

import "golang.org/x/net/ipv4"

func MulticastAddrForInterface(desiredInterfaceName string) (bestAddr net.Addr, err error) {
  interfaces,err := net.Interfaces()
  if err != nil {
    return
  }

  for _,inf := range interfaces {
    if inf.Name != desiredInterfaceName {
      continue
    }


    Trace.Printf("inf: %s", inf.Name)
    mulAddrs,_ := inf.MulticastAddrs()
    if len(mulAddrs) < 1 {
      err = fmt.Errorf("interface `%s` did not have a multicast interface", desiredInterfaceName)
      return
    }

    // find first IPv4 multicast address
    for _,mulAddr := range mulAddrs {
      Trace.Printf("looking at: `%s`", mulAddr.String())
      parsedIP := net.ParseIP(mulAddr.String())
      if parsedIP == nil {
        return nil, fmt.Errorf("could not parse IP: `%s`", mulAddr.String())
      }
      if parsedIP.To4() == nil {
        continue
      }
      bestAddr = mulAddr
      return bestAddr, nil
    }
  }

  err = fmt.Errorf("could no such interface `%s`", desiredInterfaceName)
  return
}

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
