package watchdog2

import (
  "fmt"
  "strings"

  "io/ioutil"
  "net/http"
  "encoding/json"

)

type PagerDutyEventContext struct {
  Href string `json:"href"`
  Type string `json:"type"`
}

type PagerDutyEvent struct {
  ServiceKey  string `json:"service_key"`
  Description string `json:"description"`
  EventType   string `json:"event_type"`

  Client    string `json:"client,omitempty"` // optional
  ClientURL string `json:"client_url,omitempty"` // optional
  Contexts  []PagerDutyEventContext `json:"contexts,omitempty"` // optional
  Details     interface{} `json:"details,omitempty"` // optional
  IncidentKey string `json:"incident_key,omitempty"` // optional
}

var incidentKey string = `the-one-true-key`
var apiUrl string = `https://events.pagerduty.com/generic/2010-04-15/create_event.json`

func TriggerEvent(pdKey, description string, details interface{}) (key string, err error) {
  detailsBytes,err := json.MarshalIndent(details, "", "  ")
  if err != nil {
    return
  }
  fmt.Println("sending details: ", string(detailsBytes))

  err = postPDEvent(pdKey, &PagerDutyEvent {
    EventType: "trigger",
    IncidentKey: incidentKey,
    Description: description,
    Details: details,
  })

  return incidentKey, err
}

func ResolveEvent(pagerdutyKey, incidentId, reason string) (err error) {
  err = postPDEvent(pagerdutyKey, &PagerDutyEvent {
    EventType: "resolve",
    IncidentKey: incidentId,
    Description: reason,
  })
  return
}

func postPDEvent(pagerdutyKey string, evt *PagerDutyEvent) (err error) {
  evt.ServiceKey = pagerdutyKey
  eventBytes,err := json.Marshal(evt)
  if err != nil {
    return
  }
  eventString := string(eventBytes)
  eventReader := strings.NewReader(eventString)

  resp,err := http.Post(apiUrl, "application/json", eventReader)
  if err != nil {
    return
  } else if resp.StatusCode != 200 {
    responseBytes,err := ioutil.ReadAll(resp.Body)
    if err != nil {
      err = fmt.Errorf("pager duty request failed and I can't read the message sent back: `%s`", err)
      return err
    }
    err = fmt.Errorf("did not get 200 from pagerduty: `%s`", responseBytes)
    return err
  }

  responseBytes,err := ioutil.ReadAll(resp.Body)
  if err != nil {
    err = fmt.Errorf("couldn't read pager duty response: `%s`", err)
    return err
  }
  fmt.Printf("%s\n", responseBytes)

  return
}
