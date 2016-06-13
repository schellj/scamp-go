package watchdog2

import (
  "github.com/gudtech/scamp-go/scamp"
)

/*
  TODO: the ease of retrieving a mangled name should be pushed in
  to the scamp.ServiceCache and this file should go away
*/

type ServiceDesc struct {
  ShortHostname string `json:"short_hostname"`
  ServiceName string `json:"service_name"`
}

type InventoryType map[string][]ServiceDesc

type Inventory struct {
  Cache *scamp.ServiceCache
  Inventory InventoryType
}

func NewInventory(cachePath string) (i *Inventory, err error) {
  i = new(Inventory)
  i.Cache,err = scamp.NewServiceCache(cachePath)
  if err != nil {
    return
  }
  i.Inventory = make(InventoryType)
  return
}

func (i *Inventory)Clone() (i2 *Inventory) {
  invCopy := make(InventoryType)
  for k,v := range i.Inventory {
    invCopy[k] = v
  }
  return &Inventory {
    Cache: nil,
    Inventory: invCopy,
  }
}

type InventoryDiffEntry struct {
  Missing []ServiceDesc `json:"missing"`
}

func (old *Inventory)Diff(inv *Inventory) (diff map[string]InventoryDiffEntry) {
  diff = make(map[string]InventoryDiffEntry)

  for oldK,oldList := range old.Inventory {
    list,ok := inv.Inventory[oldK]
    if !ok {
      diff[oldK] = InventoryDiffEntry {
        Missing: list,
      }
    } else if len(oldList) <= len(list) {
      // if we restart a service it will probably generate a
      // new service name so we shouldn't count it as "missing"
      // but rather as "replaced". for now we skip it.
      continue
    } else {
      missing := make([]ServiceDesc,0)

      for _,oldEntry := range oldList {
        found := false
        for _,entry := range list {
          found = oldEntry == entry
          if found {
            break
          }
        }
        if !found {
          missing = append(missing, oldEntry)
        }
      }

      if len(missing) == 0 {
        continue
      }

      diff[oldK] = InventoryDiffEntry {
        Missing: missing,
      }
    }
  }

  return
}

func (i *Inventory)Reload() (err error) {
  err = i.Cache.Refresh()
  if err != nil {
    return
  }

  // Debugging goodness
  var actionCount int = 0
  for _,service := range i.Cache.All() {
    for _,klass := range service.Classes() {
      actionCount += len(klass.Actions())
    }
  }  
  scamp.Error.Printf("individual action count: %d", actionCount)
  // END

  i.Inventory = make(InventoryType)
  for _,service := range i.Cache.All() {
    for _,klass := range service.Classes() {
      for _,action := range klass.Actions() {
        mangledName := mangleFromParts(service.Sector(), klass.Name(), action.Name(), action.Version())
        _,ok := i.Inventory[mangledName]
        serviceDesc := ServiceDesc {
          ShortHostname: service.ShortHostname(),
          ServiceName: service.BaseIdent(),
        }
        if !ok {
          i.Inventory[mangledName] = []ServiceDesc{ serviceDesc }
        } else {
          i.Inventory[mangledName] = append(i.Inventory[mangledName], serviceDesc )
        }
      }
    }
  }

  return
}

func (i *Inventory) Get(mangledName string) (val int, ok bool) {
  list,ok := i.Inventory[mangledName]
  val = len(list)
  return
}

func (i *Inventory) GetList(mangledName string) (list []ServiceDesc, ok bool) {
  list,ok = i.Inventory[mangledName]
  return
}