package scamp

import "net"

// We only want to broadcast on internal addresses so this
// filters down all available interfaces
func LocalMulticastAddrs() (addrs []net.Addr, err error) {
  interfaces,err := net.Interfaces()
  if err != nil {
    return
  }

  _,net192,err := net.ParseCIDR("192.168.0.0/16")
  if err != nil {
    return
  }
  _,net10,err := net.ParseCIDR("10.0.0.0/8")
  if err != nil {
    return
  }
  _,net172,err := net.ParseCIDR("172.16.0.0/12")
  if err != nil {
    return
  }

  addrs = make([]net.Addr,0)

  for _,inf := range interfaces {
    infAddrs,err := inf.Addrs()
    if err != nil {
      return nil, err
    }

    for _,addr := range infAddrs {
      // TODO: OSX provide all infAddrs as CIDR? Is that normal?
      ip,_,err := net.ParseCIDR(addr.String())
      if err != nil {
        Trace.Printf("could not parse %s", err)
        continue
      }

      if !(net192.Contains(ip) || net10.Contains(ip) || net172.Contains(ip)) {
        continue
      }

      addrs = append(addrs, addr)
    }

  }

  return
}

func BestLocalMulticastAddr() (addr net.Addr, err error) {
  addrs,err := LocalMulticastAddrs()
  if err != nil {
    return
  }

  addr = addrs[0]
  return
}