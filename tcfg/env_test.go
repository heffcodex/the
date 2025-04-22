package tcfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "some", Env("some").String())
	assert.Equal(t, "dev", EnvDev.String())
	assert.Equal(t, "test", EnvTest.String())
	assert.Equal(t, "staging", EnvStage.String())
	assert.Equal(t, "production", EnvProd.String())
}
