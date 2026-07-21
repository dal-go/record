package record

import (
	"fmt"

	"github.com/dal-go/record/update"
)

// Updates is a declarative update command for one record.
type Updates struct {
	Record  Record
	Updates []update.Update
}

// Changes is an in-memory envelope of record mutations to be applied by a DAL facade.
type Changes struct {
	recordsToInsert []Record
	RecordsToUpdate []*Updates
	RecordsToDelete []*Key
}

// RecordsToInsert returns a copy of the queued insert records.
func (v *Changes) RecordsToInsert() []Record {
	if len(v.recordsToInsert) == 0 {
		return v.recordsToInsert
	}
	records := make([]Record, len(v.recordsToInsert))
	copy(records, v.recordsToInsert)
	return records
}

// QueueForInsert adds records to the insert queue.
func (v *Changes) QueueForInsert(records ...Record) {
	for i, record := range records {
		if record == nil {
			panic(fmt.Sprintf("record #%d is required", i))
		}
		key := record.Key()
		if key == nil {
			panic(fmt.Sprintf("record #%d.Key() is required", i))
		}
		if record.Data() == nil {
			panic(fmt.Sprintf("record #%d.Data() is required", i))
		}
		for _, queuedRecord := range v.recordsToInsert {
			if queuedRecord.Key().Equal(key) {
				panic(fmt.Sprintf("record with key=%s is already queued for insert", key))
			}
		}
		v.recordsToInsert = append(v.recordsToInsert, record)
	}
}

// Reset clears all queued changes after successful application.
func (v *Changes) Reset() {
	v.recordsToInsert = nil
	v.RecordsToUpdate = nil
	v.RecordsToDelete = nil
}
