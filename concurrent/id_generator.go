package concurrent

import (
    "runtime"
    "strconv"
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

//cas实现的雪花算法,高并发时比原生雪花算法性能更好
type snowFlake struct {
    timestamp uint64
    _         [56]byte //优化伪共享
    sequence0 uint64
    sequence1 uint64
    _         [56]byte
    workerId  uint64
    timeShift uint64
    lastId    uint64

    seqBit uint64
    seqMax uint64
}

func (s *snowFlake) gen(time, sequence uint64) uint64 {
    return (time << 23) | (s.workerId << s.seqBit) | sequence
}

func (s *snowFlake) init() {
    if s.seqBit == 0 {
        s.seqBit = 16 //约为普通机器极限值,每秒最多能产生65w(65535*1000)个id
    }
    s.seqMax = 1 << s.seqBit
}

func (s *snowFlake) NextId() uint64 {
    for {
        timestamp := uint64(time.Now().UnixMilli()) - s.timeShift
        //同一毫秒,高并发时大概率走这个分支
        if atomic.CompareAndSwapUint64(&s.timestamp, timestamp, timestamp) {
            if timestamp&1 == 0 {
                sequence := atomic.AddUint64(&s.sequence0, 1)
                if sequence < s.seqMax {
                    return s.gen(timestamp, sequence)
                } else {
                    if s.seqMax == 0 {
                        s.init()
                    }
                    runtime.Gosched()
                }
            } else {
                sequence := atomic.AddUint64(&s.sequence1, 1)
                if sequence < s.seqMax {
                    return s.gen(timestamp, sequence)
                } else {
                    if s.seqMax == 0 {
                        s.init()
                    }
                    runtime.Gosched()
                }
            }
        } else {
            if timestamp < s.timestamp {
                timestamp = uint64(time.Now().UnixMilli()) - s.timeShift
                x := atomic.LoadUint64(&s.timestamp) - timestamp
                if x > 0 {
                    if x > 200 {
                        panic("时钟回退: " + strconv.Itoa(int(atomic.LoadUint64(&s.timestamp))) + "->" + strconv.Itoa(int(timestamp)))
                    } else if x > 10 {
                        time.Sleep(time.Millisecond*time.Duration(x) - time.Millisecond*10)
                    }
                    runtime.Gosched()
                }
                continue
            }
            if atomic.CompareAndSwapUint64(&s.timestamp, s.timestamp, timestamp) {
                if timestamp&1 == 0 {
                    atomic.StoreUint64(&s.sequence0, 1)
                    return s.gen(timestamp, 1)
                } else {
                    atomic.StoreUint64(&s.sequence1, 0)
                    return s.gen(timestamp, 0)
                }
            }
        }
    }
}
