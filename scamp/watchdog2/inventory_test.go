package watchdog2

import (
  "flag"
  "os"

  "github.com/gudtech/scamp-go/scamp"

  "testing"
)

func TestInventory(t *testing.T) {
  i,err := NewInventory("/Users/xavierlange/code/gudtech/workspace/src/github.com/gudtech/scamp-go/fixtures/watchdog/discovery.sample")
  if err != nil {
    t.Fatalf(err.Error())
  }

  i.Reload()
}

func TestMain(m *testing.M) {
  flag.Parse()
  scamp.Initialize("/etc/SCAMP/soa.conf")
  os.Exit(m.Run())
}