package tzap

import "go.uber.org/zap/zapcore"

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
