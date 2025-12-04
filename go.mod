module github.com/smallbiznis/corebilling

go 1.22

require (
go.uber.org/fx v1.20.0
go.uber.org/zap v1.26.0
github.com/jackc/pgx/v5 v5.5.4
github.com/joho/godotenv v1.5.1
go.opentelemetry.io/otel v1.26.0
go.opentelemetry.io/otel/sdk v1.26.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.26.0
go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0
google.golang.org/grpc v1.63.2
)
require github.com/golang/protobuf v1.5.3 // indirect
