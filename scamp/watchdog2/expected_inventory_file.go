package watchdog2

import (
  "fmt"
  "encoding/json"
  "os"
  "io"
)

// Two levels of indirection for sector -> actions.
// Bad idea?
type ExpectedInventoryFile map[string]ExpectedSectorInventory
type ExpectedSectorInventory map[string]ExpectedInventoryEntry

func (eif ExpectedInventoryFile) toExpectedInventory() (ei ExpectedInventory) {
  ei = NewExpectedInventory()

  for sector,inventory := range eif {
    for action,inventoryEntry := range inventory {
      mangledName := mangleFromExpInvFile(sector, action)
      ei[mangledName] = inventoryEntry
    }
  }

  return
}

func NewExpectedInventoryFile() (eif ExpectedInventoryFile) {
  eif = make(map[string]ExpectedSectorInventory)
  return
}

func LoadExpectedInventoryFromFile(path string) (ei ExpectedInventory, err error) {
  file,err := os.Open(path)
  if err != nil {
    err = fmt.Errorf("erroring opening file `%s`: `%s`", path, err)
    return
  }

  eif,err := eifFromReader(file)
  if err != nil {
    return
  }

  ei = eif.toExpectedInventory()

  return
}

func eifFromReader(reader io.Reader) (eif ExpectedInventoryFile, err error) {
  eif = make(ExpectedInventoryFile)
  err = json.NewDecoder(reader).Decode(&eif)
  return
}