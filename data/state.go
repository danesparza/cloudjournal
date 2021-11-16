package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/buntdb"
)

type LogState struct {
	Unit       string    `json:"unit"`
	LastCursor string    `json:"last_cursor"`
	LastSynced time.Time `json:"last_synced"`
}

// UpdateLogState updates the log state for a given unit
func (store Manager) UpdateLogState(unit, lastCursor string) (LogState, error) {
	//	Our return item
	retval := LogState{
		Unit:       unit,
		LastCursor: lastCursor,
		LastSynced: time.Now(),
	}

	//	Serialize to JSON format
	encoded, err := json.Marshal(retval)
	if err != nil {
		return retval, fmt.Errorf("problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(GetKey("State", unit), string(encoded), nil)
		return err
	})

	//	Return our data:
	return retval, err
}

// GetLogStateForUnit gets log state for a given unit
func (store Manager) GetLogStateForUnit(unit string) (LogState, error) {
	//	Our return item
	retval := LogState{}

	err := store.systemdb.View(func(tx *buntdb.Tx) error {
		item, err := tx.Get(GetKey("State", unit))
		if err != nil {
			return err
		}

		if len(item) > 0 {
			//	Unmarshal data into our item
			val := []byte(item)
			if err := json.Unmarshal(val, &retval); err != nil {
				return err
			}
		}

		return nil
	})

	//	Return our data:
	return retval, err
}

// GetLogStateForAllUnits gets log state for all units
func (store Manager) GetLogStateForAllUnits() ([]LogState, error) {
	//	Our return item
	retval := []LogState{}

	//	Set our prefix
	prefix := GetKey("State")

	//	Iterate over our values:
	err := store.systemdb.View(func(tx *buntdb.Tx) error {
		tx.Descend(prefix, func(key, val string) bool {

			if len(val) > 0 {
				//	Create our item:
				item := LogState{}

				//	Unmarshal data into our item
				bval := []byte(val)
				if err := json.Unmarshal(bval, &item); err != nil {
					return false
				}

				//	Add to the array of returned users:
				retval = append(retval, item)
			}

			return true
		})
		return nil
	})

	//	Return our data:
	return retval, err
}
