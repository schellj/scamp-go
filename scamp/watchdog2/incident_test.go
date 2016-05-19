package watchdog2

import (
  "testing"
)

func TestFormatting(t *testing.T) {
  var missing = map[string]InventoryDiffEntry{
    "main:Classy.thingy~3": InventoryDiffEntry {
      Missing: []string{
        "HostA", "HostB",
      },
    },
    "main:Classy.stuffy~2": InventoryDiffEntry {
      Missing: []string{
        "HostB",
      },
    },
  }

  var inc Incident = NewPagerdutyIncident("", missing)
  str,err := inc.Description()
  if err != nil {
    t.Fatalf(err.Error())
  }

  exp := `hey`
  if str != `hey` {
    t.Fatalf("expected `%s` got `%s`", exp, str)
  }

}