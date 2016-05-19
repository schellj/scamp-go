package watchdog2

import (
  "fmt"
  "bytes"
  "strings"
  "text/template"

  "github.com/gudtech/scamp-go/scamp"
)

type Incident interface {
  Description() (string, error)
  // Notify() (error)
  Resolve() (error)
}

type PagerdutyIncident struct {
  Missing map[string]InventoryDiffEntry
}

func NewPagerdutyIncident(pagerdutyKey string, missing map[string]InventoryDiffEntry) (yi *PagerdutyIncident) {
  return &PagerdutyIncident {
    Missing: missing,
  }
}

var piTmpl = template.Must(
  template.New("desc").Parse(
`System is in a degraded state ({{.Level}}):
{{range $index, $element := .DegradedActions}}Action "{{$index}}" has lost:
{{range $element.Missing}} - {{.}}
{{end}}
{{end}}
`,
  ),
)
type piTmplArgs struct {
  DegradedActions map[string]InventoryDiffEntry
  Level string
}

/*
  Next gen formatting
*/
type pdTmplArgs struct {
  sectors map[string]pdTmplArgsSector
}

type pdTmplArgsSector struct {
  digests map[string][]pdTmplServiceDigest
}

type pdTmplServiceDigest struct {
  serverHost string
  serviceName string
  missingActions []string
}

func (pi *PagerdutyIncident)MissingToPdTmplArgs() (args pdTmplArgs) {
  args.sectors = make(map[string]pdTmplArgsSector)

  for k,missingIdents := range pi.Missing {
    // main:Classy.thingy~3
    parts := strings.SplitN(k,":",2)
    if len(parts) != 2 {
      scamp.Error.Printf("could not parse %s", k)
      continue
    }

    sector := parts[0]

    tmplArgsSector,ok := args.sectors[sector]
    if !ok {
      tmplArgsSector = pdTmplArgsSector{
        digests: make(map[string][]pdTmplServiceDigest),
      }
      args.sectors[sector] = tmplArgsSector
    }

    rest := parts[1]
    restParts := strings.SplitN(rest,"~",2)

    digestStr := fmt.Sprintf("missing %d instances", len(missingIdents.Missing))

    // var action,version string
    if len(restParts) != 2 {
      scamp.Error.Printf("could not parse rest: %s", rest)
      continue
    }

    // action := restParts[0]
    // version := restParts[1]

    tmpDigest := pdTmplServiceDigest {
      serverHost: "hostasdf",
      serviceName: "servasdf",
      missingActions: []string{},
    }

    _,ok = tmplArgsSector.digests[digestStr]
    if !ok {
      tmplArgsSector.digests[digestStr] = []pdTmplServiceDigest{ tmpDigest }
    } else {
      tmplArgsSector.digests[digestStr] = append(tmplArgsSector.digests[digestStr], tmpDigest)
    }


    }

    return

  }


func (pi *PagerdutyIncident)Description() (desc string, err error) {
  pi.MissingToPdTmplArgs()

  var descBuf bytes.Buffer
  tmplArgs := piTmplArgs {
    DegradedActions: pi.Missing,
    Level: "red",
  }
  err = piTmpl.Execute(&descBuf, tmplArgs)
  if err != nil {
    return "", err
  }

  desc = string(descBuf.Bytes())
  if len(desc) > 1024 {
    desc = desc[0:1023]
  }

  return
}

func (pi *PagerdutyIncident)Resolve() (error) {
  panic("SUP")
}


type SlackIncident struct {
  expInv *ExpectedInventory
  inv *Inventory
}

func NewSlackIncident(expInv *ExpectedInventory, inv *Inventory) (yi *SlackIncident) {
  return &SlackIncident {
    expInv: expInv,
    inv: inv,
  }
}