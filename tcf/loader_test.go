package tcf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoader(t *testing.T) {
	t.Parallel()

	l := NewLoader[Config](nil)
	assert.Nil(t, l.viper)
}

func TestNewDefaultLoader(t *testing.T) {
	t.Parallel()

	l := NewDefaultLoader[Config]()
	assert.NotNil(t, l.viper)
}
