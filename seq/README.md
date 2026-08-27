# seq

一个高性能的golang的泛型链式调用库,实现了类似java stream逻辑,不依赖chan和goroutine,支持任意类型的链式调用,支持并行化(可限制并行数量),排序等

> **版本要求**: Go 1.27+ 使用下述泛型方法 API (Map/GroupBy/Cast 等为强类型方法);Go 1.18~1.26 构建时自动回退到旧版 any
> 形态函数 (`*_pre127.go`),功能对应但类型信息丢失,如 `CastAny`/`MapString` 等。

```go
// print 0,1,2,3,4,5,6,7,8,9
FromIntSeq().Take(10).ForEach(func(i int) {
  fmt.t.Log(i)
})

// 自定义生产者,生成无限长度随机序列,循环打印,过滤出偶数,丢弃前10个,然后取前5个,生成切片
From(func(f func(i int)) {
  for {
      f(rand.Int())
  }
}).OnEach(func(i int) {
  println(i)
}).Filter(func(i int) bool {
  return i%2 == 0
}).Drop(10).Take(5).ToSlice()

//结果 "10,9,8 ... 3,2,1"
FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Sort(func(i, j int) bool {
  return i > j
}).JoinString(strconv.Itoa,",")


// 远端下载多个文件,限制并发,限制顺序 (go1.27 泛型方法版本)
FromSlice(urls).Enumerate(1).OnEach(func(i int, s string) {
    fmt.Printf("开始下载第%d个文件:%s\n", i, s)
}).
MapVParallel(func(id int, s string) []byte {
    //下载文件
    return DownloadFile(s)
},
//按照顺序消费 2 强排序
2,
//并发数
thread).
OnBefore(1, func(i int, v []byte) {
    name := string(FromSlice(fileName).Take(40).ToSlice())
    writer.Header().Set("Content-Type", "text/plain")
    writer.Header().Set("Content-Disposition", `attachment; filename="`+time.Now().Format("02_15:04:05")+"_"+name+`.merge.txt"`)
}).
OnLast(func(i *int, i2 *[]byte) {
    fmt.Printf("第%d个文件已写入\n", *i)
}).
ForEach(func(i int, bytes []byte) {
    writer.Write(bytes)
})
```

多元素,如map使用BiSeq

```go
// a 1 b 2 c 3
// "a,b,c"
BiFromMap(map[string]int{"a": 1, "b": 2, "c": 3}).OnEach(func(k string, v int) {
println(k, v)
}).
	//转换为单元素的Seq
	FromBiK().JoinString(func(s string){retrun s}, ",")
```

更多例子见: [seq_test.go](./seq__test.go) [bi_seq_test.go](./bi_seq_test.go)

优势:

- 统一的命名规范,简单易用
- 基于函数回调,[超高性能](seq_bench_test.go)
- 懒加载,无消费不生产元素
- 生产者与消费者交替执行,不会造成相互阻塞
- 可[透明并发](seq_enhance_test.go),并控制并发数量,[对于异步io操作极其友好](seq_demo_download_test.go)
- 链式调用,可读性强,执行顺序与声明一致,可严格保证执行顺序
- 无需关心Seq的长度,可单元素,可无限长度元素
- 双方可终止任务,不会造成资源浪费
- 双方都认为自己为消费者,开发难度极低
- 透明挂载回调,透明异步,透明并行,透明排序,透明去重,不影响后续操作

> 额外说明: 需要使用消费方法才能触发执行,如ForEach,ToSlice,JoinString等

> 参考来源: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw


签名表:  

