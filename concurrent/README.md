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
...
Benchmark1Int64Adder/Int64Adder_32-6        	  137929	      9034 ns/op
Benchmark1Int64Adder/Atomic_32-6            	   51117	     34734 ns/op
Benchmark1Int64Adder/Int64Adder_128-6       	  182433	      6676 ns/op
Benchmark1Int64Adder/Atomic_128-6           	   60130	     31813 ns/op

$ go test -tags=concurrent_fast -bench=Benchmark1.+ ./concurrent
...
Benchmark1Int64Adder/Int64Adder_32-6        	  144579	      7944 ns/op
Benchmark1Int64Adder/Atomic_32-6            	   35713	     34461 ns/op
Benchmark1Int64Adder/Int64Adder_128-6       	  214286	      5586 ns/op
Benchmark1Int64Adder/Atomic_128-6           	   38584	     31026 ns/op

$ go test -tags=concurrent_memory -bench=Benchmark1.+ ./concurrent
...
Benchmark1Int64Adder/Int64Adder_32-6        	   97975	     11300 ns/op
Benchmark1Int64Adder/Atomic_32-6            	   50070	     33834 ns/op
Benchmark1Int64Adder/Int64Adder_128-6       	  123711	      9588 ns/op
Benchmark1Int64Adder/Atomic_128-6           	   58536	     30516 ns/op
```
## 队列

一个简单的队列,目前有待优化,大部分功能可以用chan代替

```go
q:=BlockQueueWrapper(NewQueue[int]())

q.Enqueue(1)
q.Enqueue(2)
q.Enqueue(3)
q.Dequeue()
q.Dequeue()
q.Dequeue()
```