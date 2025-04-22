package tdep

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

var (
	ErrClosed = errors.New("dep is already closed")
)

type (
	Closer interface {
		Close() error
	}
	CtxCloser interface {
		Close(ctx context.Context) error
	}
)

type (
	HealthFunc[T any]  func(ctx context.Context, d *D[T]) error
	ResolveFunc[T any] func(opts OptSet) (T, error)
)

type D[T any] struct {
	mu      sync.Mutex
	typ     string
	opts    OptSet
	health  HealthFunc[T]
	resolve ResolveFunc[T]

	// updated in behaviour of Get(), MustGet() or Close()
	instance T
	resolved bool
	closed   bool
}

func New[T any](resolve ResolveFunc[T], options ...Option) *D[T] {
	return &D[T]{
		typ:     typeOfT[T](),
		opts:    newOptSet(options...),
		resolve: resolve,
	}
}

func (d *D[T]) WithHealthCheck(fn HealthFunc[T]) *D[T] {
	d.health = fn
	return d
}

func (d *D[T]) Options() OptSet {
	return d.opts
}

func (d *D[T]) Get() (T, error) {
	if d == nil {
		panic("nil dep")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return *new(T), ErrClosed
	}

	if !d.opts.singleton || !d.resolved {
		instance, err := d.resolve(d.opts)
		if err != nil {
			return *new(T), err
		}

		d.instance = instance
		d.resolved = true

		d.debugWrite("resolved")
	}

	return d.instance, nil
}

func (d *D[T]) MustGet() T {
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
	d.mu.Lock()
	defer d.mu.Unlock()

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
	if d.opts.IsDebug() {
		d.opts.Log().Debug(msg, zap.String("typ", d.typ))
	}
}
