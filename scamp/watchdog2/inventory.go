package watchdog2

import (
  "github.com/gudtech/scamp-go/scamp"
)

type Inventory struct {
  cache *scamp.ServiceCache
  inventory map[string]int
}

func NewInventory(path string) (i *Inventory, err error) {
  i.cache,err = scamp.NewServiceCache(path)
  if err != nil {
    return
  }
  i.inventory = make(map[string]int)
  return
}

func (i *Inventory)Reload() (err error) {
  err = i.cache.Scan()
  if err != nil {
    return
  }

  // Debugging goodness
  var actionCount int = 0
  for _,service := range i.cache.All() {
    for _,klass := range service.Classes() {
      actionCount += len(klass.Actions())
    }
  }  
  scamp.Error.Printf("individual action count: %d", actionCount)
  // END

  return
}
