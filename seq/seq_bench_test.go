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

func Benchmark_SeqRang(b *testing.B) {
    FromIntSeq(0, b.N-1).Complete()
}

func Benchmark_SeqTack(b *testing.B) {
    FromIntSeq().Take(b.N).Complete()
}

func Benchmark_R_For(b *testing.B) {
    for i := b.N; i > 0; i-- {
    }
}

func Benchmark_R_SeqEach(b *testing.B) {
    From(func(t func(int)) {
        for i := b.N; i > 0; i-- {
            t(i)
        }
    }).Complete()
}

func Benchmark_R_SeqRang(b *testing.B) {
    FromIntSeq(b.N-1, 0).Complete()
}

//内存消耗与原生一致,性能与原生调用性能差距已经不大了,吊打市面上其他类似库

//go test -v -run=Benchmark.+ -count 3 -bench=. -benchmem ./seq
//goos: windows
//goarch: amd64
//cpu: Intel(R) Core(TM) i5-8500 CPU
