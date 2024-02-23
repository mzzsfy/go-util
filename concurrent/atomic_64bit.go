//go:build !concurrent_fast && !concurrent_memory

package concurrent

type c struct {
    int64
    //对齐字节,cpu缓存一般为128字节,在测试用这样的设置性价比最高,如果你需要更大的并发量,可以使用 tag:concurrent_128bit 或者 tag:concurrent_32bit 减少内存占用
    //$ go test -bench=Benchmark_bit.+ ./concurrent
    //goos: windows
    //goarch: amd64
    //pkg: github.com/mzzsfy/go-util/concurrent
    //cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
    //Benchmark_bitInt64Adder_0Bit-6             61538             19522 ns/op
    //Benchmark_bitInt64Adder_8Bit-6             85107             13847 ns/op
    //Benchmark_bitInt64Adder_16Bit-6           111110              9855 ns/op
    //Benchmark_bitInt64Adder_24Bit-6           154597              8164 ns/op
    //Benchmark_bitInt64Adder_56Bit-6           239994              5650 ns/op
    //Benchmark_bitInt64Adder_120Bit-6          272733              5006 ns/op
    _ [56]byte
}
