package logger_test

import (
    "go.uber.org/zap"
    "math/rand"
    "os"
    "testing"
    "time"
)

func Benchmark_Zap(b *testing.B) {
    os.Remove("zap.log")
    time.Sleep(time.Second)
    b.ResetTimer()
    logger, _ := zap.Config{
        Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
        Development:      false,
        Encoding:         "json",
        EncoderConfig:    zap.NewProductionEncoderConfig(),
        OutputPaths:      []string{"zap.log"},
        ErrorOutputPaths: []string{"zap.log"},
    }.Build(zap.WithCaller(false))
    for i := 0; i < b.N; i++ {
        logger.Info("test", zap.Any("i", i))
    }
}

func Benchmark_Concurrent_Zap(b *testing.B) {
    os.Remove("zap.log")
    time.Sleep(time.Second)
    logger, _ := zap.Config{
        Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
        Development:      false,
        Encoding:         "json",
        EncoderConfig:    zap.NewProductionEncoderConfig(),
        OutputPaths:      []string{"zap.log"},
        ErrorOutputPaths: []string{"zap.log"},
    }.Build(zap.WithCaller(false))
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            logger.Info("test", zap.Any("i", rand.Int()))
        }
    })
}
