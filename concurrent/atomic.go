package concurrent

import (
    "github.com/mzzsfy/go-util/helper"
    "runtime"
    "sync/atomic"
)

var slotNumber, modNumber = func() (int, int) {
    n := runtime.NumCPU() - 1
    n |= n >> 1
    n |= n >> 2
    n |= n >> 4
    n |= n >> 8
    n |= n >> 16
    //复制于java.util.HashMap
    n = helper.Ternary(n < 0, 1, n+1)
    return n, n - 1
}()

// Int64Adder 用于统计int64类型的数据
// 作用类似于java.util.concurrent.atomic.LongAdder,并参考了部分代码
type Int64Adder struct {
    init   int32
    base   int64
    values []c
}

// Add 增加v,手动提供goid来提高性能
func (l *Int64Adder) Add(goid int64, v int64) {
    if l.init == 0 {
        old := l.base
        //没有并发竞争的场景下,直接CAS
        if atomic.CompareAndSwapInt64(&l.base, old, old+v) {
            return
        }
        if atomic.CompareAndSwapInt32(&l.init, 0, 1) {
            l.values = make([]c, slotNumber)
        }
    }
    if len(l.values) == 0 {
        //等待初始化
        for {
            if len(l.values) > 0 {
                break
            }
            runtime.Gosched()
        }
    }
    //不扩容,使用该工具场景时,并不会特别需要节省内存
    atomic.AddInt64(&l.values[int(goid)&modNumber].int64, v)
}

func (l *Int64Adder) Decrement(goid int64) {
    l.Add(goid, -1)
}

func (l *Int64Adder) Increment(goid int64) {
    l.Add(goid, 1)
}

func (l *Int64Adder) IncrementSimple() {
    l.Add(GoID(), 1)
}

func (l *Int64Adder) DecrementSimple() {
    l.Add(GoID(), -1)
}

func (l *Int64Adder) AddSimple(v int64) {
    l.Add(GoID(), v)
}

func (l *Int64Adder) Sum() int64 {
    r := l.base
    for i := range l.values {
        r += atomic.LoadInt64(&l.values[i].int64)
    }
    return r
}

func (l *Int64Adder) SumInt() int {
    return int(l.Sum())
}

func (l *Int64Adder) Reset() {
    l.base = 0
    if len(l.values) == 0 {
        return
    }
    for i := range l.values {
        l.values[i].int64 = 0
    }
}
