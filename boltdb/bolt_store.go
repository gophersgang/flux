package boltdb

import (
	"bytes"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	. "github.com/yehohanan7/flux/cqrs"
)

const (
	EVENTS_BUCKET         = "EVENTS"
	EVENT_METADATA_BUCKET = "EVENT_METADATA"
	AGGREGATES_BUCKET     = "AGGREGATES_BUCKET"
)

//InMemory implementation of the event store
type BoltEventStore struct {
	db *bolt.DB
}

func (store *BoltEventStore) GetEvent(id string) Event {
	var event = new(Event)
	store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_BUCKET))

		if v := b.Get([]byte(id)); v != nil {
			event.Deserialize(v)
		}
		return nil
	})
	return *event
}

func (store *BoltEventStore) GetEvents(aggregateId string) []Event {
	events := make([]Event, 0)
	store.db.View(func(tx *bolt.Tx) error {

		c := tx.Bucket([]byte(AGGREGATES_BUCKET)).Cursor()

		prefix := []byte(aggregateId)

		for k, eventId := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, eventId = c.Next() {
			events = append(events, store.GetEvent(string(eventId)))
		}

		return nil
	})

	return events
}

func (store *BoltEventStore) SaveEvents(aggregateId string, events []Event) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(EVENTS_BUCKET))
		mb := tx.Bucket([]byte(EVENT_METADATA_BUCKET))
		ab := tx.Bucket([]byte(AGGREGATES_BUCKET))

		_, lastMeta := mb.Cursor().Last()

		if lastMeta != nil && len(lastMeta) > 0 {
			m := new(EventMetaData)
			m.Deserialize(lastMeta)
			if m.AggregateVersion+1 != events[0].AggregateVersion {
				return Conflict
			}
		}

		for _, event := range events {
			ak := []byte(aggregateId + "::" + string(event.AggregateVersion))

			if err := ab.Put(ak, []byte(event.Id)); err != nil {
				return err
			}

			bytes := event.Serialize()
			if err := eb.Put([]byte(event.Id), bytes); err != nil {
				return err
			}

			offset, _ := mb.NextSequence()
			bytes = event.EventMetaData.Serialize()
			if err := mb.Put([]byte(strconv.FormatUint(offset, 10)), bytes); err != nil {
				return err
			}
		}
		return nil
	})
}

func (store *BoltEventStore) GetEventMetaDataFrom(offset, count int) []EventMetaData {
	metas := make([]EventMetaData, 0)
	store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENT_METADATA_BUCKET))
		for i := offset; i <= offset+count; i++ {
			v := b.Get([]byte(strconv.Itoa(i)))
			if v != nil && len(v) > 0 {
				m := new(EventMetaData)
				m.Deserialize(v)
				metas = append(metas, *m)
			}
		}
		return nil
	})
	return metas
}

func NewBoltStore(path string) *BoltEventStore {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		glog.Fatal("Error while opening bolt db", err)
	}
	db.Update(func(tx *bolt.Tx) error {
		createBucket(tx, EVENTS_BUCKET)
		createBucket(tx, EVENT_METADATA_BUCKET)
		createBucket(tx, AGGREGATES_BUCKET)
		return nil
	})
	return &BoltEventStore{db}
}
