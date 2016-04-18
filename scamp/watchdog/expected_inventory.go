package watchdog

import (
  "fmt"
  "os"
  "io"

  "encoding/json"
)

// Two levels of indirection for sector -> actions.
// Bad idea?
type ExpectedInventoryFile struct {
  Sectors map[string]ExpectedSectorInventory `json:"sectors"`
}
type ExpectedSectorInventory map[string]int

type ExpectedInventory map[string]int

func LoadExpectedInventoryFromFile(path string) (ei ExpectedInventory, err error) {
  file,err := os.Open(path)
  if err != nil {
    err = fmt.Errorf("erroring opening file `%s`: `%s`", path, err)
    return
  }

  inventoryFromFile,err := decodeInventoryFileFromReader(file)
  if err != nil {
    return
  }

  ei = expectedInventoryFileToExpectedInventory(inventoryFromFile)

  return
}

func expectedInventoryFileToExpectedInventory(eif ExpectedInventoryFile) (ei ExpectedInventory) {
  ei = make(ExpectedInventory)

  for sector,inventory := range eif.Sectors {
    for action,expectedCount := range inventory {
      mangledName := mangledNameFromParts(sector, action)
      ei[mangledName] = expectedCount
    }
  }

  return
}

func decodeInventoryFileFromReader(reader io.Reader) (eif ExpectedInventoryFile, err error) {
  err = json.NewDecoder(reader).Decode(&eif)
  return
}

type DeficientActionDescription struct {
  CountBefore int `json:"count_before"`
  CountAfter int `json:"count_after"`
  ExpectedCount int `json:"expected_count"`
  IdentsMissing []string `json:"idents_missing"`
  ActionName string `json:"action_name"`
}
type DeficientActionsReport []DeficientActionDescription

func (ei ExpectedInventory) Check(sit *ServiceInventoryTracker) (dar DeficientActionsReport) {
  dar = make(DeficientActionsReport, 0)
  for actionName,actionTracker := range *sit {
    if expectedCount,ok := ei[actionName]; ok {
      if actionTracker.InstanceCount() != expectedCount {
        desc := DeficientActionDescription {
          CountBefore: len(actionTracker.IdentsBefore()),
          CountAfter: len(actionTracker.Idents()),
          ExpectedCount: expectedCount,
          IdentsMissing: actionTracker.MissingIdentsThisEpoch(),
          ActionName: actionName,
        }
        dar = append(dar, desc)
      }
    }
  }

  return
}