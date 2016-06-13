package watchdog2

import (
  "testing"
)

/*
{
  "main": [
    {
      "ActionName": "Inventory.Count.AddContainers~0",
      "MinHealthyCount": 4,
      "CurrentCount": 2,
      "HealthLabel": "yellow",
      "MissingInstances": [
        {
          "hostname": "10.240.70.210",
          "service_url": "beepish+tls://10.240.70.210:30152"
        },
        {
          "hostname": "10.240.242.23",
          "service_url": "beepish+tls://10.240.242.23:30275"
        }
      ]
    }
  ]
}
*/

func TestFormatting(t *testing.T) {
  var sh SectorHealth = SectorHealth {
    "main": []DegradedAction {
      {
        ActionName: "Inventory.Count.AddContainers~0",
        MinHealthyCount: 4,
        CurrentCount: 2,
        HealthLabel: "yellow",
        MissingInstances: []DegradedService {
          {ServiceName: "logger", ShortHostname: "p1-logging"},
          {ServiceName: "logger", ShortHostname: "p2-logging"},
        },
      },
    },
  }

  var inc Incident = NewPagerdutyIncident("", sh)
  str,err := inc.Description()
  if err != nil {
    t.Fatalf(err.Error())
  }

  
  if exp := `hey`; str != exp {
    t.Fatalf("expected `%s` got `%s`", exp, str)
  }

}