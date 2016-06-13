package watchdog2

import (
  // "fmt"
  "testing"
  "encoding/json"

  "strings"
)

func TestSystemHealth(t *testing.T) {
  i,err := NewInventory("/Users/xavierlange/code/gudtech/workspace/src/github.com/gudtech/scamp-go/fixtures/watchdog/discovery.sample")
  if err != nil {
    t.Fatalf(err.Error())
  }

  i.Reload()

  content := strings.NewReader(`{
  "main": {
    "Inventory.Count.AddContainers~0": {
      "red": 0,
      "yellow": 3
    }
  }
}
`)

  ei := make(ExpectedInventory)
  err = json.NewDecoder(content).Decode(&ei)
  if err != nil {
    t.Fatalf(err.Error())
  }

  healthCheck := ei.GetSystemHealth(i)
  if !healthCheck.IsDegraded() {
    t.Fatalf("expected to be degraded")
  }
  if len(healthCheck.Yellow) != 1 {
    // t.Fatalf("expected yellow entries, got %s", healthCheck.Yellow)
  }
}

func TestSectorHealth(t *testing.T) {
  i,err := NewInventory("/Users/xavierlange/code/gudtech/workspace/src/github.com/gudtech/scamp-go/fixtures/watchdog/discovery.sample")
  if err != nil {
    t.Fatalf(err.Error())
  }

  i.Reload()

  content := strings.NewReader(`{
    "main:Inventory.Count.AddContainers~0": {
      "red": 0,
      "yellow": 3
    }
}
`)

  ei := make(ExpectedInventory)
  err = json.NewDecoder(content).Decode(&ei)
  if err != nil {
    t.Fatalf(err.Error())
  }

  sectorHealth := ei.GetSectorHealth(i)
  if !sectorHealth.IsDegraded() {
    t.Fatalf("sector health should have been degraded")
  }

  // panicjson(sectorHealth)
}

func TestEifFromReader(t *testing.T) {
  content := strings.NewReader(`{
    "main:Foo.bar~1": {
      "red": 2,
      "yellow": 5
    }
}
`)


  ei := make(ExpectedInventory)
  err := json.NewDecoder(content).Decode(&ei)
  if err != nil {
    t.Fatalf(err.Error())
  }

  if ei["main:Foo.bar~1"].Red != 2 {
    t.Fatalf("expected 2, got %d", ei["main:Foo.bar~1"].Red)
  } else if ei["main:Foo.bar~1"].Yellow != 5 {
    t.Fatalf("expected 5, got %d", ei["main:Foo.bar~1"].Yellow)
  }

  expectedName := "main:Foo.bar~1"
  entry,ok := ei[expectedName];
  if  !ok {
    t.Fatalf("expected mangled name `%s`", expectedName)
  } else if entry.Red != 2 {
    t.Fatalf("expected 2, got %d", ei["main:Foo.bar~1"].Red)
  } else if entry.Yellow != 5 {
    t.Fatalf("expected 5, got %d", ei["main:Foo.bar~1"].Yellow)
  }

}