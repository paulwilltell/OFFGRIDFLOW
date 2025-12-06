# JSON Logging Configuration Guide

## Current Status

OffGridFlow imports `log/slog` but uses plain `log.Printf()` calls, which outputs text format, not JSON.

## Quick Fix: Enable JSON Logging

Add this to `cmd/api/main.go` at the start of `run()` function:

```go
func run() (err error) {
	// Set up JSON structured logging
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
	
	// Recover from any panics...
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC", "error", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	
	// Continue with rest of function...
}
```

## Replace log.Printf with slog

Find all instances of:
```go
log.Printf("[offgridflow] message: %v", value)
```

Replace with:
```go
slog.Info("message", "key", value)
```

## Example Replacements

**Before**:
```go
log.Printf("[offgridflow] booting api (env=%s port=%d)", cfg.Server.Env, cfg.Server.Port)
```

**After**:
```go
slog.Info("booting api", 
    "env", cfg.Server.Env, 
    "port", cfg.Server.Port)
```

## Quick Script to Enable JSON Logging

Run this in PowerShell:

```powershell
cd C:\Users\pault\OffGridFlow

# Create a simple patch file
@'
// Add at start of run() function after error recovery setup

	// Set up JSON structured logging
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		AddSource: true, // Include file:line in logs
	})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
	
	logger.Info("offgridflow starting", "component", "api")
'@ | Out-File -FilePath scripts\enable-json-logging.txt

Write-Host "✅ JSON logging setup guide created"
Write-Host "To enable: Add the code from scripts\enable-json-logging.txt to cmd/api/main.go"
```

## Verification

After implementing, run:
```bash
go run cmd/api/main.go 2>&1 | head -5
```

Expected output (JSON format):
```json
{"time":"2025-12-05T10:00:00.000Z","level":"INFO","msg":"offgridflow starting","component":"api"}
{"time":"2025-12-05T10:00:00.100Z","level":"INFO","msg":"booting api","env":"development","port":8080}
{"time":"2025-12-05T10:00:00.200Z","level":"INFO","msg":"connected to Postgres"}
```

## Current Logging Examples

The codebase currently has JSON logging ready via `slog.Default()` usage:

```go
// Already using slog in some places:
Logger: slog.Default(),  // Line 167
```

Just need to:
1. Set default handler to JSON at startup
2. Replace `log.Printf` with `slog.Info/Warn/Error`

## STATUS

⚠️ **PARTIAL IMPLEMENTATION**
- ✅ slog imported and used in some places
- ❌ Default handler not set to JSON
- ❌ Most logs still use plain log.Printf

**Action Required**: Add 5 lines of code to enable JSON logging (see quick fix above)
