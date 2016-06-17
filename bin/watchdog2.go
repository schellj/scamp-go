package main

import (
  "os"
  "fmt"
  "flag"
  "time"

  "encoding/json"

  "github.com/gudtech/scamp-go/scamp"
  w "github.com/gudtech/scamp-go/scamp/watchdog2"
)

func main() {
  var err error
  var ok bool

  gtConfigPathPtr := flag.String("config", "/backplane/etc/soa.conf", "path to the discovery file")
  discoveryPathOverride := flag.String("discovery-path", "", "override configured discovery path (useful for testing)")
  inventoryPath := flag.String("expected-inventory", "", "config file specifying which actions and in what quantity are expected")
  scanWaitPtr := flag.Int("wait", 10, "how long to wait between action scans")
  pagerdutyKeyPtr := flag.String("pagerduty-key-path", "/etc/PagerDutyServiceKey", "location of the api key for pagerduty")
  modePtr := flag.String("mode", "watch", "one of 'watch' or 'dump-inventory'")
  flag.Parse()

  err = scamp.Initialize(*gtConfigPathPtr)
  if err != nil {
    fmt.Println("could not initialize scamp: `",err,"`")
    return
  }

  if *inventoryPath == "" {
    fmt.Println("must provide expected-inventory path")
    return
  }

  var discoveryPath string
  if *discoveryPathOverride != "" {
    discoveryPath = *discoveryPathOverride
  } else {
    discoveryPath,ok = scamp.DefaultConfig().Get("discovery.cache_path")
    if !ok {
      fmt.Println("config is missing `discovery.cache_path`\n")
      return
    }
  }

  if *modePtr == "watch" {
    err = doWatch(*inventoryPath, discoveryPath, *scanWaitPtr, *pagerdutyKeyPtr)
  } else if *modePtr == "dump-inventory" {
    err = doDump(*inventoryPath, discoveryPath)
    if err != nil {
      fmt.Println("failed to dump inventory:", err.Error())
      os.Exit(1)
    }

    fmt.Println("data successfully written")

  } else {
    fmt.Println("unsupported mode ", *modePtr)
    os.Exit(1)
  }

  if err != nil {
    scamp.Error.Printf("failed: `%s`", err)
  }
}

func doDump(expInvPath, discoveryPath string) (err error) {
  inv,err := w.NewInventory(discoveryPath)
  if err != nil {
    return
  }

  err = inv.Reload()
  if err != nil {
    return
  }

  expInv := w.NewExpectedInventory()
  for mangledName,list := range inv.Inventory {
    expInv[mangledName] = w.ExpectedInventoryEntry {
      Red: len(list)-2,
      Yellow: len(list)-1,
    }
  }

  expInvBytes,err := json.MarshalIndent(expInv, "", "  ")
  if err != nil {
    return
  }

  f,err := os.Create(expInvPath)
  if err != nil {
    return
  }
  _,err = f.Write(expInvBytes)
  if err != nil {
    return
  }

  return
}

type PagerdutyConfig struct {
  UrgentKey string `json:"urgent"`
  NotificationKey string `json:"notification"`
}
var pagerdutyConfig PagerdutyConfig

// var descTmpl = template.Must(
//   template.New("desc").Parse(
// `System is in a degraded state ({{.Level}}):
// {{range $index, $element := .DegradedActions}}Action "{{$index}}" has lost:
// {{range $element.Missing}} - {{.}}
// {{end}}
// {{end}}
// `,
//   ),
// )
// type descTmplArgs struct {
//   DegradedActions map[string]w.InventoryDiffEntry
//   Level string
// }

func doWatch(expInvPath, discoveryPath string, scanWait int, pagerDutyKeyPath string) (err error) {
  pdf,err := os.Open(pagerDutyKeyPath)
  if err != nil {
    return
  }
  json.NewDecoder(pdf).Decode(&pagerdutyConfig)

  expInv,err := w.LoadExpectedInventoryFromFile(expInvPath)
  if err != nil {
    return
  }

  inv,err := w.NewInventory(discoveryPath)
  if err != nil {
    return
  }

  var lastGoodInv *w.Inventory
  var urgentIncidentId string
  var notifIncidentId string

  var urgentIncident w.Incident = nil
  var notifyIncident w.Incident = nil

  for {
    time.Sleep(time.Duration(scanWait)*time.Second)

    err = inv.Reload()
    if err != nil {
      return err
    }

    health := expInv.GetSystemHealth(inv)

    // New style of code, driven by interfaces
    if len(health.Red) == 0 && urgentIncident != nil {
      urgentIncident.Resolve()
    }
    if len(health.Yellow) == 0 && notifyIncident != nil {
      notifyIncident.Resolve()
    }

    var actionsToMissingIdents map[string]w.InventoryDiffEntry
    if health.IsDegraded() {
      scamp.Error.Printf("system is in a degraded state")
      if lastGoodInv == nil {
        scamp.Error.Printf("we've never seen a healthy state. watchdog was restarted during outage?")

        continue
      }
      actionsToMissingIdents = lastGoodInv.Diff(inv)
    } else {
      lastGoodInv = inv.Clone()
      continue
    }

    if len(health.Red) > 0 /* && urgentIncidentId == "" */ {
      red := make(map[string]w.InventoryDiffEntry)
      for _,redAction := range health.Red {
        red[redAction] = actionsToMissingIdents[redAction]
      }

      urgentIncident = w.NewPagerdutyIncident(pagerdutyConfig.UrgentKey, red)
    }

    if len(health.Yellow) > 0 /* && notifIncidentId == "" */ {
    }

    if urgentIncidentId != "" {
      scamp.Error.Printf("we have an open urgent incident")
    }
    if notifIncidentId != "" {
      scamp.Error.Printf("we have an open notification incident")
    }

  }

  return
}
