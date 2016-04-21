package watchdog

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

func TriggerEvent(serviceKeyPath string, details interface{}) (err error) {
  serviceKey,err := ioutil.ReadFile(serviceKeyPath)
  if err != nil {
    return
  }

  event := PagerDutyEvent {
    ServiceKey: string(serviceKey),
    EventType: "trigger",
    Description: "testing for now",
    IncidentKey: "the-one-true-key",
    Details: details,
  }

  eventBytes,err := json.Marshal(event)
  if err != nil {
    return
  }
  eventString := string(eventBytes)
  eventReader := strings.NewReader(eventString)

  resp,err := http.Post(`https://events.pagerduty.com/generic/2010-04-15/create_event.json`, "application/json", eventReader)
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