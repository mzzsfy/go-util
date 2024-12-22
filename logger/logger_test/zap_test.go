package logger_test

import (
    "github.com/mzzsfy/go-util/helper"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "os"
    "testing"
    "time"
)

func Benchmark_Zap(b *testing.B) {
    os.Remove("zap.log")
    time.Sleep(time.Second)
    file, _ := os.OpenFile("zap.log", os.O_CREATE, 0666)
    defer file.Close()
    c := zapcore.NewCore(
        zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(helper.NewAsyncWriter(file)),
        zap.NewAtomicLevelAt(zap.InfoLevel),
    )
    logger := zap.New(c)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("test", zap.Int("i", i))
    }
}

func Benchmark_Concurrent_Zap(b *testing.B) {
    os.Remove("zap.log")
    time.Sleep(time.Second)
    file, _ := os.OpenFile("zap.log", os.O_CREATE, 0666)
    defer file.Close()
    c := zapcore.NewCore(
        zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(helper.NewAsyncWriter(file)),
        zap.NewAtomicLevelAt(zap.InfoLevel),
    )
    logger := zap.New(c)
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for i := 0; pb.Next(); i++ {
            logger.Info("test", zap.Int("i", i))
        }
    })
}
