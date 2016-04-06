package scamp

import (
  "time"

  "fmt"

  "io"
  "os"

  "bytes"

  "crypto/tls"
  "crypto/rand"
  "encoding/base64"

  "sync/atomic"
)

// TODO: crazy debug mode enabled for now
var randomDebuggerString string
var enableWriteTee bool = true
var writeTeeTargetPath string = "/tmp/scamp_proto.bin"
type ScampDebugger struct {
  file *os.File
  wrappedWriter io.Writer
}


var scampDebuggerId uint64 = 0
func NewScampDebugger(conn *tls.Conn, clientType string) (handle *ScampDebugger, err error) {
  var worked bool = false
  var thisDebuggerId uint64 = 0
  for i:=0; i<10; i++{
    loadedVal := atomic.LoadUint64(&scampDebuggerId)
    thisDebuggerId = loadedVal + 1
    worked = atomic.CompareAndSwapUint64(&scampDebuggerId, loadedVal, thisDebuggerId)
    if worked {
      break
    }
  }
  if !worked {
    panic("never should happen...")
  }

  handle = new(ScampDebugger)

  var path = fmt.Sprintf("%s.%s.%s.%d", writeTeeTargetPath, randomDebuggerString, clientType, thisDebuggerId)

  handle.file,err = os.Create(path)
  if err != nil {
    return
  }

  return
}
func (handle *ScampDebugger)Write(p []byte) (n int, err error) {
  formattedStr := fmt.Sprintf("write: %d %s", time.Now().Unix(), p)
  _,err = handle.file.Write([]byte(formattedStr))
  if err != nil {
    return
  }

  return len(p), nil
}

func (handle *ScampDebugger)ReadWriter(p []byte) (n int, err error) {
  formattedStr := fmt.Sprintf("read: %d %s", time.Now().Unix(), p)
  _,err = handle.file.Write([]byte(formattedStr))
  if err != nil {
    return
  }

  return len(p), nil
}

func scampDebuggerRandomString() (string) {
  randBytes := make([]byte, 4, 4)
  _,err := rand.Read(randBytes)
  if err != nil {
    panic("shouldn't happen")
  }
  base64RandBytes := base64.StdEncoding.EncodeToString(randBytes)

  var buffer bytes.Buffer
  buffer.WriteString(base64RandBytes[0:])
  return string(buffer.Bytes())
}

type ScampDebuggerReader struct {
  wraps *ScampDebugger
}

func (sdr *ScampDebuggerReader)Write(p []byte) (n int, err error) {
  return sdr.wraps.ReadWriter(p)
}
