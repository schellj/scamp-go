package watchdog2

import (
  "os"
  "fmt"
  "encoding/json"
  "strings"

  "github.com/gudtech/scamp-go/scamp"
)

type ExpectedInventory map[string]ExpectedInventoryEntry
type ExpectedInventoryEntry struct {
  Red int `json:"red"`
  Yellow int `json:"yellow"`
}


func NewExpectedInventory() (ei ExpectedInventory) {
  return make(ExpectedInventory)
}

func LoadExpectedInventoryFromFile(path string) (ei ExpectedInventory, err error) {
  file,err := os.Open(path)
  if err != nil {
    err = fmt.Errorf("error opening file `%s`: `%s`", path, err)
    return
  }

  ei = make(ExpectedInventory)
  err = json.NewDecoder(file).Decode(&ei)

  return
}

func (ei ExpectedInventory) GetSystemHealth(inv *Inventory) (sh *SystemHealth) {
  sh = NewSystemHealth()
  for mangledName,eiEntry := range ei {
    count,ok := inv.Get(mangledName)
    
    if !ok {
      sh.MarkMissing(mangledName)
    } else if count <= eiEntry.Red {
      scamp.Error.Println(mangledName,"has",count,"entries and triggers yellow at", eiEntry.Yellow)
      sh.MarkRed(mangledName)
    } else if count <= eiEntry.Yellow {
      scamp.Error.Println(mangledName,"has",count,"entries and triggers yellow at", eiEntry.Yellow)
      sh.MarkYellow(mangledName)
    } else {
      // all good
    }
  }

  return
}

func (ei ExpectedInventory) GetSectorHealth(inv *Inventory) (secH SectorHealth) {
  secH = NewSectorHealth()

  for mangledName,eiEntry := range ei {
    list,ok := inv.GetList(mangledName)
    if !ok {
      continue
    }

    sector,action,_ := parseMangledName(mangledName)
    // panic(fmt.Sprintf("sector: `%s`, action: `%s`, stuff: `%s`", sector, action, list))
    // panic(err.Error())

    missingInstances := make([]DegradedService, 0, len(list))
    for _,entry := range list {
      missingInstances = append(missingInstances, DegradedService(entry))
    }

    if len(list) <= eiEntry.Red {
      da := DegradedAction {
        ActionName: action,
        MinHealthyCount: eiEntry.Yellow+1,
        CurrentCount: len(list),
        HealthLabel: "red",
        MissingInstances: missingInstances,
      }

      _,ok := secH[sector]
      if !ok {
        secH[sector] = make([]DegradedAction,0)
      }

      secH[sector] = append(secH[sector], da)
    } else if len(list) <= eiEntry.Yellow {
      da := DegradedAction {
        ActionName: action,
        MinHealthyCount: eiEntry.Yellow+1,
        CurrentCount: len(list),
        HealthLabel: "yellow",
        MissingInstances: missingInstances,
      }

      _,ok := secH[sector]
      if !ok {
        secH[sector] = make([]DegradedAction,0)
      }

      secH[sector] = append(secH[sector], da)

      // panicjson(secH)
    } else {
      /* ALL GOOD */
    }

  }

  return
}

func parseMangledName(mangledName string) (sector,action string, err error) {
  sectorAndRest := strings.SplitN(mangledName,":",2)
  if len(sectorAndRest) != 2 {
    return "", "", fmt.Errorf("could not find : in `%s`", mangledName)
  }

  sector = sectorAndRest[0]
  action = sectorAndRest[1]

  return
}