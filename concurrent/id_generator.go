package concurrent

import (
    "sync/atomic"
    "time"
)

type IdGenerator interface {
    NextId() uint64
}

type atomIdGenerator struct {
    g uint64
}

func (a *atomIdGenerator) NextId() uint64 {
    return atomic.AddUint64(&a.g, 1)
}

type snowFlake struct {
    timestamp uint64
    sequence0 uint64 //奇数时间用
    sequence1 uint64 //偶数时间用
    workerId  uint64
    timeShift uint64
    lastId    uint64
}

func (s *snowFlake) NextId() uint64 {
    for {
        timestamp := uint64(time.Now().UnixMilli()) - s.timeShift
        //同一毫秒,高并发时大概率走这个分支
        if atomic.CompareAndSwapUint64(&s.timestamp, timestamp, timestamp) {
            if timestamp&1 == 0 {
                sequence := atomic.AddUint64(&s.sequence0, 1)
                if sequence < 0x100000 {
                    return (timestamp)<<23 | s.workerId<<16 | sequence
                }
            } else {
                sequence := atomic.AddUint64(&s.sequence1, 1)
                if sequence < 0x100000 {
                    return (timestamp)<<23 | s.workerId<<16 | sequence
                }
            }
        } else {
            if timestamp < s.timestamp {
                //时间回拨
                if timestamp < s.timestamp-1000 {
                    panic("time error")
                } else {
                    time.Sleep(time.Duration(s.timestamp-timestamp)*time.Millisecond - time.Millisecond)
                }
                continue
            }
            if atomic.CompareAndSwapUint64(&s.timestamp, s.timestamp, timestamp) {
                if timestamp&1 == 0 {
                    atomic.StoreUint64(&s.sequence0, 0)
                } else {
                    atomic.StoreUint64(&s.sequence1, 0)
                }
                return timestamp<<23 | s.workerId<<16
            }
        }
    }
}
