package scamp

import "fmt"
import "strconv"
import "syscall"

type HighResTimestamp float64
func (ts HighResTimestamp) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf("%f", ts)), nil
}

func Gettimeofday() (ts HighResTimestamp, err error){
  var tval syscall.Timeval
  syscall.Gettimeofday(&tval)
  
  f,err := strconv.ParseFloat(fmt.Sprintf("%d.%d", tval.Sec, tval.Usec), 64)
  if err != nil {
    fmt.Printf("error creating timestamp: `%s`", err)
    return
  }
  
  ts = HighResTimestamp(f)
  return
}