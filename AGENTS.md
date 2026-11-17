# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Project Overview

This is a Go utility library (`github.com/mzzsfy/go-util`) that provides high-performance, generic-based utility functions and components. The project has a strict philosophy of **zero third-party dependencies** (go.mod has no `require` entries) and heavily leverages Go generics (requires Go 1.18+).

## Architecture

The codebase is organized into modular packages, each serving a specific purpose:

### Core Packages

- **`seq`**: High-performance generic chain operation library inspired by Java Stream API
  - Supports parallel processing with configurable concurrency
  - Implements filtering, mapping, sorting, and aggregation operations
  - Zero allocations in hot paths where possible
  - Dual-track build: generic-method API (`Map[E]`/`GroupBy[K]`/`Cast[N]`) requires Go 1.27+ (`*_method_go127.go`); Go 1.18~1.26 build the `*_pre127.go` any-typed implementation instead. The two tracks are source-incompatible (same-named `Map`/`GroupBy` have different signatures; `CastAny` and the package-level `Join` exist only on the old track), so seq changes must build on both tracks (CI covers 1.18~1.27)
  - Key files: `seq.go`, `bi_seq.go`, `iterator.go`, `parallel.go`, `seq_pre127.go`, `seq_method_go127.go`, `util.go`

- **`di`**: High-performance dependency injection container
  - Complete lifecycle management (creation, destruction hooks)
  - Lazy loading with cycle detection
  - Configuration injection system
  - Hook system for before/after create/destroy
  - Performance statistics tracking
  - Key files: `container.go`, `container_config.go`, `container_creation.go`, `container_helpers.go`, `container_injection.go`, `container_instance.go`, `container_lifecycle.go`, `container_provide.go`, `container_type_conversion.go`, `container_utils.go`, `api.go`, `config_source.go`, `core.go`, `opt.go`, `package.go`, `doc.go`

- **`config`**: Configuration management system
  - Parser registry (`Parser` map; the `json` parser registers under the `json_parser` build tag)
  - Path-based value retrieval (`GetByPath`, `ValueFromPath`)
  - Nested map merge/tiling utilities (`MergeAndTilingMap`, `TilingMap`, `UntilingMap`)
  - Typed `Value` accessors with default-value variants (`Int`, `IntD`, `Float`, `Bool`, ...)
  - Key files: `config.go`, `item.go`, `parser_json.go`, `doc.go`

- **`concurrent`**: Concurrency utilities
  - `Int64Adder`: High-performance atomic counter (with 32-bit system fixes)
  - Reentrant locks (read/write)
  - Sliding window implementation
  - Queue implementations
  - Key files: `int64_adder.go`, `reentrant_lock.go`, `rw_lock.go`, `queue.go`, `queue_block.go`, `queue_ring.go`, `queue_segment.go`, `queue_chan.go`, `queue_delay.go`, `sliding_window.go`, `goid.go`, `cpu_cache_killer.go`

- **`storage`**: Advanced data structures
  - Multiple map implementations (Go native, Swiss, concurrent variants)
  - GLS (Goroutine Local Storage)
  - Performance-optimized storage primitives
  - Key files: `map.go`, `map_base.go`, `map_golang.go`, `map_array.go`, `map_swiss.go`, `map_concurrent_swiss.go`, `map_concurrent_wrapper.go`, `gls.go`, `cache.go`, `fastrand.go`, `doc.go`

- **`logger`**: Logging utilities
  - Configurable log levels and formats
  - Performance-oriented design
  - Integration with other components
  - Key files: `logger.go`, `event.go`, `formatter.go`, `format.go`, `level.go`, `hook.go`, `caller.go`, `global.go`, `manager.go`

- **`helper`**: General utilities
  - String and time processing
  - Cron job scheduler
  - Timer wheel scheduling
  - Call stack tracing and panic recovery
  - Function reflection utilities
  - Key files: `cron.go`, `cron_parse.go`, `cron_timer.go`, `timer_wheel.go`, `stack.go`, `try.go`, `reflect.go`, `string.go`, `bytes.go`, `time.go`, `time_local.go`, `async_writer.go`, `symbol.go`, `util_debounce.go`, `number.go`

