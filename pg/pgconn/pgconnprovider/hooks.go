package pgconnprovider

import (
	"context"

	"github.com/lithammer/shortuuid"
	"github.com/Deimvis/go-ext/go1.25/xcontext"
)

type EventContext interface {
	EventId() EventId
	xcontext.Map
}

// EventId is used to identify related hooks (e.g. start and finish of single operation)
type EventId string

func NewEventContext(ctx context.Context) EventContext {
	return &eventContext{
		eventId: genEventId(),
		Map:     xcontext.NewMap(ctx),
	}
}

type eventContext struct {
	eventId EventId
	xcontext.Map
}

var _ EventContext = (*eventContext)(nil)

func (ec eventContext) EventId() EventId {
	return ec.eventId
}

func genEventId() EventId {
	return EventId(shortuuid.New())
}
