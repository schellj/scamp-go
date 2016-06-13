package watchdog2

import (
  // "fmt"
  "bytes"
  // "strings"
  "text/template"

  // "github.com/gudtech/scamp-go/scamp"
)

type Incident interface {
  Description() (string, error)
  // Notify() (error)
  Resolve() (error)
}

type PagerdutyIncident struct {
  Missing SectorHealth
}

func NewPagerdutyIncident(pagerdutyKey string, missing SectorHealth) (yi *PagerdutyIncident) {
  return &PagerdutyIncident {
    Missing: missing,
  }
}

// var piTmpl = template.Must(
//   template.New("desc").Parse(
// `System is in a degraded state ({{.Level}}):
// {{range $index, $element := .DegradedActions}}Action "{{$index}}" has lost:
// {{range $element.Missing}} - {{.}}
// {{end}}
// {{end}}
// `,
//   ),
// )

/*
type SectorHealthByInventory map[string][]InventoryHealth
type InventoryHealth struct {
  Digest string
  Services []ServiceDesc
}
*/
var piTmpl = template.Must(
  template.New("desc").Parse(`{{ range $sector,$digMap := . }}{{$sector}}
{{ range $digest,$degSvcs := $digMap }}{{$digest}}
{{ range $i2,$degSvc := $degSvcs }}  {{$degSvc.ShortHostname}}
{{end}}{{end}}{{end}}
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


func (pi *PagerdutyIncident)Description() (desc string, err error) {
  var descBuf bytes.Buffer

  shi := pi.Missing.ToSectorHealthByInventory()

  err = piTmpl.Execute(&descBuf, shi)
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