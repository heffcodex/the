package tdi

import (
	"context"
	"reflect"
	"sync"

	"github.com/elliotchance/orderedmap/v3"
)

func typeNameOf[T any]() string {
	tof := reflect.TypeFor[T]()
	if tof.Kind() == reflect.Pointer {
		tof = tof.Elem()
	}

	return tof.String()
}

func addToChain(ctx context.Context, typ string) (context.Context, error) {
	type (
		depChainKey struct{}
		depChain    struct {
			sync.Mutex

			m *orderedmap.OrderedMap[string, struct{}]
		}
	)

	chain, ok := ctx.Value(depChainKey{}).(*depChain)
	if ok {
		chain.Lock()
		defer chain.Unlock()

		if chain.m.Has(typ) {
			return nil, newCircularDependencyError(typ, chain.m)
		}
	} else {
		chain = &depChain{m: orderedmap.NewOrderedMap[string, struct{}]()}
		ctx = context.WithValue(ctx, depChainKey{}, chain)
	}

	chain.m.Set(typ, struct{}{})
	return ctx, nil
}
