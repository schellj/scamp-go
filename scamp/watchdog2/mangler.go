package watchdog2

import (
  "fmt"
)

func mangleFromExpInvFile(sector, actionSpec string) (output string) {
  return fmt.Sprintf("%s:%s", sector, actionSpec)
}

func mangleFromParts(sector, class, action string, version int) (output string) {
  return fmt.Sprintf("%s:%s.%s~%d", sector, class, action, version)
}