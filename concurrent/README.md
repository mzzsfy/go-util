# 锁工具

## 可重入锁

依赖goid,可修改GoID为自定义的获取goroutine id的函数

```go
lock:=NewReentrantLock()
lock.Lock()
lock.Lock()
lock.Unlock()
lock.Unlock()
```

## Int64Adder

类似java的LongAdder,在高并发场景下可以少量提升atomic.AddInt64的性能,不建议滥用,内存消耗比atomic.AddInt64大
> 备注: cpu缓存一般为128字节,使用条件编译 tags=concurrent_fast 提高性能,或者 tags=concurrent_memory 减少内存占用  
> 原理是改变struct大小,缓解伪共享程度不同.运行`go test -bench=Benchmark_bit.+ ./concurrent`查看内存布局对性能影响

```go
adder := &Int64Adder{}
adder.AddSimple(10)
adder.AddSimple(20)
```

以下为测试结果

```shell
$ go test -bench=Benchmark1.+ ./concurrent
goos: windows
goarch: amd64                               
pkg: github.com/mzzsfy/go-util/concurrent   
cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
Benchmark1Int64Adder32-6          107142             12951 ns/op
Benchmark1Atomic32-6               54544             23898 ns/op
Benchmark1Int64Adder128-6         165574              7406 ns/op
Benchmark1Atomic128-6              60301             20293 ns/op

$ go test -tags=concurrent_fast -bench=Benchmark1.+ ./concurrent
...
Benchmark1Int64Adder32-6          117646             12306 ns/op
Benchmark1Atomic32-6               52669             24060 ns/op
Benchmark1Int64Adder128-6         190480              6501 ns/op
Benchmark1Atomic128-6              60460             19921 ns/op

$ go test -tags=concurrent_memory -bench=Benchmark1.+ ./concurrent
...
Benchmark1Int64Adder32-6           85508             15680 ns/op
Benchmark1Atomic32-6               55555             23125 ns/op
Benchmark1Int64Adder128-6         115384             10757 ns/op
Benchmark1Atomic128-6              59254             19733 ns/op
```