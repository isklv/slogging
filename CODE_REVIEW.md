# Code Review: slogging Library

## Executive Summary

Comprehensive review of the `github.com/isklv/slogging` library focusing on:
- **Critical Issues**: Error handling, panic safety
- **Performance**: Logging overhead, memory allocation
- **Style**: Go idioms, documentation
- **Features**: Missing functionality, edge cases
- **Testing**: Coverage gaps, test quality

---

## Critical Issues 🔴

### 1. Error Handling in `InGraylog()` (Fixed)
**File:** `options.go:InGraylog()`
**Severity:** Critical

**Issue:** `log.Fatal(err)` on invalid Graylog URL crashes entire application.

**Current:**
```go
func (c *LoggerOptions) InGraylog(graylogURL, containerName string) *LoggerOptions {
    w, err := gelf.NewWriter(graylogURL)
    if err != nil {
        log.Fatal(err)  // ❌ Crashes app
    }
    // ...
}
```

**Fixed:**
```go
func (c *LoggerOptions) InGraylog(graylogURL, containerName string) *LoggerOptions {
    if graylogURL == "" {
        return c  // Graceful degradation
    }
    w, err := gelf.NewWriter(graylogURL)
    if err != nil {
        log.Printf("[slogging] Failed to connect to Graylog %s: %v", graylogURL, err)
        return c  // Continue without Graylog
    }
    // ...
}
```

### 2. Missing Nil Checks
**File:** `logger.go`, `graylog.go`
**Severity:** Medium

**Issue:** No nil checks on `opts.inGraylog` before use.

**Recommendation:**
```go
if opts.inGraylog != nil && opts.inGraylog.w != nil {
    // Safe to write
}
```

### 3. Panic on Empty Context
**File:** `context.go`, `alias.go`
**Severity:** Medium

**Issue:** `L(ctx)` might panic if context has no logger and no default set.

**Recommendation:** Add fallback:
```go
func L(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(loggerContextKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()  // Fallback
}
```

---

## Performance ⚡

### 1. UDP Batching Efficiency
**File:** `graylog.go`

**Current:** Uses `slog-graylog` library for UDP batching (good).

**Improvement:** Configure batch size and timeout:
```go
type gelfData struct {
    w             *gelf.Writer
    containerName string
    batchSize     int
    timeout       time.Duration
}
```

### 2. String Formatting Overhead
**File:** `alias.go`

**Issue:** `StringAttr`, `IntAttr`, etc. create new strings for every log call.

**Recommendation:** Use `slog.String`, `slog.Int` directly or pre-allocate:
```go
// Instead of:
slogging.StringAttr("key", "value")

// Use:
slog.String("key", "value")
```

### 3. Context Value Allocation
**File:** `http/*`, `grpc/*`

**Issue:** Each middleware creates new context with `context.WithValue` (allocates).

**Recommendation:** Use context keys from sync.Pool or reuse keys.

### 4. Logger Creation
**File:** `logger.go`

**Issue:** `NewLogger` creates new handlers on every call.

**Recommendation:** Cache handlers or use singleton pattern for standard configs.

---

## Style & Go Idioms 📐

### 1. Error Naming Convention
**File:** Throughout

**Issue:** Errors should start with `Err` noun, not verb.

**Current:** `ErrAttr` (OK), but consider `NewStringAttr` pattern vs `StringAttr`.

