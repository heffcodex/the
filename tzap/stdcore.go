package tzap

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type StdCoreConfig struct {
	zapcore.EncoderConfig
	LevelEnabler zapcore.LevelEnabler
}

func DefaultStdCoreConfig(le zapcore.LevelEnabler) StdCoreConfig {
	return StdCoreConfig{
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

	leOut := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return le.Enabled(lvl) && lvl < zapcore.ErrorLevel
	})
	leErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return le.Enabled(lvl) && lvl >= zapcore.ErrorLevel
	})

	return zapcore.NewTee(
		zapcore.NewCore(enc, zapcore.Lock(os.Stdout), leOut),
		zapcore.NewCore(enc, zapcore.Lock(os.Stderr), leErr),
	)
}
