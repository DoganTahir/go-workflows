package sync

import "github.com/cschleiden/go-workflows/internal/converter"

type Future[T any] interface {
	// Set stores the value and unblocks any waiting consumers
	Set(v T, err error)

	// Get returns the value if set, blocks otherwise
	Get(ctx Context) (T, error)
}

func NewFuture[T any]() Future[T] {
	return &futureImpl[T]{
		converter: converter.DefaultConverter,
	}
}

type futureImpl[T any] struct {
	hasValue  bool
	v         T
	err       error
	converter converter.Converter
}

func (f *futureImpl[T]) Set(v T, err error) {
	if f.hasValue {
		panic("future already set")
	}

	f.v = v
	f.err = err
	f.hasValue = true
}

func (f *futureImpl[T]) Get(ctx Context) (T, error) {
	for {
		cr := getCoState(ctx)

		if f.hasValue {
			cr.MadeProgress()

			if f.err != nil {
				var zero T
				return zero, f.err
			}

			var r T
			err := converter.AssignValue(f.converter, f.v, &r)
			return r, err
		}

		cr.Yield()
	}
}

func (f *futureImpl[T]) Ready() bool {
	return f.hasValue
}