**Recommendation:** Stick to `StringAttr`, `IntAttr` pattern (it's fine), or use `slog.String` directly.

### 2. Interface Design
**File:** `logger.go`

**Issue:** `SLogger` struct embeds `*slog.Logger` but also has methods that shadow it.

**Recommendation:**
```go
type SLogger struct {
    *slog.Logger
    graylog *gelfData  // Explicit field
}
```

Or use composition without embedding if you want to control interface.

### 3. Godoc Coverage
**File:** All

**Issue:** Missing function documentation.

**Recommendation:**
```go
// InGraylog configures UDP Graylog GELF output. Empty URL is ignored.
// Returns the options for chaining.
func (c *LoggerOptions) InGraylog(graylogURL, containerName string) *LoggerOptions
```

### 4. Package Naming
**File:** `http/chi`, `http/gin`, `http/mux`

**Issue:** Subpackages named after frameworks is fine, but ensure they export only what's needed.

---

## Features & Edge Cases 🚀

### 1. Missing: Async Logging
**Recommendation:** Add async mode with channel buffer:
```go
type AsyncOptions struct {
    BufferSize int
    Workers    int
}
```

### 2. Missing: Log Rotation
**Recommendation:** Add file handler with rotation support (max size, max files).

### 3. Missing: Sampling
**Recommendation:** Add sampling for high-throughput paths:
```go
func (c *LoggerOptions) WithSampler(sampler Sampler) *LoggerOptions
```

### 4. Missing: JSON vs Text Format
**Current:** Uses default slog JSON/text.

**Recommendation:** Add option to force JSON for Graylog compatibility:
```go
type Format string
const (
    FormatJSON Format = "json"
    FormatText Format = "text"
)
```

### 5. Edge Case: Concurrent Options Modification
**File:** `options.go`

**Issue:** `LoggerOptions` is mutable and not thread-safe.

**Recommendation:** Make options immutable after creation or use RWMutex:
```go
type LoggerOptions struct {
    mu        sync.RWMutex
    level     Level
    // ...
}
```

Or better: return new instance on each modification (current pattern is OK).

### 6. Edge Case: Graylog Connection Retry
**File:** `options.go`

**Recommendation:** Add retry logic with exponential backoff for Graylog connection.

---

## Testing 🧪

### 1. Missing Integration Tests
**Issue:** No tests with real Graylog server.

**Recommendation:** Add Docker Compose for test environment:
```yaml
version: '3'
services:
  graylog:
    image: graylog/graylog:latest
    ports:
      - "12201:12201/udp"
```

### 2. Missing Race Detection Tests
**Issue:** No `-race` flag usage in CI.

**Recommendation:** Add to CI:
```yaml
- name: Race detection
  run: go test -race ./...
```

### 3. Missing Fuzz Tests
**Issue:** No fuzzing for input validation.

**Recommendation:**
```go
func FuzzGraylogURL(f *testing.F) {
    f.Add("localhost:12201")
    f.Add("")
    f.Add("invalid")
    f.Fuzz(func(t *testing.T, url string) {
        opts := NewOptions().InGraylog(url, "test")
        _ = opts  // Should not panic
    })
}
```

### 4. Test Coverage Gaps
**Files:**
- `options_test.go` - Missing test for valid Graylog URL (integration)
- `logger_test.go` - Missing concurrent write tests
- `http/*_test.go` - Missing real request tests (use httptest)

### 5. Test Quality
**Issue:** Some tests just check "no panic" without verifying behavior.

**Recommendation:** Add assertions on output (capture stdout/stderr).

---

## Security 🔒

### 1. Format String Injection
**File:** `logger.go`

**Issue:** If user passes format strings as keys.

**Mitigation:** slog already handles this, but ensure no `fmt.Sprintf` with user input.

### 2. PII in Logs
**Issue:** No warning about PII in logs.

**Recommendation:** Add comment in README about GDPR/PII.

### 3. Graylog Authentication
**Issue:** No TLS or auth support for Graylog.

**Recommendation:** Add `InGraylogTLS(url, cert, key, containerName)` method.

---

## Documentation 📚

### 1. Missing: Migration Guide
**Recommendation:** Add MIGRATION.md for users moving from logrus.

### 2. Missing: Troubleshooting
**Recommendation:** Add TROUBLESHOOTING.md:
- "Graylog not receiving logs"
- "High CPU usage"
- "Memory leak"

### 3. Missing: Examples
**Recommendation:** Add `examples/` directory:
- `examples/basic/main.go`
- `examples/chi-middleware/main.go`
- `examples/graylog-setup/main.go`

---

## Summary

| Category | Count | Priority |
|----------|-------|----------|
| Critical | 3 | 🔴 High |
| Performance | 4 | 🟡 Medium |
| Style | 4 | 🟢 Low |
| Features | 6 | 🟡 Medium |
| Testing | 5 | 🟡 Medium |

**Immediate Actions:**
1. ✅ Fix `InGraylog` panic (already done)
2. Add nil checks for `inGraylog` writer
3. Add context fallback in `L()`
4. Add `-race` to CI
5. Add integration tests with Docker Compose

**Nice to Have:**
- Async logging mode
- TLS support for Graylog
- Fuzz testing
- Examples directory
