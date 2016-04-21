package watchdog

import (
  "testing"
)

func TestWatchdogTracker(t *testing.T) {
  at := NewActionTracker()
  at.AdvertisedBy("a-server-1234")

  pm := NewWatchdogTracker()
  pm["action-1234~1"] = at
  pm.markEpoch()

  if len(at.Idents()) != 0 {
    t.Fatalf("expected markEpoch to mark action tracker")
  }
}