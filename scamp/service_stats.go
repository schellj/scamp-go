package scamp

import (
  "time"
  "encoding/json"
)

type ServiceStats struct {
  ClientsAccepted uint64 `json:"total_clients_accepted"`
  OpenConnections uint64 `json:"open_connections"`
}

func GatherStats(service *Service) (stats ServiceStats) {
  stats.ClientsAccepted = service.connectionsAccepted
  stats.OpenConnections = uint64(len(service.clients))

  return
}

func PrintStatsLoop(service *Service, timeout time.Duration, closeChan chan bool) {
  forLoop:
  for {
    select {
    case <- time.After(timeout):
      stats := GatherStats(service)
      statsBytes,err := json.Marshal(&stats)
      if err != nil {
        continue
      }

      Info.Printf("periodic stats: `%s`", statsBytes)
    case <-closeChan:
      break forLoop
    }
  }

  Info.Printf("exiting PrintStatsLoop")
}