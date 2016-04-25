package watchdog2

type ExpectedInventory map[string]ExpectedInventoryEntry
type ExpectedInventoryEntry struct {
  Red int `json:"red"`
  Yellow int `json:"yellow"`
}


func NewExpectedInventory() (ei ExpectedInventory) {
  return make(ExpectedInventory)
}

func (ei *ExpectedInventory) DetermineSystemHealth(i Inventory) () {


  return
}