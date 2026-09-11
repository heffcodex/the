module github.com/heffcodex/the/dep/bakedin/tdep_grpc

go 1.27.1

replace github.com/heffcodex/the => ../../..

require (
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.4
	github.com/heffcodex/the v0.0.6
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.83.2
)

require (
	github.com/elliotchance/orderedmap/v3 v3.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260908043556-f8649ddbbfe6 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