```go
interface Seq[T]{
    Filter(f func(T) bool) Seq[T]
    Take(n int) Seq[T]
    TakeWhile(f func(T) bool) Seq[T]
    Limit(n int) Seq[T]
    Drop(n int) Seq[T]
    DropWhile(f func(T) bool) Seq[T]
    Skip(n int) Seq[T]
    Distinct(equals func(T, T) bool) Seq[T]
    DistinctCustomize(contains func(T) bool) Seq[T]
    MapParallel[R any](syncFn func(T) R, order ...int) Seq[R]
    MapParallelCustomize[R any](asyncFn func(T, func(R))) Seq[R]
    Map[E any](cast func(T) E) Seq[E]
    MapFlat[E any](f func(T) Seq[E]) Seq[E]
    JoinL[E any](s2 Seq[E], cast func(E) T) Seq[T]
    JoinBy[E, R any](s2 Seq[E], cast1 func(T) R, cast2 func(E) R) Seq[R]
    Cast[N any]() Seq[N]
    Join(seqs ...Seq[T]) Seq[T]
    Add(ts ...T) Seq[T]
    AddIf(condition bool, ts ...T) Seq[T]
    AddIfF(condition func(T) bool, ts ...T) Seq[T]
    ForEach(f func(T))
    FindFirstBy(f func(T) bool) *T
    First() *T
    FirstOrF(d func() T) T
    Last() *T
    LastOrF(d func() T) T
    AnyMatch(f func(T) bool) bool
    AllMatch(f func(T) bool) bool
    NonMatch(f func(T) bool) bool
    GroupBy[K comparable](f func(T) K) map[K][]T
    GroupByFirst[K comparable](f func(T) K) map[K]T
    GroupByLast[K comparable](f func(T) K) map[K]T
    Reduce[R any](f func(T, R) R, init R) R
    MergeBi[K any](iterator Iterator[K]) BiSeq[K, T]
    MergeBiR[V any](iterator Iterator[V]) BiSeq[T, V]
    MapBi[K any](f func(T) K) BiSeq[K, T]
    MapBiR[V any](f func(T) V) BiSeq[T, V]
    Enumerate(Range ...int) BiSeq[int, T]
    ToSlice() []T
    Count() int
    SumBy(f func(T) int) int
    SumByFloat64(f func(T) float64) float64
    JoinStringBy(f func(T) string, delimiter ...string) string
    JoinString(delimiter ...string) string
    RecoverErr(f func(any)) Seq[T]
    RecoverErrWithValue(f func(T, any)) Seq[T]
    Finally(f func()) Seq[T]
    OnEach(f func(T)) Seq[T]
    OnEachF(step func(T) bool, f func(T), skip ...int) Seq[T]
    OnEachN(step int, f func(T), skip ...int) Seq[T]
    OnEachNX(step int, f func(T), skip ...int) Seq[T]
    OnBefore(i int, f func(T)) Seq[T]
    OnAfter(i int, f func(T)) Seq[T]
    OnFirst(f func(T)) Seq[T]
    OnLast(f func(*T)) Seq[T]
    Sync() Seq[T]
    Parallel(concurrent ...int) Seq[T]
    ParallelCustomize(fn func(T, func())) Seq[T]
    Sort(less func(T, T) bool) Seq[T]
    SortCustomize(sort func([]T)) Seq[T]
    Reverse() Seq[T]
    Cache(init ...bool) Seq[T]
    Repeat(n ...int) Seq[T]
}

interface BiSeq[K,V]{
    Filter(f func(K, V) bool) BiSeq[K, V]
    Take(n int) BiSeq[K, V]
    Drop(n int) BiSeq[K, V]
    Distinct(equals func(K, V, K, V) bool) BiSeq[K, V]
    MapVParallel[R any](f func(k K, v V) R, order ...int) BiSeq[K, R]
    Map[A, B any](f func(K, V) (A, B)) BiSeq[A, B]
    MapK[R any](f func(K, V) R) BiSeq[R, V]
    MapV[R any](f func(K, V) R) BiSeq[K, R]
    MapFlat[A, B any](f func(K, V) BiSeq[A, B]) BiSeq[A, B]
    MapFlatK[R any](f func(K, V) Seq[R]) BiSeq[R, V]
    MapFlatV[R any](f func(K, V) Seq[R]) BiSeq[K, R]
    MapFlatSingle[E any](f func(K, V) Seq[E]) Seq[E]
    JoinL[K1, V1 any](s2 BiSeq[K1, V1], cast func(K1, V1) (K, V)) BiSeq[K, V]
    JoinBy[A, B, K2, V2 any](s2 BiSeq[K2, V2], cast1 func(K, V) (A, B), cast2 func(K2, V2) (A, B)) BiSeq[A, B]
    Cast[NK, NV any]() BiSeq[NK, NV]
    Join(seqs ...BiSeq[K, V]) BiSeq[K, V]
    Add(k K, v V) BiSeq[K, V]
    AddIf(condition bool, k K, v V) BiSeq[K, V]
    AddIfF(condition func(BiSeq[K, V]) bool, k K, v V) BiSeq[K, V]
    RecoverErr(f func(any)) BiSeq[K, V]
    RecoverErrWithValue(f func(K, V, any)) BiSeq[K, V]
    Finally(f func()) BiSeq[K, V]
    OnEach(f func(K, V)) BiSeq[K, V]
    OnEachNX(step int, f func(idx int, k K, v V), skip ...int) BiSeq[K, V]
    OnBefore(i int, f func(K, V)) BiSeq[K, V]
    OnAfter(i int, f func(K, V)) BiSeq[K, V]
    OnFirst(f func(K, V)) BiSeq[K, V]
    OnLast(f func(*K, *V)) BiSeq[K, V]
    Cache(init ...bool) BiSeq[K, V]
    Sync() BiSeq[K, V]
    Parallel(concurrent ...int) BiSeq[K, V]
    Sort(less func(K, V, K, V) bool) BiSeq[K, V]
    Reverse() BiSeq[K, V]
    Repeat(n ...int) BiSeq[K, V]
    ForEach(f func(K, V))
    First() (*K, *V)
    FirstOrF(f func() (K, V)) (K, V)
    Last() (*K, *V)
    LastOrF(f func() (K, V)) (K, V)
    AnyMatch(f func(K, V) bool) bool
    AllMatch(f func(K, V) bool) bool
    Count() int
    SumBy(f func(K, V) int) int
    SumByFloat64(f func(K, V) float64) float64
    JoinStringBy(f func(K, V) string, delimiter ...string) string
    Reduce[R any](f func(K, V, R) R, init R) R
}
```

> 以下包级函数仅 go1.27 以下轨道可用 (非穷举,详见 `*_pre127.go`);go1.27+ 使用同名语义的泛型方法:

```go
// 双轨通用(非方法形态,轨道无关)
MapSliceN[T any](t Seq[T], n int) Seq[any]
MapSliceBy[T any](t Seq[T], f func(T, []T) bool) Seq[any]

// 仅旧轨示例 → go1.27 泛型方法替代
CastAny(seq)          → seq.Cast[T]()
BiCastAny(seq)        → seq.Cast[K,V]()
BiMap(seq, cast)      → seq.Map(f)
MapString/MapInt      → seq.Map[E](f)
MergeBiInt/String/Any → seq.MergeBi(iterator)
MapBiSerialNumber     → seq.Enumerate(range...)
GroupBy(any键)        → seq.GroupBy[K comparable](f)
Reduce(any)           → seq.Reduce[R any](f, init)
```
