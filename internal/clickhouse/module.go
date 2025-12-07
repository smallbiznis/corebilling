package clickhouse

import "go.uber.org/fx"

// Module wires ClickHouse ingestion into the Fx application graph.
var Module = fx.Options(
	fx.Provide(
		LoadConfig,
		NewClient,
		NewWriter,
	),
	fx.Invoke(Run),
)
