package scamp

import "bufio"

func (conn *Connection) ReadPackets(s *bufio.Scanner) (err error) {
  var didScan bool
  s.Split(packetBytesScanner)
  for {
    didScan = s.Scan()
    if didScan {
      Trace.Printf("Cool")
    } else {
      Trace.Printf("Not cool")
    }
  }

}

// Emits a packet as a token
func packetBytesScanner(data []byte, atEOF bool) (advance int, token []byte, err error) {
  Trace.Printf("scanning through bytes: `%s`", data)

  return 0, nil, nil
}