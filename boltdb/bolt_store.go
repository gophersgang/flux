package boltdb

import (
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	. "github.com/yehohanan7/flux/cqrs"
)

const (
	EVENTS_BUCKET         = "EVENTS"
	EVENT_METADATA_BUCKET = "EVENT_METADATA"
)

//InMemory implementation of the event store
type BoltEventStore struct {
	db *bolt.DB
}

func (store *BoltEventStore) GetEvent(id string) Event {
	var event Event
	store.db.View(func(tx *bolt.Tx) error {
		eventsBucket := tx.Bucket([]byte(EVENTS_BUCKET))
		if e := eventsBucket.Get([]byte(id)); e != nil {
			event = decode(e)
		}
		return nil
	})
	return event
}

func (store *BoltEventStore) GetEvents(aggregateId string) []Event {
	return nil
}

func (store *BoltEventStore) SaveEvents(aggregateId string, events []Event) error {
	store.db.Update(func(tx *bolt.Tx) error {
		eventsBucket := tx.Bucket([]byte(EVENTS_BUCKET))
		for _, event := range events {
			err := eventsBucket.Put([]byte(event.Id), encode(event))
			if err != nil {
				glog.Error("error while saving event")
				return err
			}
		}
		return nil
	})
	return nil
}

func (store *BoltEventStore) GetEventMetaDataFrom(offset, count int) []EventMetaData {
	return nil
}

func NewBoltStore(path string) *BoltEventStore {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		glog.Fatal("Error while opening bolt db", err)
	}
	db.Update(func(tx *bolt.Tx) error {
		create := func(name string) {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				glog.Fatal("Error while initializing db", err)
			}
		}
		create(EVENTS_BUCKET)
		create(EVENT_METADATA_BUCKET)
		return nil
	})
	return &BoltEventStore{db}
}
