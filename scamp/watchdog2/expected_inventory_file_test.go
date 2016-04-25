package watchdog2

import (
  "testing"

  "strings"
)

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