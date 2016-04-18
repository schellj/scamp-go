package watchdog

import (
  "fmt"

  "github.com/gudtech/scamp-go/scamp"
)

func mangleName(proxy scamp.ServiceProxy, class scamp.ServiceProxyClass, actionDesc scamp.ActionDescription) (string) {
  return fmt.Sprintf("%s:%s.%s~%d", proxy.Sector(), class.Name(), actionDesc.Name(), actionDesc.Version())
}

func mangledNameFromParts(sector, actionShorthand string) (string) {
  return fmt.Sprintf("%s:%s", sector, actionShorthand)
}

func mangleNameWithSectorString(sector string, class scamp.ServiceProxyClass, actionDesc scamp.ActionDescription) (string) {
  return fmt.Sprintf("%s:%s.%s~%d", sector, class.Name(), actionDesc.Name(), actionDesc.Version())
} 