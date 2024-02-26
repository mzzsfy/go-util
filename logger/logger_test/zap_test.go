package logger_test

import (
    "github.com/mzzsfy/go-util/helper"
    "go.uber.org/zap"
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
    wg := helper.NewWaitGroup(10)
    b.ResetTimer()
    for i := 0; i < 10; i++ {
        go func() {
            defer wg.Done()
            for i := 0; i < b.N; i++ {
                logger.Info("test", zap.Any("i", i))
            }
        }()
    }
    wg.Wait()
}
