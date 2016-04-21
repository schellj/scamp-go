package watchdog

import (
  "testing"

  "strings"
)

func TestExpectedInventory(t *testing.T) {
  var err error

  inventoryDesc := strings.NewReader(`{
  "sectors": {
    "main": {
      "asdf1234~1": 2
    }
  }
}`)
  
  ei,err := decodeInventoryFileFromReader(inventoryDesc)
  if err != nil {
    t.Fatalf("failed: `%s`", err)
  }

  if len(ei.Sectors) != 1 {
    t.Fatalf("should have 1 sector")
  } else if len(ei.Sectors["main"]) != 1 {
    t.Fatalf("main should have 1 action")
  } else if ei.Sectors["main"]["asdf1234~1"] != 2 {
    t.Fatalf("should have `asdf1234~1` in the inventory")
  }
}

func TestCheckingInventory(t *testing.T) {
  at := NewActionTracker()
  at.AdvertisedBy("xavier")
  at.AdvertisedBy("daniel")

  wt := NewWatchdogTracker()
  wt["asdf1234~1"] = at

  inventoryDesc := strings.NewReader(`{
  "sectors": {
    "main": {
      "asdf1234~1": 2
    }
  }
}`)

  eif,err := decodeInventoryFileFromReader(inventoryDesc)
  if err != nil {
    t.Fatalf("failed: `%s`", err)
  }
  ei := expectedInventoryFileToExpectedInventory(eif)

  deficientActions := ei.Check(&wt)
  if len(deficientActions) != 0 {
    t.Fatalf("the inventory check should have succeeded: `%s`", deficientActions)
  }
}