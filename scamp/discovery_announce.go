package scamp

import "time"
import "net"

import "golang.org/x/net/ipv4"

type DiscoveryAnnouncer struct {
  services []*Service
  multicastConn *ipv4.PacketConn
  multicastDest *net.UDPAddr
  stopSig (chan bool)
}

func NewDiscoveryAnnouncer() (announcer *DiscoveryAnnouncer, err error) {
  announcer = new(DiscoveryAnnouncer)
  announcer.services = make([]*Service, 0, 0)
  announcer.stopSig = make(chan bool)

  config := DefaultConfig()
  announcer.multicastDest = &net.UDPAddr{IP: config.DiscoveryMulticastIP(), Port: config.DiscoveryMulticastPort()}
  // announcer.multicastDest = &net.UDPAddr{IP: 127.0.0.1, Port: config.DiscoveryMulticastPort()}
  announcer.multicastConn,err = LocalMulticastPacketConn()
  if err != nil {
    return
  }

  return
}

func (announcer *DiscoveryAnnouncer)Stop(){
  announcer.stopSig <- true
}

func (announcer *DiscoveryAnnouncer)Track(serv *Service){
  announcer.services = append(announcer.services, serv)
}

func (announcer *DiscoveryAnnouncer)AnnounceLoop() {
  Trace.Printf("starting announcer loop")

  for {
    select {
    case <- announcer.stopSig:
      return
    default:
      announcer.doAnnounce()
    }

    time.Sleep(time.Duration(defaultAnnounceInterval) * time.Second)
  }
}

func (announcer *DiscoveryAnnouncer)doAnnounce() (err error){
  for _,serv := range announcer.services {

    Info.Printf("\n\nserv: %+v\n\n", serv)

    serviceDesc,err := serv.MarshalText()
    if err != nil {
      Error.Printf("failed to marshal service as text: `%s`. skipping.", err)
    }
    Info.Printf("\nserviceDesc: %s\n", serviceDesc)
    // Info.Printf("multicastDest: %s", announcer.multicastDest)
    _,err = announcer.multicastConn.WriteTo(serviceDesc, nil, announcer.multicastDest)
    if err != nil {
      return err
    }
  }

  return
}


// Loop for broadcasting service in ServiceProxy format
// AnnounceService(*Service) {

// }
