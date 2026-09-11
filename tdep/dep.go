package tdep

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrClosed = errors.New("already closed")
)

type (
	Closer interface {
		Close() error
	}
	CtxCloser interface {
		Close(ctx context.Context) error
	}
	CtxHealthChecker interface {
		Health(ctx context.Context) error
	}
)

type OnCloseFunc func(context.Context) error

func (f OnCloseFunc) Close(ctx context.Context) error {
	return f(ctx)
}

type (
	HealthFunc[T any]  func(ctx context.Context, d *D[T]) error
	ResolveFunc[T any] func(opts Params) (T, error)
)

type D[T any] struct {
	mu sync.RWMutex

	typ     string
	params  Params
	health  HealthFunc[T]
	resolve ResolveFunc[T]

	// updated in behavior of Get(), Must() or Close()
	instance T
	resolved bool
	closed   bool
}

func New[T any](resolve ResolveFunc[T], paramFuncs ...ParamFunc) *D[T] {
	return NewWithHealthCheck(resolve, nil, paramFuncs...)
}

func NewWithHealthCheck[T any](resolve ResolveFunc[T], health HealthFunc[T], paramFuncs ...ParamFunc) *D[T] {
	return &D[T]{
		typ:     typeOfT[T](),
		params:  newParams(paramFuncs...),
		health:  health,
		resolve: resolve,
	}
}

func (d *D[T]) Params() Params {
	return d.params
}

func (d *D[T]) Get() (T, error) {
	if d == nil {
		panic("nil dep")
	}

	d.mu.RLock()

	if d.closed {
		defer d.mu.RUnlock()
		return *new(T), ErrClosed
	}

	if d.params.singleton && d.resolved {
		defer d.mu.RUnlock()
		return d.instance, nil
	}

	d.mu.RUnlock()
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return *new(T), ErrClosed
	}

	if !d.params.singleton || !d.resolved {
		instance, err := d.resolve(d.params)
		if err != nil {
			return *new(T), err
		}

		d.instance = instance
		d.resolved = true

		d.debugWrite("resolved")
	}

	return d.instance, nil
}

func (d *D[T]) Must() T {
	v, err := d.Get()
	if err != nil {
		panic(err)
	}

	return v
}

func (d *D[T]) Health(ctx context.Context) error {
	if d.health == nil {
		return nil
	}

	return d.health(ctx, d)
}

func (d *D[T]) Closed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.closed
}

func (d *D[T]) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if !d.resolved {
		d.debugWrite("close (nop: unresolved)")
		return nil
	}

	defer func() {
		d.instance = *new(T)
		d.resolved = false
		d.closed = true
	}()

	switch ityp := any(d.instance).(type) {
	case Closer:
		d.debugWrite("close (closer)")
		return ityp.Close()
	case CtxCloser:
		d.debugWrite("close (ctxCloser)")
		return ityp.Close(ctx)
	default:
		d.debugWrite("close (nop: no closer)")
		return nil
	}
}

func (d *D[T]) debugWrite(msg string) {
	if d.params.IsDebug() {
		d.params.Log().Named(d.typ).Debug(msg)
	}
}
