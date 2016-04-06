package watchdog

type ActionTracker struct {
  identsAActive bool
  identsA []string
  identsB []string
}

func NewActionTracker() (at *ActionTracker) {
  at = new(ActionTracker)
  at.identsA = make([]string,0)
  at.identsB = make([]string,0)
  at.identsAActive = true

  return
}

func (at *ActionTracker)InstanceCount() (count int) {
  var identsPtr *[]string
  if at.identsAActive {
    identsPtr = &at.identsA
  } else {
    identsPtr = &at.identsB
  }

  return len(*identsPtr)
}

func (at *ActionTracker)AdvertisedBy(ident string) {
  var identsPtr *[]string
  if at.identsAActive {
    identsPtr = &at.identsA
  } else {
    identsPtr = &at.identsB
  }

  found := false
  for _,entry := range *identsPtr {
    if entry == ident {
      found = true
      break
    }
  }

  if !found {
    *identsPtr = append(*identsPtr, ident)
  }
}

func (at *ActionTracker)Idents() ([]string) {
  if at.identsAActive {
    return at.identsA
  } else {
    return at.identsB
  }
}

func (at *ActionTracker)IdentsBefore() ([]string) {
  if !at.identsAActive {
    return at.identsA
  } else {
    return at.identsB
  }
}



func (at *ActionTracker)markEpoch() {
  at.identsAActive = !at.identsAActive
}

func (at *ActionTracker)ClearIdents() {
  // We reset the lists of to-be-filled in data
  if at.identsAActive {
    at.identsA = at.identsA[:0]
  } else {
    at.identsB = at.identsB[:0]
  }
}

func (at *ActionTracker)MissingIdentsThisEpoch() (missing []string) {
  before := at.IdentsBefore()
  now := at.Idents()

  for _,beforeIdent := range before {
    found := false

    for _,i := range now {
      if beforeIdent == i {
        found = true
      }
    }

    if !found {
      missing = append(missing,beforeIdent)
    }
  }

  return
}