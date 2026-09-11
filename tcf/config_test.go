package tcf

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestBaseConfig_AppName(t *testing.T) {
	t.Parallel()

	fallbackAppName := filepath.Base(os.Args[0])
	t.Log("fallback app name is", fallbackAppName)

	assert.Equal(t, fallbackAppName, Config{}.AppName())
	assert.Equal(t, "foo", Config{App: App{Name: "foo"}}.AppName())
}

func TestBaseConfig_AppKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Key(""), Config{}.AppKey())
	assert.Equal(t, Key("foo"), Config{App: App{Key: "foo"}}.AppKey())
}

func TestBaseConfig_AppEnv(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "dev", Config{}.AppEnv())
	assert.Equal(t, "foo", Config{App: App{Env: "foo"}}.AppEnv())
}

func TestBaseConfig_LogLevel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, zapcore.InfoLevel, Config{}.LogLevel())
	assert.Equal(t, zapcore.DebugLevel, Config{App: App{LogLevel: zapcore.DebugLevel}}.LogLevel())
}

func TestBaseConfig_ShutdownTimeout(t *testing.T) {
	t.Parallel()

	assert.Equal(t, appShutdownTimeoutDefault, Config{}.ShutdownTimeout())
	assert.Equal(t, 5*time.Second, Config{App: App{ShutdownTimeout: 5}}.ShutdownTimeout())
}

func TestBaseConfig_BeforeRead(t *testing.T) {
	t.Parallel()

	assert.NoError(t, Config{}.BeforeRead(nil))
}

func TestBaseConfig_AfterRead(t *testing.T) {
	t.Parallel()

	assert.NoError(t, Config{App: App{Key: testRawKey}}.AfterRead(nil))
	assert.NoError(t, Config{App: App{Key: testB64RawKey}}.AfterRead(nil))
	assert.ErrorContains(t, Config{App: App{Key: ""}}.AfterRead(nil), "app key")
}
