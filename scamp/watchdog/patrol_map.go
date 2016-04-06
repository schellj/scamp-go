package watchdog

import (
  "fmt"

  "github.com/gudtech/scamp-go/scamp"
)

func mangleName(class scamp.ServiceProxyClass, actionDesc scamp.ActionDescription) (string) {
  return fmt.Sprintf("%s.%s~%d", class.Name(), actionDesc.Name(), actionDesc.Version())
}

type PatrolMap map[string]*ActionTracker

func (pm *PatrolMap) TrackActions(serviceCache *scamp.ServiceCache) {
  for _,remoteService := range serviceCache.All() {
    classes := remoteService.Classes()
    for _,class := range classes {
      // fmt.Printf("classes len: %d\n", len(classes))
      for _,actionDesc := range class.Actions() {
        name := mangleName(class, actionDesc)

        actionTracker := (*pm)[name]
        if actionTracker == nil {
          actionTracker = new(ActionTracker)
          (*pm)[name] = actionTracker
        }

        actionTracker.AdvertisedBy(remoteService.Ident())
      }
    }
  }

  pm.markEpoch()
}

func (pm PatrolMap) markEpoch() {
  for _,actionTracker := range pm {
    actionTracker.markEpoch()
  }
}