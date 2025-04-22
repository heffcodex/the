package tdep

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/elliotchance/orderedmap/v3"
)

type C interface {
	c() *Container
}

type Container struct {
	deps     sync.Map // map[string]*D[T]
	closerMu sync.RWMutex
	closers  *orderedmap.OrderedMap[string, CtxCloser]
}

func (c *Container) c() *Container {
	return c
}

func (c *Container) registerCloser(typ string, closer CtxCloser) {
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

func Register[T any](ctn C, dep *D[T]) error {
	if _, loaded := ctn.c().deps.LoadOrStore(dep.typ, dep); loaded {
		return fmt.Errorf("already registered: %s", dep.typ) //nolint:err113 // TODO: make static
	}

	return nil
}

func MustRegister[T any](ctn C, dep *D[*T]) {
	if err := Register(ctn, dep); err != nil {
		panic(err)
	}
}

func Get[T any](ctn C) (T, error) {
	typ := typeOfT[T]()

	if anyDep, ok := ctn.c().deps.Load(typ); ok {
		tDep := anyDep.(*D[T]) //nolint:errcheck,revive // ok to panic here

		t, err := tDep.Get()
		if err == nil {
			ctn.c().registerCloser(typ, tDep)
		}

		return t, err
	}

	return *new(T), fmt.Errorf("not found: %s", typ) //nolint:err113 // TODO: make static
}

func Must[T any](ctn C) T {
	t, err := Get[T](ctn)
	if err != nil {
		panic(err)
	}

	return t
}
