package watchdog2

import (
  "testing"

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

  eif,err := eifFromReader(content)
  if err != nil {
    t.Fatalf( err.Error() )
  }
  ei := eif.toExpectedInventory()

  healthCheck := ei.GetSystemHealth(i)
  if !healthCheck.IsDegraded() {
    t.Fatalf("expected to be degraded")
  }
  if len(healthCheck.Yellow) != 1 {
    t.Fatalf("expected yellow entries, got %s", healthCheck.Yellow)
  }
}

func TestEifFromReader(t *testing.T) {
  content := strings.NewReader(`{
  "main": {
    "Foo.bar~1": {
      "red": 2,
      "yellow": 5
    }
  }
}
`)

  eif,err := eifFromReader(content)
  if err != nil {
    t.Fatalf( err.Error() )
  }
  if eif["main"]["Foo.bar~1"].Red != 2 {
    t.Fatalf("expected 2, got %d", eif["main"]["Foo.bar~1"].Red)
  } else if eif["main"]["Foo.bar~1"].Yellow != 5 {
    t.Fatalf("expected 5, got %d", eif["main"]["Foo.bar~1"].Yellow)
  }

  ei := eif.toExpectedInventory()

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