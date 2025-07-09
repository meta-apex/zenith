package trace

// TraceName represents the tracing name.
const TraceName = "zrpc"

// A Config is an opentelemetry config.
type Config struct {
	Name     string  `meta:",optional"`
	Endpoint string  `meta:",optional"`
	Sampler  float64 `meta:",default=1.0"`
	Batcher  string  `meta:",default=jaeger,options=jaeger|zipkin|otlpgrpc|otlphttp|file"`
	// OtlpHeaders represents the headers for OTLP gRPC or HTTP transport.
	// For example:
	//  uptrace-dsn: 'http://project2_secret_token@localhost:14317/2'
	OtlpHeaders map[string]string `meta:",optional"`
	// OtlpHttpPath represents the path for OTLP HTTP transport.
	// For example
	// /v1/traces
	OtlpHttpPath string `meta:",optional"`
	// OtlpHttpSecure represents the scheme to use for OTLP HTTP transport.
	OtlpHttpSecure bool `meta:",optional"`
	// Disabled indicates whether StartAgent starts the agent.
	Disabled bool `meta:",optional"`
}
