package tcfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoader(t *testing.T) {
	t.Parallel()

	l := NewLoader[BaseConfig](nil)
	assert.Nil(t, l.viper)
}

func TestNewDefaultLoader(t *testing.T) {
	t.Parallel()

	l := NewDefaultLoader[BaseConfig]()
	assert.NotNil(t, l.viper)
}
