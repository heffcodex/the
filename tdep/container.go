package tdep

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/google/uuid"
)

type Container struct {
	deps     sync.Map // map[string]*D[T]
	closerMu sync.RWMutex
	closers  *orderedmap.OrderedMap[string, CtxCloser]
}

func (c *Container) Add[T any](dep *D[T]) error {
	if _, loaded := c.deps.LoadOrStore(dep.typ, dep); loaded {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, dep.typ)
	}

	return nil
}

func (c *Container) MustAdd[T any](dep *D[T]) {
	if err := c.Add(dep); err != nil {
		panic(err)
	}
}

func (c *Container) Get[T any](ctx context.Context) (T, error) {
	var (
		typ = typeOfT[T]()
		err error
	)

	ctx, err = addToChain(ctx, typ)
	if err != nil {
		return *new(T), err
	}

	if anyDep, ok := c.deps.Load(typ); ok {
		tDep := anyDep.(*D[T]) //nolint:errcheck,revive // ok to panic here

		t, err := tDep.Get(ctx)
		if err == nil {
			c.addCloser(typ, tDep)
		}

		return t, err
	}

	return *new(T), fmt.Errorf("%w: %s", ErrNotRegistered, typ)
}

func (c *Container) MustGet[T any](ctx context.Context) T {
	t, err := c.Get[T](ctx)
	if err != nil {
		panic(err)
	}

	return t
}

func (c *Container) Health(ctx context.Context) error {
	var errs error

	c.deps.Range(func(name, dep any) bool {
		if hc, ok := dep.(CtxHealthChecker); ok {
			if err := hc.Health(ctx); err != nil {
				errs = errors.Join(errs, fmt.Errorf("%s: %w", name.(string), err)) //nolint:errcheck // should never panic
			}
		}

		return true
	})

	return errs
}

func (c *Container) OnClose(fns ...OnCloseFunc) {
	for _, fn := range fns {
		id := "ONCLOSE-" + uuid.NewString()
		c.addCloser(id, fn)
	}
}

func (c *Container) Close(ctx context.Context) (errs error) {
	c.closerMu.Lock()
	defer func() {
		c.closers = nil
		c.closerMu.Unlock()
	}()

	if c.closers == nil {
		return nil
	}

	for typ, closer := range c.closers.AllFromBack() {
		if err := closer.Close(ctx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("%s: %w", typ, err))
		}
	}

	return errs
}

func (c *Container) addCloser(typ string, closer CtxCloser) {
	c.closerMu.RLock()

	if c.closers != nil && c.closers.Has(typ) {
		c.closerMu.RUnlock()
		return
	}

	c.closerMu.RUnlock()
	c.closerMu.Lock()
	defer c.closerMu.Unlock()

	if c.closers == nil {
		c.closers = orderedmap.NewOrderedMap[string, CtxCloser]()
	} else if c.closers.Has(typ) {
		return
	}

	c.closers.Set(typ, closer)
}
