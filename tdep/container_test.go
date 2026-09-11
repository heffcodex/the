package tdep

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContainer_CircularDependencyDetection(t *testing.T) {
	t.Parallel()

	var (
		ctn Container
		err error
	)

	type (
		A struct{}
		B struct{}
	)

	err = ctn.Add(New(func(ctx context.Context, _ Params) (A, error) {
		if _, err := ctn.Get[B](ctx); err != nil {
			return A{}, err
		}

		return A{}, nil
	}))
	require.NoError(t, err)

	err = ctn.Add(New(func(ctx context.Context, _ Params) (B, error) {
		if _, err := ctn.Get[A](ctx); err != nil {
			return B{}, err
		}

		return B{}, nil
	}))
	require.NoError(t, err)

	_, err = ctn.Get[A](t.Context())
	require.ErrorIs(t, err, CircularDependencyError{})

	e, ok := errors.AsType[CircularDependencyError](err)
	require.True(t, ok)
	require.Equal(t, []string{"tdep.A", "tdep.B", "tdep.A"}, e.trace)

	t.Log(err.Error())
}
