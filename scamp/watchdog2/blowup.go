package watchdog2

import (
  "encoding/json"
  "os"
  "fmt"
)

func panicjson(thing interface{}) () {
  panicBytes,_ := json.MarshalIndent(thing, "", "  ")
  fmt.Println(string(panicBytes))
  os.Exit(1)
}