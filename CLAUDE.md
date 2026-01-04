# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go utility library (`github.com/mzzsfy/go-util`) that provides high-performance, generic-based utility functions and components. The project has a strict philosophy of **zero third-party dependencies** (except gopkg.in/yaml.v3 for config) and heavily leverages Go generics (requires Go 1.18+).

## Architecture

The codebase is organized into modular packages, each serving a specific purpose:

### Core Packages

- **`seq`**: High-performance generic chain operation library inspired by Java Stream API
  - Supports parallel processing with configurable concurrency
  - Implements filtering, mapping, sorting, and aggregation operations
  - Zero allocations in hot paths where possible
  - Key files: `seq__test.go`, `bi_seq_test.go`, `parallel_test.go`

- **`di`**: High-performance dependency injection container
  - Complete lifecycle management (creation, destruction hooks)
  - Lazy loading with cycle detection
  - Configuration injection system
  - Hook system for before/after create/destroy
  - Performance statistics tracking
  - Key files: `container.go`, `config_source.go`, `opt.go`

- **`config`**: Configuration management system
  - Spring-like configuration patterns
  - Path-based value retrieval
  - Multiple value types (string, any, nil)
  - Key files: `item.go`, `config_test.go`

- **`concurrent`**: Concurrency utilities
  - `Int64Adder`: High-performance atomic counter (with 32-bit system fixes)
  - Reentrant locks (read/write)
  - Sliding window implementation
  - ID generators
  - Queue implementations
  - Key files: `Int64_adder.go`, `reentrant_lock.go`, `sliding_window_test.go`

- **`storage`**: Advanced data structures
  - Multiple map implementations (Go native, Swiss, concurrent variants)
  - GLS (Goroutine Local Storage)
  - Performance-optimized storage primitives
  - Key files: `map_test.go`, `gls_test.go`, `map_concurrent_swiss_test.go`

- **`logger`**: Logging utilities
  - Configurable log levels and formats
  - Performance-oriented design
  - Integration with other components
  - Key files: `logger_test.go`, `logger_test/*.go`

- **`helper`**: General utilities
  - String and time processing
  - Cron job scheduler
  - Delayed task execution
  - Bloom filters
  - Function reflection utilities
  - Key files: `cron_test.go`, `scheduler_test.go`

- **`unsafe`**: Low-level operations
  - Goroutine ID retrieval
  - Runtime hash functions
  - Unsafe memory operations
  - Key files: `goid_test.go`, `hasher_runtime_test.go`

- **`cmd`**: Command-line utilities (if present)

### Key Design Patterns

1. **Generic-First**: All utilities use Go generics extensively
2. **Zero Dependencies**: Only essential dependencies (yaml.v3)
3. **Performance-Oriented**: Focus on minimal allocations and optimal algorithms
4. **Test-Driven**: Comprehensive test suite with benchmarks
5. **Modular Design**: Each package is self-contained

## Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test -v ./seq
go test -v ./di

# Run specific test
go test -v ./seq -run Test_1

# Run benchmarks
go test -bench=. -benchmem ./seq
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
```

## Important Notes

### Go Version Requirements
- **Minimum**: Go 1.18 (for generics support)
- **Current**: Go 1.25.0 (in development environment)

### Test Structure
- Tests are comprehensive and often include benchmarks
- Many packages have multiple test files for different scenarios
- Test files often include demo/example usage
- Parallel testing is used extensively (`t.Parallel()`)

### Breaking Changes
The README explicitly states: **"本项目可能会有破坏性的修改函数签名行为,不要轻易升级"** (This project may have breaking changes to function signatures, do not upgrade casually)

### Performance Considerations
- Heavy use of generics can impact compilation speed for large interfaces
- Many utilities are optimized for specific use cases
- Benchmark tests are included for performance validation

### Common Development Patterns
- Use `TodoWrite` tool for tracking multi-step tasks
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