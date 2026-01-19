# How to use APMPlus Go SDK

### 1. Init OTel SDK

Refer to the official OpenTelemetry manual [Getting-Started](https://opentelemetry.io/docs/languages/go/getting-started/). Or you can also refer to the following code [example](../instrumentation/trpc.group/trpc-go/oteltrpc/example/otel.go).

### 2. Use APMPlus Library package

#### 2.1. TRPC Instrumentation
1. import instrumentation package
```go
package main
import (
    _ "github.com/volcengine/apmplus-opentelemetry-go-contrib/instrumentation/trpc.group/trpc-go/oteltrpc"
)
```
2. configure trpc filter
```yaml
# client-conf.yaml
client:
  filter:
    - apmplus         # add apmplus filter

# server-conf.yaml
server:
  filter:
    - apmplus         # add apmplus filter
```

### 3. Configure otel sdk

| Environment                 | Description                                                                         |
|-----------------------------|-------------------------------------------------------------------------------------|
| OTEL_SERVICE_NAME           | Service name.                                                                       |
| OTEL_RESOURCE_ATTRIBUTES    | Service attributes, the format is `"attr1=v1,attr2=v2"`.                            |
| OTEL_METRICS_EXPORTER       | Metrics exporter, default is `otlp`, support: `otlp`,`console`,`none`.              |
| OTEL_TRACES_EXPORTER        | Traces exporter, default is `otlp`, support: `otlp`,`console`,`none`.               |
| OTEL_LOGS_EXPORTER          | Logs exporter, default is `otlp`,support: `otlp`,`console`,`none`.                  |
| OTEL_EXPORTER_OTLP_PROTOCOL | Export OTLP protocol, default is `http/protobuf`, support: `grpc`, `http/protobuf`. |
| OTEL_EXPORTER_OTLP_ENDPOINT | Export OTLP endpoint, such as: `apmplus-cn-beijing.volces.com:4317`.                |
| OTEL_EXPORTER_OTLP_HEADERS  | Custom OTLP request headers, the format is `"k1=v1,k2=v2"`.                         |