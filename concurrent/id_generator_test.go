package concurrent

import (
    "sync"
    "testing"
    "time"
)

func TestSnowFlake(t *testing.T) {
    s := snowFlake{}
    for i := 0; i < 10; i++ {
        t.Log(s.NextId())
    }
    time.Sleep(1 * time.Millisecond)
    for i := 0; i < 10; i++ {
        t.Log(s.NextId())
    }
    time.Sleep(10 * time.Millisecond)
    for i := 0; i < 10; i++ {
        t.Log(s.NextId())
    }
    time.Sleep(100 * time.Millisecond)
    for i := 0; i < 10; i++ {
        t.Log(s.NextId())
    }
    time.Sleep(1000 * time.Millisecond)
    for i := 0; i < 10; i++ {
        t.Log(s.NextId())
    }
}

type snowFlakeLock struct {
    timestamp uint64
    sequence  uint64
    workerId  uint64
    timeShift uint64
    lastId    uint64
    lock      sync.Mutex
}

func (s *snowFlakeLock) NextId() uint64 {
    s.lock.Lock()
    defer s.lock.Unlock()
    for {
        timestamp := uint64(time.Now().UnixMilli()) - s.timeShift
        if timestamp < s.timestamp {
            panic("time error")
        }
        if timestamp == s.timestamp {
            s.sequence++
            if s.sequence < 0x100000 {
                return (timestamp)<<23 | s.workerId<<16 | s.sequence
            }
        } else {
            s.timestamp = timestamp
            s.sequence = 0
        }
    }
}

func BenchmarkSnowFlake_NextId(b *testing.B) {
    s := snowFlake{}
    b.SetParallelism(128)
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            s.NextId()
        }
    })
}

func BenchmarkSnowFlake_lock_NextId(b *testing.B) {
    s := snowFlakeLock{}
    b.SetParallelism(128)
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            s.NextId()
        }
    })
}
