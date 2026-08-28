package seq

import (
	"math/rand"
	"sort"
	"testing"
)

//======性能基准:覆盖 Last/OnLast/OnEachNX/Parallel/MapParallel有序/Sort/Distinct 热路径======

// benchSlice 固定规模数据,避免基准内构造开销
var benchSlice = makeRangeSlice(1000)

func makeRangeSlice(n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = i
	}
	return r
}

func sinkInt(int) {}

func Benchmark_Last(b *testing.B) {
	s := FromSlice(benchSlice)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalBenchPtr = s.Last()
	}
}

func Benchmark_LastOrF(b *testing.B) {
	s := FromSlice(benchSlice)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalBenchVal = s.LastOrF(nil)
	}
}

func Benchmark_OnLast(b *testing.B) {
	s := FromSlice(benchSlice)
	f := func(*int) {}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.OnLast(f).ForEach(sinkInt)
	}
}

func Benchmark_OnEachNX(b *testing.B) {
	s := FromSlice(benchSlice)
	f := func(int) {}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.OnEachNX(16, f).ForEach(sinkInt)
	}
}

func Benchmark_Parallel(b *testing.B) {
	s := FromSlice(benchSlice)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Parallel(4).ForEach(sinkInt)
	}
}

func Benchmark_MapParallelOrdered(b *testing.B) {
	s := FromSlice(benchSlice)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.MapParallel(func(v int) any { return v }, 2, 4).ForEach(func(any) {})
	}
}

func Benchmark_Sort(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromSlice(benchSlice).Sort(func(a, b int) bool { return a < b }).ForEach(sinkInt)
	}
}

// benchShuffled 固定种子的乱序数据
var benchShuffled = func() []int {
	r := makeRangeSlice(1000)
	rnd := rand.New(rand.NewSource(42))
	rnd.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
	return r
}()

func Benchmark_SortShuffled(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := make([]int, len(benchShuffled))
		copy(s, benchShuffled)
		sortSlice(s, func(a, b int) bool { return a < b })
	}
}

// Benchmark_SortRefShuffled 参照组:sort.Slice排同一乱序数据
func Benchmark_SortRefShuffled(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := make([]int, len(benchShuffled))
		copy(s, benchShuffled)
		sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	}
}

func Benchmark_Distinct(b *testing.B) {
	//有重复元素的场景,去重后保留一半
	half := makeRangeSlice(len(benchSlice) / 2)
	s := FromSlice(benchSlice).Join(FromSlice(half))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Distinct().ForEach(sinkInt)
	}
}

var globalBenchPtr *int
var globalBenchVal int
