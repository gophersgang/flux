package boltdb

import (
	"encoding/gob"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/yehohanan7/flux/cqrs"
)

const DB_PATH = "flux_test.db"

type EventPayload struct {
	Data string
}

var _ = Describe("Bolt Event Store", func() {

	var store EventStore

	BeforeEach(func() {
		gob.Register(EventPayload{})
		store = NewBoltStore(DB_PATH)
	})

	AfterEach(func() {
		os.Remove(DB_PATH)
	})

	It("Should save events", func() {
		expected := NewEvent("sample_aggregate", 1, EventPayload{"payload"})

		err := store.SaveEvents("aggregate-1", []Event{expected})

		actual := store.GetEvent(expected.Id)
		Expect(err).To(BeNil())
		Expect(actual.Id).To(Equal(expected.Id))
		Expect(actual.Payload).To(Equal(expected.Payload))
	})

	It("Should save event metadata", func() {
		e1 := NewEvent("sample_aggregate", 0, EventPayload{"payload"})
		e2 := NewEvent("sample_aggregate", 1, EventPayload{"payload"})

		err := store.SaveEvents("aggregate-1", []Event{e1, e2})

		Expect(err).To(BeNil())
		Expect(store.GetEventMetaDataFrom(0, 1)).To(HaveLen(1))
		Expect(store.GetEventMetaDataFrom(0, 2)).To(HaveLen(2))
		Expect(store.GetEventMetaDataFrom(0, 3)).To(HaveLen(2))
	})

	It("Should retrieve event meta data with all attributes", func() {
		event := NewEvent("sample_aggregate", 0, EventPayload{"payload"})

		err := store.SaveEvents("aggregate-1", []Event{event})

		Expect(err).To(BeNil())

		meta := store.GetEventMetaDataFrom(0, 1)[0]
		Expect(meta).To(Equal(event.EventMetaData))
	})

})
