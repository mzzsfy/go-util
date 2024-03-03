package helper

import (
    "math"
    "math/rand"
    "testing"
    "time"
)

func TestAddDelayTask(t *testing.T) {
    t.Parallel()
    interval := time.Millisecond * 10
    s := NewScheduler(interval)
    start := time.Now()
    maxTime := time.Millisecond
    for i := 0; i < 8; i++ {
        i := i
        duration := time.Microsecond * time.Duration(rand.Int63n(int64(math.Pow10(i))))
        if duration < 0 {
            duration = -duration
        }
        maxTime = duration
        s.AddDelayTask(duration, func() {
            if duration > time.Millisecond {
                if time.Since(start) < duration {
                    t.Error("Task ran too early", duration.String(), time.Since(start))
                }
                if time.Since(start) > duration+interval*10 {
                    t.Error("Task ran too late", duration.String(), time.Since(start))
                }
            }
            t.Log("task", i, time.Since(start).String())
        })
    }
    defer s.Stop()
    t.Log("maxTime", maxTime.String())
    time.Sleep(maxTime + time.Millisecond*10)
}

func TestAddIntervalTask(t *testing.T) {
    t.Parallel()
    s := NewScheduler()
    counter := 0
    s.AddIntervalTask(time.Millisecond*200, func() {
        counter++
        if counter >= 5 {
            s.Stop()
        }
    })

    time.Sleep(time.Second * 2)

    if counter != 5 {
        t.Errorf("Expected counter to be 5, got %d", counter)
    }
}

func BenchmarkScheduler_AddCronTask(b *testing.B) {
    s := NewScheduler(time.Millisecond * 10)
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            s.AddDelayTask(time.Millisecond, func() {})
            s.AddDelayTask(time.Millisecond*10, func() {})
            s.AddDelayTask(time.Millisecond*100, func() {})
            s.AddDelayTask(time.Second, func() {})
            s.AddDelayTask(time.Second*10, func() {})
            s.AddDelayTask(time.Minute, func() {})
            s.AddDelayTask(time.Minute*10, func() {})
            s.AddDelayTask(time.Hour, func() {})
        }
    })
    b.Cleanup(func() {
        if b.N > 1000000 {
            b.Log(s.log())
        }
        s.Stop()
    })
}
