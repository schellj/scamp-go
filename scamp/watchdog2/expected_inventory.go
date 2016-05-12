package watchdog2

import (
  "os"
  "fmt"
  "encoding/json"

  "github.com/gudtech/scamp-go/scamp"
)

type ExpectedInventory map[string]ExpectedInventoryEntry
type ExpectedInventoryEntry struct {
  Red int `json:"red"`
  Yellow int `json:"yellow"`
}


func NewExpectedInventory() (ei ExpectedInventory) {
  return make(ExpectedInventory)
}

func LoadExpectedInventoryFromFile(path string) (ei ExpectedInventory, err error) {
  file,err := os.Open(path)
  if err != nil {
    err = fmt.Errorf("error opening file `%s`: `%s`", path, err)
    return
  }

  ei = make(ExpectedInventory)
  err = json.NewDecoder(file).Decode(&ei)

  return
}

func (ei ExpectedInventory) GetSystemHealth(inv *Inventory) (sh *SystemHealth) {
  sh = NewSystemHealth()
  for mangledName,eiEntry := range ei {
    count,ok := inv.Get(mangledName)
    
    if !ok {
      sh.MarkMissing(mangledName)
    } else if count <= eiEntry.Red {
      scamp.Error.Println(mangledName,"has",count,"entries and triggers yellow at", eiEntry.Yellow)
      sh.MarkRed(mangledName)
    } else if count <= eiEntry.Yellow {
      scamp.Error.Println(mangledName,"has",count,"entries and triggers yellow at", eiEntry.Yellow)
      sh.MarkYellow(mangledName)
    } else {
      // all good
    }
  }

  return
}