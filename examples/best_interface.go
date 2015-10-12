package main

import "fmt"
import "net"

func main(){
  infs,err := net.Interfaces()
  if err != nil {
    fmt.Printf("err: `%s`", err)
    return
  }

  var ip net.IP

  for _,inf := range infs {
    if ( inf.Flags & net.FlagLoopback != 0 ){
      continue
    }

    addrs,err := inf.Addrs()
    if err != nil {
      return
    }

    for _,addr := range addrs {
      ip,_,err = net.ParseCIDR(addr.String())
      if err != nil {
        fmt.Printf("ParseCIDR err: `%s`\n", err)
        continue
      } else if ip.To4() == nil {
        fmt.Printf("IP is not IPv4: `%s`\n", ip)
        continue
      }
      break
    }

    if ip != nil {
      break
    }
  }

  fmt.Printf("IP: %s\n", ip.String())
}