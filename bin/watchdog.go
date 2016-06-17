package main

import (
  "os"
  "fmt"
  "flag"
  "time"

  "encoding/json"

  "github.com/gudtech/scamp-go/scamp"
  w "github.com/gudtech/scamp-go/scamp/watchdog"
)

func main() {
  var err error
  var ok bool

  gtConfigPathPtr := flag.String("config", "/backplane/discovery/discovery", "path to the discovery file")
  discoveryPathOverride := flag.String("discovery-path", "", "override configured discovery path (useful for testing)")
  inventoryPath := flag.String("expected-inventory", "", "config file specifying which actions and in what quantity are expected")
  scanWaitPtr := flag.Int("wait", 10, "how long to wait between action scans")
  pagerdutyKeyPtr := flag.String("pagerduty-key-path", "/etc/PagerDutyServiceKey", "location of the api key for pagerduty")
  modePtr := flag.String("mode", "watch", "one of 'watch' or 'dump-inventory'")
  flag.Parse()

  err = scamp.Initialize(*gtConfigPathPtr)
  if err != nil {
    fmt.Printf("could not initialize scamp: `%s`\n",err)
    return
  }

  if *inventoryPath == "" {
    fmt.Printf("must provide expected-inventory path")
    return
  }

  var discoveryPath string
  if *discoveryPathOverride != "" {
    discoveryPath = *discoveryPathOverride
  } else {
    discoveryPath,ok = scamp.DefaultConfig().Get("discovery.cache_path")
    if !ok {
      fmt.Printf("config is missing `discovery.cache_path`\n")
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

func doDump(inventoryPath, discoveryPath string) (err error) {
  serviceCache,err := scamp.NewServiceCache(discoveryPath)
  if err != nil {
    fmt.Println("could not create service cache:", err.Error())
    return
  }
  serviceCache.Scan()

  eif,err := w.ExpectedInventoryFileFromServiceCache(serviceCache)
  if err != nil {
    panic(err.Error())
  }

  inventoryFile,err := os.Create(inventoryPath)
  if err != nil {
    panic(err.Error())
  }


  indentedJson,err := json.MarshalIndent(eif, "", "  ")
  // json.NewEncoder(inventoryFile).Encode(eif)
  bytesWritten,err := inventoryFile.Write(indentedJson)
  if err != nil {
    panic(err.Error())
  } else if bytesWritten != len(indentedJson) {
    panic("did not write all data")
  }
  inventoryFile.Close()

  return
}

func doWatch(inventoryPath, discoveryPath string, scanWait int, pagerDutyKeyPath string) (err error) {
  tracker := w.NewWatchdogTracker()
  expectedInventory,err := w.LoadExpectedInventoryFromFile(inventoryPath)
  if err != nil {
    fmt.Printf("could not load expected inventory: `%s`\n", err)
    return
  }

  serviceCache,err := scamp.NewServiceCache(discoveryPath)
  if err != nil {
    fmt.Println("could not create service cache:", err.Error())
    return
  }

  /*
    The first pass on the file
  */
  err = doScanAndTrack(&tracker,serviceCache)
  if err != nil {
    fmt.Println("scan and compare failed:", err.Error())
    return
  }

  ScanLoop:
  for {
    select {
    case <-time.After(time.Duration(scanWait)*time.Second):
      err = doScanAndTrack(&tracker,serviceCache)
      if err != nil {
        fmt.Println("scan and compare failed:", err.Error())
        break ScanLoop
      }

      doDiff(&tracker, &expectedInventory, pagerDutyKeyPath)

      tracker.MarkEpoch()
    }
  }

  fmt.Printf("DONE!\n")

  return
}

func doScanAndTrack(tracker *w.WatchdogTracker, serviceCache *scamp.ServiceCache) (err error) {
  err = serviceCache.Scan()
  if err != nil {
    fmt.Printf("err loading cache: %s", err)
    return
  }

  var actionCount int = 0
  for _,service := range serviceCache.All() {
    for _,klass := range service.Classes() {
      actionCount += len(klass.Actions())
    }
  }  
  scamp.Error.Printf("individual action count: %d", actionCount)

  var trackedActionCount int = 0
  var trackedActionCountBefore int = 0
  for _,actionTracker := range *tracker {
    trackedActionCount += len(actionTracker.Idents())
    trackedActionCountBefore += len(actionTracker.IdentsBefore())
  }
  scamp.Error.Printf("tracked action count: %d", trackedActionCount)
  scamp.Error.Printf("tracked action counter (before): %d", trackedActionCountBefore)

  tracker.TrackActions(serviceCache)

  return
}

func doDiff(tracker *w.WatchdogTracker, expectedInventory *w.ExpectedInventory, pagerDutyKeyPath string) (err error) {
  // scamp.Error.Printf("HI")
  deficientActions := expectedInventory.Check(tracker)
  if len(deficientActions) != 0 {
    scamp.Error.Printf("detected deficient actions. firing pagerduty event.")
    // scamp.Error.Printf("%d deficient actions", deficientActions)
    err = w.TriggerEvent(pagerDutyKeyPath, deficientActions)
    if err != nil {
      return
    }
    // panic("aiee")
  } else {
    // scamp.Error.Printf("SUP")
  }

  return
}
