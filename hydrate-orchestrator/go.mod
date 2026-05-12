module dagger/hydrate-orchestrator

go 1.23.2

require (
	github.com/gobwas/glob v0.2.3
	github.com/samber/lo v1.47.0
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	golang.org/x/text v0.26.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

replace go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc => go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.12.2

replace go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp => go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.12.2

replace go.opentelemetry.io/otel/log => go.opentelemetry.io/otel/log v0.12.2

replace go.opentelemetry.io/otel/sdk/log => go.opentelemetry.io/otel/sdk/log v0.12.2
