package concurrent

import (
    "runtime"
    "sync/atomic"
)

var slotNumber, modNumber = func() (int, int) {
    n := runtime.NumCPU() - 1
    if n <= 1 {
        n = 1
    }
    //复制于java.util.HashMap
    n |= n >> 1
    n |= n >> 2
    n |= n >> 4
    n |= n >> 8
    n |= n >> 16
    if n < 0 {
        n = 1
    } else {
        n = n + 1
    }
    return n, n - 1
}()

type int64AdderCeil struct {
    int64
    // 对齐字节,cpu缓存一般为128字节,默认设置为性价比最高的配置,如果你有其他需求,可以使用 tag:concurrent_128bit 或者 tag:concurrent_32bit 减少内存占用
    // 详细说明参考README.md
    // $ go test -bench=Benchmark_bit.+ ./concurrent
    // goos: windows
    // goarch: amd64
    // pkg: github.com/mzzsfy/go-util/concurrent
    // cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
    // Benchmark_bitInt64Adder_0Bit-6             61538             19522 ns/op
    // Benchmark_bitInt64Adder_8Bit-6             85107             13847 ns/op
    // Benchmark_bitInt64Adder_16Bit-6           111110              9855 ns/op
    // Benchmark_bitInt64Adder_24Bit-6           154597              8164 ns/op
    // Benchmark_bitInt64Adder_56Bit-6           239994              5650 ns/op
    // Benchmark_bitInt64Adder_120Bit-6          272733              5006 ns/op
    _ [cpuCacheKillerPaddingLength]byte
}

// Int64Adder 用于统计int64类型的数据
// 作用类似于java.util.concurrent.atomic.LongAdder,并参考了部分代码
type Int64Adder struct {
    init   int32
    base   int64
    values []int64AdderCeil
}

// Add 增加v,手动提供goid来提高性能
func (l *Int64Adder) Add(goid int64, v int64) {
    if l.addNoCompete(v) {
        return
    }
    l.addCompete(goid, v)
}

func (l *Int64Adder) addNoCompete(v int64) bool {
    // init 三态: 0=未初始化, 1=初始化中, 2=就绪
    if atomic.LoadInt32(&l.init) == 0 {
        old := atomic.LoadInt64(&l.base)
        //没有并发竞争的场景下,直接CAS
        if atomic.CompareAndSwapInt64(&l.base, old, old+v) {
            return true
        }
        if atomic.CompareAndSwapInt32(&l.init, 0, 1) {
            //无扩容功能,使用该工具场景,并不会特别需要节省内存
            ceils := make([]int64AdderCeil, slotNumber)
            for i := range ceils {
                ceils[i] = int64AdderCeil{int64: 0}
            }
            l.values = ceils
            // 必须先写 values 再发布: store 之前的写入对读到该值的 load 可见
            atomic.StoreInt32(&l.init, 2)
            return false
        }
    }
    if atomic.LoadInt32(&l.init) != 2 {
        //等待初始化完成, init==2 后 values 必然可见
        for i := 0; ; i++ {
            if atomic.LoadInt32(&l.init) == 2 {
                break
            }
            if i > 10 {
                runtime.Gosched()
            }
        }
    }
    return false
}

func (l *Int64Adder) addCompete(goid int64, v int64) {
    atomic.AddInt64(&l.values[int(goid)&modNumber].int64, v)
}

func (l *Int64Adder) Decrement(goid int64) {
    l.Add(goid, -1)
}

func (l *Int64Adder) Increment(goid int64) {
    l.Add(goid, 1)
}

func (l *Int64Adder) IncrementSimple() {
    if l.addNoCompete(1) {
        return
    }
    l.addCompete(GoID(), 1)
}

func (l *Int64Adder) DecrementSimple() {
    if l.addNoCompete(-1) {
        return
    }
    l.addCompete(GoID(), -1)
}

func (l *Int64Adder) AddSimple(v int64) {
    if l.addNoCompete(v) {
        return
    }
    l.addCompete(GoID(), v)
}

func (l *Int64Adder) Sum() int64 {
    r := atomic.LoadInt64(&l.base)
    if atomic.LoadInt32(&l.init) != 2 {
        return r
    }
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
    if atomic.LoadInt32(&l.init) != 2 {
        return
    }
    for i := range l.values {
        atomic.StoreInt64(&l.values[i].int64, 0)
    }
}
