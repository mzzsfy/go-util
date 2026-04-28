package logger_test

import (
    "io"
    "testing"

    "github.com/mzzsfy/go-util/logger"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// 统一 io.Discard, 纯粹测日志构建开销

func newZapLogger() *zap.Logger {
    core := zapcore.NewCore(
        zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(io.Discard),
        zap.NewAtomicLevelAt(zap.InfoLevel),
    )
    return zap.New(core)
}

func newMzzsfyLogger() *logger.Log {
    logger.SetDefaultWriterTarget(io.Discard)
    return logger.Logger("bench")
}

// --- 单线程 ---

func BenchmarkZap_NoArgs(b *testing.B) {
    l := newZapLogger()
    for i := 0; i < b.N; i++ {
        l.Info("simple message")
    }
}

func BenchmarkMzzsfy_NoArgs(b *testing.B) {
    l := newMzzsfyLogger()
    for i := 0; i < b.N; i++ {
        l.I("simple message")
    }
}

func BenchmarkZap_OneArg(b *testing.B) {
    l := newZapLogger()
    for i := 0; i < b.N; i++ {
        l.Info("value", zap.Int("i", i))
    }
}

func BenchmarkMzzsfy_OneArg(b *testing.B) {
    l := newMzzsfyLogger()
    for i := 0; i < b.N; i++ {
        l.I("value={}", i)
    }
}

func BenchmarkZap_ThreeArgs(b *testing.B) {
    l := newZapLogger()
    for i := 0; i < b.N; i++ {
        l.Info("request", zap.Int("id", i), zap.String("method", "GET"), zap.Int("status", 200))
    }
}

func BenchmarkMzzsfy_ThreeArgs(b *testing.B) {
    l := newMzzsfyLogger()
    for i := 0; i < b.N; i++ {
        l.I("request id={} method={} status={}", i, "GET", 200)
    }
}

func BenchmarkZap_FilteredOut(b *testing.B) {
    l := newZapLogger()
    for i := 0; i < b.N; i++ {
        l.Debug("filtered", zap.Int("i", i))
    }
}

func BenchmarkMzzsfy_FilteredOut(b *testing.B) {
    l := newMzzsfyLogger()
    for i := 0; i < b.N; i++ {
        l.D("filtered {}", i)
    }
}

func BenchmarkZap_With(b *testing.B) {
    l := newZapLogger().With(zap.String("svc", "api"), zap.Int("ver", 2))
    for i := 0; i < b.N; i++ {
        l.Info("request", zap.Int("id", i))
    }
}

func BenchmarkMzzsfy_With(b *testing.B) {
    l := newMzzsfyLogger().With("svc", "api", "ver", 2)
    for i := 0; i < b.N; i++ {
        l.I("request id={}", i)
    }
}

// --- 并行 ---

func BenchmarkZap_Parallel(b *testing.B) {
    l := newZapLogger()
    b.RunParallel(func(pb *testing.PB) {
        for i := 0; pb.Next(); i++ {
            l.Info("value", zap.Int("i", i))
        }
    })
}

func BenchmarkMzzsfy_Parallel(b *testing.B) {
    l := newMzzsfyLogger()
    b.RunParallel(func(pb *testing.PB) {
        for i := 0; pb.Next(); i++ {
            l.I("value={}", i)
        }
    })
}

func BenchmarkZap_ParallelFiltered(b *testing.B) {
    l := newZapLogger()
    b.RunParallel(func(pb *testing.PB) {
        for i := 0; pb.Next(); i++ {
            l.Debug("filtered", zap.Int("i", i))
        }
    })
}

func BenchmarkMzzsfy_ParallelFiltered(b *testing.B) {
    l := newMzzsfyLogger()
    b.RunParallel(func(pb *testing.PB) {
        for i := 0; pb.Next(); i++ {
            l.D("filtered {}", i)
        }
    })
}
