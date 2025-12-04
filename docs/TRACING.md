# OffGridFlow Tracing Guide

OpenTelemetry distributed tracing for OffGridFlow carbon emissions platform.

## Overview

Tracing provides visibility into:
- AI router request flows (cloud/local provider routing)
- Database operations and query performance
- HTTP API request handling
- Multi-tenant request isolation
- Emission calculations and data processing

## Quick Start

### 1. Start AI Toolkit Trace Collector

Before running the application, open the trace collector in VS Code:

1. Open VS Code Command Palette (`Ctrl+Shift+P` or `Cmd+Shift+P`)
2. Run command: **"AI Toolkit: Open Trace Viewer"**
3. This starts the OTLP collector on `http://localhost:4318`

### 2. Run Your Application

Tracing is enabled by default:

```powershell
# Run the API server
go run ./cmd/api
```

The application will automatically:
- Connect to the AI Toolkit trace collector at `localhost:4318`
- Send traces for all operations
- Display trace endpoint in logs

### 3. View Traces

Open the AI Toolkit trace viewer in VS Code to see:
- Request traces with timing information
- Span hierarchy showing operation flow
- Attributes and events for each operation
- Error tracking and status

## Configuration

### Environment Variables

```powershell
# Enable/disable tracing (enabled by default)
$env:OFFGRIDFLOW_TRACING_ENABLED = "true"

# Custom OTLP endpoint (defaults to http://localhost:4318)
$env:OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318"

# Service environment
$env:OFFGRIDFLOW_ENV = "production"
```

### Trace Sampling

By default, all traces are sampled (100%). To adjust:

Edit `cmd/api/main.go`:
```go
traceProvider, err := tracing.Setup(tracing.Config{
    ServiceName:    "offgridflow-api",
    ServiceVersion: "1.0.0",
    Environment:    cfg.Server.Env,
    OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
    SamplingRate:   0.5, // Sample 50% of traces
    Enabled:        tracingEnabled,
})
```

## Adding Traces to Your Code

### Basic Span

```go
import (
    "github.com/example/offgridflow/internal/tracing"
)

func YourFunction(ctx context.Context) error {
    ctx, span := tracing.StartSpan(ctx, "your-operation-name")
    defer span.End()
    
    // Your code here
    
    return nil
}
```

### Adding Attributes

```go
tracing.SetAttributes(span, map[string]interface{}{
    "user.id": userID,
    "tenant.id": tenantID,
    "operation": "calculate_emissions",
})
```

### Recording Errors

```go
result, err := performOperation(ctx)
if err != nil {
    tracing.RecordError(span, err, "operation failed")
    return err
}
```

### Adding Events

```go
tracing.AddEvent(span, "cache.hit", map[string]interface{}{
    "cache.key": key,
    "cache.type": "redis",
})
```

## Example: Tracing AI Operations

```go
func (r *Router) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
    // Start root span
    ctx, span := tracing.StartSpan(ctx, "ai.router.chat")
    defer span.End()
    
    // Add request attributes
    tracing.SetAttributes(span, map[string]interface{}{
        "ai.mode": r.modeManager.GetMode().String(),
        "ai.request.messages": len(req.Messages),
    })
    
    // Determine provider
    mode := r.modeManager.GetMode()
    if mode == offgrid.ModeOffline {
        tracing.AddEvent(span, "using.local.provider", nil)
        return r.callLocal(ctx, req)
    }
    
    // Try cloud provider
    tracing.AddEvent(span, "using.cloud.provider", nil)
    resp, err := r.callCloud(ctx, req)
    if err != nil {
        tracing.RecordError(span, err, "cloud provider failed")
        
        // Fallback to local
        if r.enableFallback {
            tracing.AddEvent(span, "fallback.to.local", nil)
            return r.callLocal(ctx, req)
        }
        return ChatResponse{}, err
    }
    
    return resp, nil
}
```

## Trace Visualization

### AI Toolkit Trace Viewer

The AI Toolkit provides:
- **Timeline View**: Visual representation of span durations
- **Span Details**: Attributes, events, and metadata
- **Error Tracking**: Highlighted errors with stack traces
- **Performance Analysis**: Timing breakdowns by operation

### Span Attributes

Standard attributes added automatically:
- `service.name`: offgridflow-api
- `service.version`: Application version
- `deployment.environment`: development/staging/production

Custom attributes you can add:
- `tenant.id`: Multi-tenant context
- `user.id`: User identification
- `ai.provider`: cloud/local
- `db.operation`: Database operation type
- `emission.scope`: Scope 1/2/3

## Troubleshooting

### Traces Not Appearing

1. **Check AI Toolkit trace collector is running:**
   ```
   Command Palette → "AI Toolkit: Open Trace Viewer"
   ```

2. **Verify endpoint configuration:**
   ```powershell
   $env:OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318"
   ```

3. **Check application logs:**
   Look for `[offgridflow] tracing enabled` message

4. **Ensure tracing is enabled:**
   ```powershell
   $env:OFFGRIDFLOW_TRACING_ENABLED = "true"
   ```

### Performance Impact

- Tracing adds minimal overhead (<1% for most operations)
- Use sampling in high-traffic production environments
- Span data is batched and sent asynchronously
- Local collector has negligible network latency

### Memory Usage

- Spans are held briefly in memory before export
- Batch exporter flushes every few seconds
- Configure batch size if memory constrained

## Best Practices

### DO:
- ✅ Add traces to critical paths (AI operations, DB queries, API handlers)
- ✅ Include relevant attributes (tenant ID, user ID, operation type)
- ✅ Record errors with context
- ✅ Use child spans for sub-operations
- ✅ Keep span names concise and consistent

### DON'T:
- ❌ Add traces to every function (creates noise)
- ❌ Include sensitive data in attributes (passwords, API keys)
- ❌ Create spans for trivial operations (<1ms)
- ❌ Forget to call `span.End()`
- ❌ Add high-cardinality attributes (unique IDs as span names)

## Integration with Other Systems

### Azure Application Insights

To send traces to Azure:

```go
// Use Azure Monitor exporter instead of OTLP
import "github.com/Azure/azure-sdk-for-go/sdk/monitor/azexporter"

exporter, err := azexporter.New(azexporter.Options{
    ConnectionString: os.Getenv("APPLICATIONINSIGHTS_CONNECTION_STRING"),
})
```

### Jaeger

For Jaeger UI:

```powershell
# Run Jaeger all-in-one
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

# Point to Jaeger
$env:OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318"
```

## Advanced Topics

### Context Propagation

Traces automatically propagate across service boundaries:

```go
// HTTP client - context carries trace
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

### Custom Samplers

Implement custom sampling logic:

```go
type CustomSampler struct {}

func (s *CustomSampler) ShouldSample(
    ctx context.Context,
    traceID trace.TraceID,
    name string,
    spanKind trace.SpanKind,
    attributes []attribute.KeyValue,
    links []trace.Link,
) sdktrace.SamplingResult {
    // Custom logic - e.g., always sample errors
    for _, attr := range attributes {
        if attr.Key == "error" && attr.Value.AsBool() {
            return sdktrace.AlwaysSample()
        }
    }
    return sdktrace.NeverSample()
}
```

## Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)
- [AI Toolkit Tracing Guide](https://github.com/microsoft/vscode-ai-toolkit)
- [Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
