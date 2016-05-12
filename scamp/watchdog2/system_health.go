package watchdog2

import (
  "github.com/gudtech/scamp-go/scamp"
)

type SystemHealth struct {
  Missing []string `json:"missing"`
  Red []string `json:"red"`
  Yellow []string `json:"yellow"`
}

func NewSystemHealth() (sh *SystemHealth) {
  sh = new(SystemHealth)
  sh.Missing = make([]string,0)
  sh.Red = make([]string,0)
  sh.Yellow = make([]string,0)

  return sh
}

func (sh *SystemHealth) IsDegraded() (bool) {
  if len(sh.Missing) > 0 {
    scamp.Error.Println("sh.Missing:", sh.Missing)
    return true
  } else if len(sh.Red) > 0 {
    scamp.Error.Println("sh.Red:", sh.Red)
    return true
  } else if len(sh.Yellow) > 0 {
    scamp.Error.Println("sh.Yellow:", sh.Yellow)
    return true
  }

  return false
}

func (sh *SystemHealth) MarkMissing(mangledName string) {
  sh.Missing = append(sh.Missing, mangledName)
}

func (sh *SystemHealth) MarkRed(mangledName string) {
  sh.Red = append(sh.Red, mangledName)
}

func (sh *SystemHealth) MarkYellow(mangledName string) {
  sh.Yellow = append(sh.Yellow, mangledName)
}