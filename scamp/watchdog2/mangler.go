package watchdog2

import (
  "fmt"
)

func mangleFromExpInvFile(sector, actionSpec string) (output string) {
  return fmt.Sprintf("%s:%s", sector, actionSpec)
}