- **`unsafe`**: Low-level operations
  - Goroutine ID retrieval
  - Runtime hash functions
  - Unsafe memory operations
  - Key files: `goid_s.go`, `goid_s_*.go` (per-version offsets, 32-bit variants included), `goid_stack.go`, `hasher.go`, `hasher_runtime.go`

- **`pool`**: Object reuse infrastructure
  - Goroutine pool with graceful shutdown and restart
  - Object pool (generic sync.Pool wrapper)
  - Byte pool and buffer pool
  - String pool with reference counting
  - Key files: `goroutine_pool.go`, `object_pool.go`, `bytes_pool.go`, `string.go`

- **`script`**: Embedded scripting language engine
  - Lexer, parser, compiler and virtual machine
  - Variable binding and external function calls
  - Full control flow (if/for/break/continue/return)
  - Dynamic typing with runtime type checking
  - Key files: `script.go`, `lexer.go`, `parser.go`, `compiler.go`, `runtime.go`

### Key Design Patterns

1. **Generic-First**: All utilities use Go generics extensively
2. **Zero Dependencies**: No third-party dependencies
3. **Performance-Oriented**: Focus on minimal allocations and optimal algorithms
4. **Test-Driven**: Comprehensive test suite with benchmarks
5. **Modular Design**: Each package is self-contained

## Development Commands

### Testing

> tests must add timeouts to prevent bugs from causing high CPU usage for a long time

```bash
# Run all tests
go test -timeout 30s ./...

# Run specific package tests
go test -timeout 30s -v ./seq
go test -timeout 30s -v ./di 

# Run specific test
go test -timeout 30s -v ./seq -run Test_1

# Run benchmarks
go test -timeout 120s -bench=. -benchmem ./seq

# Build-tag variants: these files are excluded from the default build, so default tests never compile them
go test -timeout 30s -tags json_parser ./config
go build -tags json_parser ./...
go build -tags concurrent_fast ./...
go build -tags concurrent_memory ./...
go build -tags nosimd ./...
```

### Building
```bash
# Build all packages
go build ./...

# Build specific package
go build ./seq
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...
# Known baseline: 3 `possible misuse of unsafe.Pointer` warnings from runtime hacks
# (goid_s.go, hasher_runtime.go, escape_test.go). Do not "fix" them to silence vet.
```

## Important Notes

### Go Version Requirements
- **Minimum**: Go 1.18 (for generics support, as declared in go.mod); local toolchain and CI latest track are Go 1.27
- **seq generic-method API**: two source-incompatible tracks — Go 1.27+ builds `*_method_go127.go` (generic methods `Map[E]`/`GroupBy[K]`/`Cast[N]`), Go 1.18~1.26 builds `*_pre127.go` (any-typed `Map`/`GroupBy`, `CastAny` and package-level `Join`). Verify seq changes on both tracks; CI covers 1.18~1.27

### Test Structure
- Tests are comprehensive and often include benchmarks
- Many packages have multiple test files for different scenarios
- Test files often include demo/example usage
- Some packages use `t.Parallel()`; adding it to tests that share global state can introduce flakiness

### Breaking Changes
The README explicitly states: **"本项目可能会有破坏性的修改函数签名行为,不要轻易升级"** (This project may have breaking changes to function signatures, do not upgrade casually)

### Performance Considerations
- Heavy use of generics can impact compilation speed for large interfaces
- Many utilities are optimized for specific use cases
- Benchmark tests are included for performance validation

### Common Development Patterns
- Use the task-list tool of the current agent environment (e.g. `todo_write`) for tracking multi-step tasks
- Prefer reading existing files over creating new ones
- Test changes thoroughly before committing
- Consider performance implications of changes

### Module Dependencies
- Module path: `github.com/mzzsfy/go-util`

## Testing Strategy

The project uses a comprehensive testing approach:
- Unit tests for all public APIs
- Integration tests for complex workflows
- Benchmark tests for performance validation
- Edge case testing for robustness
- Parallel execution testing for concurrency utilities

When working on this codebase, always ensure tests pass before making changes and consider the performance implications of any modifications.