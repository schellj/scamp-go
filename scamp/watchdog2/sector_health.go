package watchdog2

import (
  "fmt"
)

type SectorHealth map[string][]DegradedAction
func (sh SectorHealth)IsDegraded() (bool) {
  return len(sh) != 0
}

func (sh SectorHealth)Red() (red SectorHealth) {
  red = NewSectorHealth()

  for sectorName,degradedActions := range sh {
    for _,degradedAction := range degradedActions {
      if degradedAction.HealthLabel == "red" {
        _,ok := red[sectorName]
        if !ok {
          red[sectorName] = []DegradedAction{ degradedAction }
        } else {
          red[sectorName] = append(red[sectorName], degradedAction)
        }
      }
    }
  }

  return
}

func (sh SectorHealth)Yellow() (red SectorHealth) {
  red = NewSectorHealth()

  for sectorName,degradedActions := range sh {
    for _,degradedAction := range degradedActions {
      if degradedAction.HealthLabel == "yellow" {
        _,ok := red[sectorName]
        if !ok {
          red[sectorName] = []DegradedAction{ degradedAction }
        } else {
          red[sectorName] = append(red[sectorName], degradedAction)
        }
      }
    }
  }

  return
}

func NewSectorHealth() (SectorHealth) {
  return make(SectorHealth)
}

type DegradedAction struct {
  ActionName string // class:method~version
  MinHealthyCount int
  CurrentCount int
  HealthLabel string // red/yellow
  MissingInstances []DegradedService
}

type DegradedService ServiceDesc
// type ServiceDesc struct {
//   ShortHostname string `json:"short_hostname"`
//   ServiceName string `json:"service_name"`
// }

func (sh SectorHealth)ToSectorHealthByInventory() (shi SectorHealthByInventory) {
  shi = NewSectorHealthByInventory()

  for sector,degActs := range sh {
    for _,degAct := range degActs {
      for _,degServ := range degAct.MissingInstances {
        shi.Add(sector, degServ.ShortHostname, degServ.ServiceName, 0, degAct.CurrentCount, degAct.MinHealthyCount)
      }
    }
  }

  return
}

type SectorHealthByInventory map[string]SectorDigestMap
type SectorDigestMap map[string][]ServiceDesc

func NewSectorHealthByInventory() (shi SectorHealthByInventory) {
  shi = make(SectorHealthByInventory)
  return
}

func (shi SectorHealthByInventory)Add(sector, hostName, serviceName string, actionCount, actionInstances, actionTarget int) {
  if _,ok := shi[sector]; !ok {
    shi[sector] = make(SectorDigestMap)    
  }

  digestMap := shi[sector] 
  digest := fmt.Sprintf("%s has %d instances but needs %d. Missing:", serviceName, actionInstances, actionTarget)
  if _,ok := digestMap[digest]; !ok {
    digestMap[digest] = make([]ServiceDesc, 0)
  }

  digestMap[digest] = append(digestMap[digest], ServiceDesc {
    ShortHostname: hostName,
    ServiceName: serviceName,
  })

  return
}