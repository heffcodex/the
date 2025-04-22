package tzap

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	KeyMessage     = "msg"
	KeyLevel       = "level"
	KeyTime        = "ts"
	KeyName        = "logger"
	KeyCaller      = "caller"
	KeyFunction    = zapcore.OmitKey
	KeyStacktrace  = "stacktrace"
	KeyHTTPRequest = "http_request"
)

type StdCoreConfig struct {
	zapcore.EncoderConfig
	LevelEnabler zapcore.LevelEnabler
}

func DefaultStdCoreConfig(le zapcore.LevelEnabler) *StdCoreConfig {
	return &StdCoreConfig{
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:          KeyMessage,
			LevelKey:            KeyLevel,
			TimeKey:             KeyTime,
			NameKey:             KeyName,
			CallerKey:           KeyCaller,
			FunctionKey:         KeyFunction,
			StacktraceKey:       KeyStacktrace,
			SkipLineEnding:      false,
			LineEnding:          zapcore.DefaultLineEnding,
			EncodeLevel:         zapcore.CapitalLevelEncoder,
			EncodeTime:          zapcore.RFC3339TimeEncoder,
			EncodeDuration:      zapcore.SecondsDurationEncoder,
			EncodeCaller:        zapcore.ShortCallerEncoder,
			EncodeName:          zapcore.FullNameEncoder,
			NewReflectedEncoder: nil, // uses json.Encoder by default
			ConsoleSeparator:    "\t",
		},
		LevelEnabler: le,
	}
}

func (c *StdCoreConfig) Console() zapcore.Core {
	enc := zapcore.NewConsoleEncoder(c.EncoderConfig)
	return c.core(enc)
}

func (c *StdCoreConfig) JSON() zapcore.Core {
	enc := zapcore.NewJSONEncoder(c.EncoderConfig)
	return c.core(enc)
}

func (c *StdCoreConfig) core(enc zapcore.Encoder) zapcore.Core {
	le := c.LevelEnabler
	if le == nil {
		le = zap.LevelEnablerFunc(func(zapcore.Level) bool { return true })
	}

	return zapcore.NewCore(enc, zapcore.Lock(os.Stderr), le)
}
