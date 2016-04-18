package watchdog

import (
  "testing"
)

func TestActionTracker(t *testing.T) {
  at := NewActionTracker()
  if at.InstanceCount() != 0 {
    t.Fatalf("action tracker should start empty")
  }

  // We sweep the discovery cache and see one ident
  at.AdvertisedBy("a-server-1234")
  if at.InstanceCount() != 1 {
    t.Fatalf("action tracker should count ident")
  } else if idents := at.Idents(); len(idents) != 1 && idents[0] != "a-server-1234" {
    t.Fatalf("should remember exact name")
  }

  // We are done scanning so we mark the epoch/sweep as done
  at.markEpoch()
  if at.InstanceCount() != 0 {
    t.Fatalf("action tracker should be cleared for new epoch")
  } else if idents := at.Idents(); len(idents) != 0 {
    t.Fatalf("should not return idents")
  }

  // We initiate another scan and see a DIFFERENT ident
  at.AdvertisedBy("b-server-1234")
  if at.InstanceCount() != 1 {
    t.Fatalf("action tracker should count ident")
  } else if idents := at.Idents(); len(idents) != 1 || idents[0] != "b-server-1234" {
    t.Fatalf("should remember exact name")
  }

  missing := at.MissingIdentsThisEpoch()
  if len(missing) != 1 || missing[0] != "a-server-1234" {
    t.Fatalf("we expected to be missing the first ident we saw")
  }
}