package cqrs

import (
	"encoding/gob"
	"reflect"

	"github.com/golang/glog"
)

//Aggregate as in the DDD world
type Aggregate struct {
	Id       string
	name     string
	Version  int
	Events   []Event
	entity   interface{}
	handlers Handlers
	store    EventStore
}

//Save the events accumulated so far
func (aggregate *Aggregate) Save() error {
	err := aggregate.store.SaveEvents(aggregate.Id, aggregate.Events)
	if err != nil {
		glog.Warning("error while saving events for aggregate ", aggregate, err)
	}
	aggregate.Events = []Event{}
	return err
}

//Update the event
func (aggregate *Aggregate) Update(payloads ...interface{}) {
	for _, payload := range payloads {
		event := NewEvent(aggregate.name, aggregate.Version, payload)
		aggregate.Events = append(aggregate.Events, event)
		aggregate.Apply(event)
	}
}

//Apply events
func (aggregate *Aggregate) Apply(events ...Event) {
	for _, e := range events {
		payload := e.Payload
		if handler, ok := aggregate.handlers[reflect.TypeOf(payload)]; ok {
			handler(aggregate.entity, payload)
			aggregate.Version++
		}
	}
}

//Create new aggregate with a backing event store
func NewAggregate(id string, entity interface{}, store EventStore) Aggregate {
	hm := NewHandlers(entity)
	for eventType := range hm {
		gob.Register(reflect.New(eventType))
	}

	aggregate := Aggregate{
		Id:       id,
		Version:  0,
		Events:   []Event{},
		entity:   entity,
		handlers: hm,
		store:    store,
		name:     reflect.TypeOf(entity).String(),
	}

	aggregate.Apply(store.GetEvents(id)...)
	return aggregate
}