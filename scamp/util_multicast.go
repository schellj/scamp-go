package scamp

import "fmt"
import "net"


// var localInterface = []byte("lo0")

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

// // We only want to broadcast on internal addresses so this
// // filters down all available interfaces
// func LocalMulticastAddrs() (interfaceName string, addrs []net.Addr, err error) {
//   interfaces,err := net.Interfaces()
//   if err != nil {
//     return
//   }

//   // _,net192,err := net.ParseCIDR("192.168.0.0/16")
//   // if err != nil {
//   //   return
//   // }
//   // _,net10,err := net.ParseCIDR("10.0.0.0/8")
//   // if err != nil {
//   //   return
//   // }
//   // _,net172,err := net.ParseCIDR("172.16.0.0/12")
//   // if err != nil {
//   //   return
//   // }

//   addrs = make([]net.Addr,0)

//   for _,inf := range interfaces {
//     Trace.Printf("inf: %s", inf.Name)
//     mulAddrs,_ := inf.MulticastAddrs()
//     for _,maddr := range mulAddrs {
//       Trace.Printf("maddr: %s", maddr)
//       ip,_,err := net.ParseCIDR(maddr.String())
//       if err != nil {
//         // Trace.Printf("could not parse %s", err)
//         continue
//       }

//       Trace.Printf("maddr IsInterfaceLocalMulticast: %s", ip.IsInterfaceLocalMulticast())

//     }

//     // infAddrs,err := inf.Addrs()
//     // if err != nil {
//     //   return nil, err
//     // }

//     // for _,addr := range infAddrs {
//     //   // TODO: OSX provide all infAddrs as CIDR? Is that normal?
//     //   ip,_,err := net.ParseCIDR(addr.String())
//     //   if err != nil {
//     //     Trace.Printf("could not parse %s", err)
//     //     continue
//     //   }

//     //   Trace.Printf("addr IsInterfaceLocalMulticast: %s", ip.IsInterfaceLocalMulticast())

//     //   if !(net192.Contains(ip) || net10.Contains(ip) || net172.Contains(ip)) {
//     //     continue
//     //   }

//     //   addrs = append(addrs, addr)
//     // }

//   }

//   return []net.Addr{}, nil
// }

// func BestLocalMulticastAddr() (addr net.Addr, err error) {
//   addrs,err := LocalMulticastAddrs()
//   if err != nil {
//     return
//   }

//   addr = addrs[0]
//   return
// }