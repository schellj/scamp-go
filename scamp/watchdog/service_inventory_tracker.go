package watchdog

import (
  // "fmt"

  "github.com/gudtech/scamp-go/scamp"
)

type ServiceInventoryTracker map[string]*ActionTracker

func NewServiceInventoryTracker() (ServiceInventoryTracker) {
  return make(ServiceInventoryTracker)
}

func (sit *ServiceInventoryTracker) TrackActions(serviceCache *scamp.ServiceCache) {
  for _,remoteService := range serviceCache.All() {
    classes := remoteService.Classes()
    for _,class := range classes {
      // fmt.Printf("classes len: %d\n", len(classes))
      for _,actionDesc := range class.Actions() {
        name := mangleNameWithSectorString(remoteService.Sector(), class, actionDesc)

        actionTracker := (*sit)[name]
        if actionTracker == nil {
          actionTracker = new(ActionTracker)
          (*sit)[name] = actionTracker
        }

        actionTracker.AdvertisedBy(remoteService.Ident())
      }
    }
  }

  sit.markEpoch()
}

func (sit ServiceInventoryTracker) markEpoch() {
  for _,actionTracker := range sit {
    actionTracker.markEpoch()
  }
}