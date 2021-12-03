package workflow

import (
	"context"
	"sync/atomic"
)

type key int

var coroutinesCtxKey key

func newCoroutine(ctx context.Context, fn func(ctx context.Context)) *coState {
	s := &coState{
		blocking: make(chan bool, 1),
		unblock:  make(chan bool),
	}

	s.blocked.Store(false)
	s.done.Store(false)

	ctx = withCoState(ctx, s)

	go func() {
		defer s.finished() // Ensure we always mark the coroutine as finished
		defer func() {
			// TODO: panic handling
		}()

		fn(ctx)
	}()

	return s
}

type coState struct {
	blocking chan bool
	unblock  chan bool
	blocked  atomic.Value
	done     atomic.Value
}

func (s *coState) finished() {
	s.done.Store(true)
	s.blocking <- true
}

func (s *coState) yield() {
	s.blocked.Store(true)
	s.blocking <- true

	<-s.unblock

	s.blocked.Store(false)
}

func (s *coState) cont() {
	s.unblock <- true

	// TODO: Add some timeout

	// Run until blocked (which is also true when finished)
	select {
	case <-s.blocking:
	}
}

func withCoState(ctx context.Context, s *coState) context.Context {
	return context.WithValue(ctx, coroutinesCtxKey, s)
}

func getCoState(ctx context.Context) *coState {
	s, ok := ctx.Value(coroutinesCtxKey).(*coState)
	if !ok {
		panic("could not find coroutine state")
	}

	return s
}