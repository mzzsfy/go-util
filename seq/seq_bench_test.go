package seq

import "testing"

func Benchmark_For(b *testing.B) {
    for i := 0; i < b.N; i++ {
    }
}

func Benchmark_SeqEach(b *testing.B) {
    From(func(t func(int)) {
        for i := 0; i < b.N; i++ {
            t(i)
        }
    }).Complete()
}

//与原生调用性能差距已经不大了,吊打市面上其他类似库
