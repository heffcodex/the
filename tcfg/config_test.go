package tcfg

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseConfig_AppName(t *testing.T) {
	t.Parallel()

	fallbackAppName := filepath.Base(os.Args[0])
	t.Log("fallback app name is", fallbackAppName)

	assert.Equal(t, fallbackAppName, BaseConfig{}.AppName())
	assert.Equal(t, "foo", BaseConfig{App: App{Name: "foo"}}.AppName())
}

func TestBaseConfig_AppKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Key(""), BaseConfig{}.AppKey())
	assert.Equal(t, Key("foo"), BaseConfig{App: App{Key: "foo"}}.AppKey())
}

func TestBaseConfig_AppEnv(t *testing.T) {
	t.Parallel()

	assert.Equal(t, EnvDev, BaseConfig{}.AppEnv())
	assert.Equal(t, Env("foo"), BaseConfig{App: App{Env: "foo"}}.AppEnv())
}

func TestBaseConfig_LogLevel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "info", BaseConfig{}.LogLevel())
	assert.Equal(t, "foo", BaseConfig{App: App{LogLevel: "foo"}}.LogLevel())
}

func TestBaseConfig_ShutdownTimeout(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 30*time.Second, BaseConfig{}.ShutdownTimeout())
	assert.Equal(t, 5*time.Second, BaseConfig{App: App{ShutdownTimeout: 5}}.ShutdownTimeout())
}

func TestBaseConfig_BeforeRead(t *testing.T) {
	t.Parallel()

	assert.NoError(t, BaseConfig{}.BeforeRead(nil))
}

func TestBaseConfig_AfterRead(t *testing.T) {
	t.Parallel()

	assert.NoError(t, BaseConfig{App: App{Key: testRawKey}}.AfterRead(nil))
	assert.NoError(t, BaseConfig{App: App{Key: testB64RawKey}}.AfterRead(nil))
	assert.ErrorContains(t, BaseConfig{App: App{Key: ""}}.AfterRead(nil), "app key")
}
