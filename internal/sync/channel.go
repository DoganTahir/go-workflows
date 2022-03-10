package sync

import (
	"github.com/cschleiden/go-workflows/internal/converter"
	"github.com/pkg/errors"
)

type Channel[T any] interface {
	Send(ctx Context, v T)

	SendNonblocking(ctx Context, v T) (ok bool)

	Receive(ctx Context) (v T, more bool)

	ReceiveNonblocking(ctx Context) (v T, more bool)

	Close()
}

type ChannelInternal interface {
	Closed() bool

	ReceiveNonBlocking(ctx Context, cb func(v interface{})) (ok bool)

	AddReceiveCallback(cb func(v interface{}))
}

func NewChannel[T any]() Channel[T] {
	return &channel[T]{
		c:         make([]T, 0),
		converter: converter.DefaultConverter,
	}
}

func NewBufferedChannel[T any](size int) Channel[T] {
	return &channel[T]{
		c:         make([]T, 0, size),
		size:      size,
		converter: converter.DefaultConverter,
	}
}

type channel[T any] struct {
	c         []T
	receivers []func(T)
	senders   []func() T
	closed    bool
	size      int
	converter converter.Converter
}

var _ Channel = (*channel)(nil)
var _ ChannelInternal = (*channel)(nil)

func (c *channel) Close() {
	c.closed = true

	// If there are still blocked senders, error
	if len(c.senders) > 0 {
		panic("send on closed channel")
	}

	for len(c.receivers) > 0 {
		r := c.receivers[0]
		c.receivers[0] = nil
		c.receivers = c.receivers[1:]

		r(nil)
	}
}

func (c *channel) Send(ctx Context, v interface{}) {
	cr := getCoState(ctx)

	addedSender := false
	sentValue := false

	for {
		if c.trySend(v) {
			cr.MadeProgress()
			return
		}

		if !addedSender {
			addedSender = true

			cb := func() interface{} {
				sentValue = true
				return v
			}

			c.senders = append(c.senders, cb)
		}

		cr.Yield()

		if sentValue {
			cr.MadeProgress()
			return
		}
	}
}

func (c *channel) SendNonblocking(ctx Context, v interface{}) bool {
	return c.trySend(v)
}

func (c *channel) Receive(ctx Context, vptr interface{}) (more bool) {
	cr := getCoState(ctx)

	addedListener := false
	receivedValue := false

	for {
		// Try to receive from buffered channel or blocked sender
		if c.tryReceive(vptr) {
			cr.MadeProgress()
			return !c.closed
		}

		// Register handler to receive value once
		if !addedListener {
			cb := func(v interface{}) {
				receivedValue = true

				if vptr != nil {
					if err := converter.AssignValue(c.converter, v, vptr); err != nil {
						panic(err)
					}
				}
			}

			c.receivers = append(c.receivers, cb)
			addedListener = true
		}

		cr.Yield()

		// If we received a value via the callback, return
		if receivedValue {
			cr.MadeProgress()
			return !c.closed
		}
	}
}

func (c *channel) ReceiveNonblocking(ctx Context, vptr interface{}) (ok bool) {
	return c.tryReceive(vptr)
}

func (c *channel) hasValue() bool {
	return len(c.c) > 0
}

func (c *channel) canReceive() bool {
	return c.hasValue() || len(c.senders) > 0 || c.closed
}

func (c *channel) trySend(v interface{}) bool {
	// If closed, we can't send, exit.
	if c.closed {
		panic("channel closed")
	}

	// Are there any existing blocked receivers? If so, unblock the first one with
	// the value.
	if len(c.receivers) > 0 {
		r := c.receivers[0]
		c.receivers[0] = nil
		c.receivers = c.receivers[1:]
		r(v)
		return true
	}

	// No waiting receiver, if we have capacity try to add the value to the buffer
	if c.hasCapacity() {
		c.c = append(c.c, v)
		return true
	}

	// No receiver waiting and no capacity, we can't send.
	return false
}

func (c *channel) tryReceive(vptr interface{}) bool {
	// If channel is buffered, return value if available
	if c.hasValue() {
		v := c.c[0]
		c.c = c.c[1:]

		if vptr != nil {
			if err := converter.AssignValue(c.converter, v, vptr); err != nil {
				panic(errors.Wrap(err, "could not assign value when receiving from channel"))
			}
		}

		return true
	}

	// If channel has been closed and no values in buffer (if buffered) return zero
	// element
	if c.closed {
		if vptr != nil {
			if err := converter.AssignValue(c.converter, nil, vptr); err != nil {
				panic(err)
			}
		}

		return true
	}

	if len(c.senders) > 0 {
		s := c.senders[0]
		c.senders[0] = nil
		c.senders = c.senders[1:]

		v := s()

		if vptr != nil {
			if err := converter.AssignValue(c.converter, v, vptr); err != nil {
				panic(err)
			}
		}

		return true
	}

	return false
}

func (c *channel) hasCapacity() bool {
	return len(c.c) < c.size
}

func (c *channel) AddReceiveCallback(cb func(v interface{})) {
	c.receivers = append(c.receivers, cb)
}

func (c *channel) ReceiveNonBlocking(ctx Context, cb func(v interface{})) (ok bool) {
	var vptr interface{}
	if c.tryReceive(vptr) {
		cb(vptr)
		return true
	}

	c.AddReceiveCallback(cb)

	return false
}

func (c *channel) Closed() bool {
	return c.closed
}
