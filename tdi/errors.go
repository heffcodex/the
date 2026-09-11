package tdi

import (
	"errors"
	"slices"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
)

var (
	ErrClosed            = errors.New("already closed")
	ErrAlreadyRegistered = errors.New("already registered")
	ErrNotRegistered     = errors.New("not registered")
)

type CircularDependencyError struct {
	trace []string
}

func newCircularDependencyError(typ string, chain *orderedmap.OrderedMap[string, struct{}]) CircularDependencyError {
	var (
		el    = chain.GetElement(typ)
		trace []string
	)

	for el != nil {
		trace = append(trace, el.Key)
		el = el.Next()
	}

	return CircularDependencyError{trace: append(trace, typ)}
}

func (e CircularDependencyError) Trace() []string {
	return slices.Clone(e.trace)
}

func (e CircularDependencyError) Error() string {
	return "circular dependency: " + strings.Join(e.trace, " -> ")
}

func (e CircularDependencyError) Is(target error) bool {
	//goland:noinspection GoTypeAssertionOnErrors
	_, ok := target.(CircularDependencyError)
	return ok
}
