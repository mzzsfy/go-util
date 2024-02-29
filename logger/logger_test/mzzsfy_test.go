package logger_test

import (
    "github.com/mzzsfy/go-util/helper"
    "github.com/mzzsfy/go-util/logger"
    "math/rand"
    "os"
    "testing"
    "time"
)

//$ go test -v -bench=Benchmark_Mzzsfy -benchmem -memprofile=mem.pprof -cpuprofile=cpu.pprof
//$ go test -v -bench=.+ -benchmem
func Benchmark_Mzzsfy(b *testing.B) {
    os.Remove("mzzsfy.log")
    time.Sleep(time.Second)
    file, _ := os.OpenFile("mzzsfy.log", os.O_CREATE, 0666)
    defer file.Close()
    logger.SetDefaultWriterTarget(helper.NewAsyncWriter(file))
    time.Sleep(time.Second)
    b.ResetTimer()
    log := logger.Logger("test")
    for i := 0; i < b.N; i++ {
        log.I("test", i)
    }
}

func Benchmark_Concurrent_Mzzsfy(b *testing.B) {
    os.Remove("mzzsfy.log")
    time.Sleep(time.Second)
    file, _ := os.OpenFile("mzzsfy.log", os.O_CREATE, 0666)
    defer file.Close()
    logger.SetDefaultWriterTarget(helper.NewAsyncWriter(file))
    time.Sleep(time.Second)
    log := logger.Logger("test")
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            log.I("test", rand.Int())
        }
    })
}
