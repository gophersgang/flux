package cqrs

import "testing"

import "github.com/stretchr/testify/assert"

type TestEvent struct {
}

type TestEntity struct {
	Aggregate
	handled bool
}

func (entity *TestEntity) Handle(event TestEvent) {
	entity.handled = true
}

func TestEventHandling(t *testing.T) {
	entity := new(TestEntity)
	entity.Aggregate = NewAggregate(entity)

	entity.Update(TestEvent{})

	assert.True(t, entity.handled)
}

func TestUnknownEvent(t *testing.T) {
	entity := new(TestEntity)
	entity.Aggregate = NewAggregate(entity)

	assert.NotPanics(t, func() { entity.Update("unknown string event") })
	assert.False(t, entity.handled)
}
