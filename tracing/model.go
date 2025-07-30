package tracing

type TracingConfig struct {
	ServiceName  string
	OTLPEndpoint string
	Environment  string
}

func NewTracingConfig(serviceName, OTLPEndpoint, environment string) TracingConfig {
	return TracingConfig{
		ServiceName:  serviceName,
		OTLPEndpoint: OTLPEndpoint,
		Environment:  environment,
	}
}
