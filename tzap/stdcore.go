package tzap

import (
	"os"

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

type CoreFunc func(le zapcore.LevelEnabler) zapcore.Core

func DefaultCore(newEncoderFunc func(cfg zapcore.EncoderConfig) zapcore.Encoder) CoreFunc {
	enc := newEncoderFunc(zapcore.EncoderConfig{
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
	})

	return func(le zapcore.LevelEnabler) zapcore.Core {
		return zapcore.NewCore(enc, zapcore.Lock(os.Stderr), le)
	}
}
