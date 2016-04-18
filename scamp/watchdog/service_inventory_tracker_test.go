package watchdog

import (
  "testing"
)

func TestServiceInventoryTracker(t *testing.T) {
  at := NewActionTracker()
  at.AdvertisedBy("a-server-1234")

  pm := NewServiceInventoryTracker()
  pm["action-1234~1"] = at
  pm.markEpoch()

  if len(at.Idents()) != 0 {
    t.Fatalf("expected markEpoch to mark registered action trackers")
  }
}