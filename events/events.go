package events

import "context"

type Event struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

type Stream struct {
	channel chan Event

	subcriptions chan *subscription
	unsubcribes  chan *subscription
}

type subscription struct {
	events chan Event
}

func NewStream() *Stream {
	return &Stream{
		channel:      make(chan Event),
		subcriptions: make(chan *subscription),
		unsubcribes:  make(chan *subscription),
	}
}

func (s *Stream) Publish(e Event) {
	s.channel <- e
}

func (s *Stream) Listen(ctx context.Context, f func(e Event)) {
	subscription := s.subscribe(ctx)
	for {
		select {
		case e := <-subscription.events:
			f(e)
		case <-ctx.Done():
			s.unsubcribes <- subscription
			return
		}
	}
}

func (s *Stream) subscribe(ctx context.Context) *subscription {
	sub := &subscription{
		events: make(chan Event),
	}
	s.subcriptions <- sub
	return sub
}

func (s *Stream) Dispatch(ctx context.Context) {
	var subscribers []*subscription
	for {
		select {
		case sub := <-s.subcriptions:
			// New subscribe
			subscribers = append(subscribers, sub)
		case sub := <-s.unsubcribes:
			// Removed subscriber
			idx := -1
			for i := range subscribers {
				if subscribers[i] == sub {
					idx = i
					break
				}
			}
			if idx != -1 {
				close(subscribers[idx].events)
				subscribers = append(subscribers[:idx], subscribers[idx+1:]...)
			}
		case e := <-s.channel:
			for _, s := range subscribers {
				s.events <- e
			}
		case <-ctx.Done():
			// Terminate
			return
		}
	}
}
