package sync

func Select(ctx Context, cases ...selectorCase) {
	s := &selector{
		cases: cases,
	}

	s.Select(ctx)
}

type selector struct {
	cases []selectorCase
}

func SelectFuture[T any](f Future[T], handler func(ctx Context, f Future[T])) selectorCase {
	return &futureCase[T]{
		f:  f.(*futureImpl[T]),
		fn: handler,
	}
}

func ReceiveChannel[T any](c Channel[T], handler func(ctx Context, c Channel[T])) selectorCase {
	return &channelCase{
		c:  channel,
		fn: handler,
	}
}

func Default(handler func(ctx Context)) selectorCase {
	return &defaultCase{
		fn: handler,
	}
}

func Select(ctx Context, cases ...SelectCase) {
	cs := getCoState(ctx)

	for {
		// Is any case ready?
		for _, c := range cases {
			if c.Ready() {
				c.Handle(ctx)
				return
			}
		}

		// else, yield and wait for result
		cs.Yield()
	}
}

type selectorCase interface {
	Ready() bool
	Handle(ctx Context)
}

type futureCase[T any] struct {
	f  *futureImpl[T]
	fn func(Context, Future[T])
}

func (fc *futureCase[T]) Ready() bool {
	return fc.f.Ready()
}

func (fc *futureCase[T]) Handle(ctx Context) {
	fc.fn(ctx, fc.f)
}

var _ = SelectCase(&channelCase{})

type channelCase struct {
	c  *channel
	fn func(Context, Channel)
}

func (cc *channelCase) Ready() bool {
	return cc.c.canReceive()
}

func (cc *channelCase) Handle(ctx Context) {
	cc.fn(ctx, cc.c)
}

type defaultCase struct {
	fn func(Context)
}

func (dc *defaultCase) Ready() bool {
	return true
}

func (dc *defaultCase) Handle(ctx Context) {
	dc.fn(ctx)
}
