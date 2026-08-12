package bench_compare

import (
    "io"
    "testing"

    "github.com/mzzsfy/go-util/logger"
    "github.com/rs/zerolog"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// --- 初始化各库 Logger ---

func newOurs() *logger.Log {
    return logger.New("bench", logger.WithLevel(logger.InfoLevel), logger.WithWriter(io.Discard))
}

func newZerolog() zerolog.Logger {
    return zerolog.New(io.Discard).Level(zerolog.InfoLevel).With().Logger()
}

func newZap() *zap.Logger {
    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(io.Discard),
        zapcore.InfoLevel,
    )
    return zap.New(core)
}

// =============================
// 单线程: 无字段
// =============================

func Benchmark_NoFields_OursChained(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info().Msg("hello world")
    }
}

func Benchmark_NoFields_OursLegacy(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.I("hello world")
    }
}

func Benchmark_NoFields_Zerolog(b *testing.B) {
    l := newZerolog()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info().Msg("hello world")
    }
}

func Benchmark_NoFields_Zap(b *testing.B) {
    l := newZap()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info("hello world")
    }
}

// =============================
// 单线程: 3字段
// 3种API: 新kv(链式类型安全) / 旧kv(Any装箱) / 占位符(I)
// =============================

func Benchmark_3Fields_OursChained(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info().Str("user", "moke").Int("id", 42).Bool("ok", true).Msg("login")
    }
}

func Benchmark_3Fields_OursAnyKv(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info().Any("user", "moke").Any("id", 42).Any("ok", true).Msg("login")
    }
}

func Benchmark_3Fields_OursLegacy(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.I("user={} id={%d} ok={%v}", "moke", 42, true)
    }
}

func Benchmark_3Fields_Zerolog(b *testing.B) {
    l := newZerolog()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info().Str("user", "moke").Int("id", 42).Bool("ok", true).Msg("login")
    }
}

func Benchmark_3Fields_Zap(b *testing.B) {
    l := newZap()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info("login", zap.String("user", "moke"), zap.Int("id", 42), zap.Bool("ok", true))
    }
}

// =============================
// 单线程: 级别过滤 (不输出)
// =============================

func Benchmark_Filtered_OursChained(b *testing.B) {
    l := logger.New("bench", logger.WithLevel(logger.WarnLevel), logger.WithWriter(io.Discard))
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info().Str("k", "v").Msg("filtered")
    }
}

func Benchmark_Filtered_OursLegacy(b *testing.B) {
    l := logger.New("bench", logger.WithLevel(logger.WarnLevel), logger.WithWriter(io.Discard))
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.I("filtered {}", "v")
    }
}

func Benchmark_Filtered_Zerolog(b *testing.B) {
    l := zerolog.New(io.Discard).Level(zerolog.WarnLevel).With().Logger()
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info().Str("k", "v").Msg("filtered")
    }
}

func Benchmark_Filtered_Zap(b *testing.B) {
    l := zap.New(zapcore.NewCore(
        zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(io.Discard),
        zapcore.WarnLevel,
    ))
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l.Info("filtered", zap.String("k", "v"))
    }
}

// =============================
// 并发: 无字段
// =============================

func Benchmark_Parallel_NoFields_OursChained(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.Info().Msg("hello world")
        }
    })
}

func Benchmark_Parallel_NoFields_OursLegacy(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.I("hello world")
        }
    })
}

func Benchmark_Parallel_NoFields_Zerolog(b *testing.B) {
    l := newZerolog()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.Info().Msg("hello world")
        }
    })
}

func Benchmark_Parallel_NoFields_Zap(b *testing.B) {
    l := newZap()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.Info("hello world")
        }
    })
}

// =============================
// 并发: 3字段
// =============================

func Benchmark_Parallel_3Fields_OursChained(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.Info().Str("user", "moke").Int("id", 42).Bool("ok", true).Msg("login")
        }
    })
}

func Benchmark_Parallel_3Fields_OursAnyKv(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.Info().Any("user", "moke").Any("id", 42).Any("ok", true).Msg("login")
        }
    })
}

func Benchmark_Parallel_3Fields_OursLegacy(b *testing.B) {
    l := newOurs()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.I("user={} id={%d} ok={%v}", "moke", 42, true)
        }
    })
}

func Benchmark_Parallel_3Fields_Zerolog(b *testing.B) {
    l := newZerolog()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.Info().Str("user", "moke").Int("id", 42).Bool("ok", true).Msg("login")
        }
    })
}

func Benchmark_Parallel_3Fields_Zap(b *testing.B) {
    l := newZap()
    b.ReportAllocs()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            l.Info("login", zap.String("user", "moke"), zap.Int("id", 42), zap.Bool("ok", true))
        }
    })
}
