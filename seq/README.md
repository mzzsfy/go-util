一个高性能的golang的泛型链式调用库,实现了类似java stream逻辑,不依赖chan和goroutine,支持任意类型的链式调用,支持并行化(可限制并行数量),排序等

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


// 远端下载多个文件,限制并发,限制顺序, 测试代码在 seq_download_test.go
BiCastAnyT(FromSlice(urls).MapBiSerialNumber(1).OnEach(func(i int, s string) {
    fmt.Printf("开始下载第%d个文件:%s\n", i, s)
}).MapVParallel(func(id int, s string) any {
    //下载文件,返回[]byte
    return DownloadFile(s)
},
//设置为按照顺序下载 1弱排序 2强排序
2,
//并发数
thread).
//强制转换
Map(AnyBiTK[int]), 0, []byte{}).
//写入下载文件名称
OnBefore(1, func(_ i, _ []byte) {
    name := string(FromSlice(fileName).Take(40).ToSlice())
    writer.Header().Set("Content-Type", "text/plain")
    writer.Header().Set("Content-Disposition", `attachment; filename="`+time.Now().Format("02_15:04:05")+"_"+name+`.merge.txt"`)
}).
OnLast(func(i *int, i2 *[]byte) {
    fmt.Printf("第%d个文件已写入\n", i)
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
	KSeq().JoinString(func(s string){retrun s}, ",")
```

更多例子见: [seq_test.go](./seq__test.go) [biSeq_test.go](./biSeq_test.go)

优势:

- 统一的命名规范,简单易用
- 基于函数回调,[超高性能](seq_bench_test.go)
- 懒加载,无消费不生产元素
- 生产者与消费者交替执行,不会造成相互阻塞
- 可[透明并发](seq_enhance_test.go),并控制并发数量,[对于异步io操作极其友好](seq_download_test.go)
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
    Drop(n int) Seq[T]
    Distinct(equals func(T, T) bool) Seq[T]
    MergeBiInt(iterator Iterator[int]) BiSeq[int, T]
    MergeBiIntRight(iterator Iterator[int]) BiSeq[T, int]
    MergeBiString(iterator Iterator[string]) BiSeq[string, T]
    MergeBiStringRight(iterator Iterator[string]) BiSeq[T, string]
    MergeBiAny(iterator Iterator[any]) BiSeq[any, T]
    MergeBiAnyRight(iterator Iterator[any]) BiSeq[any, T]
    MapBiSerialNumber(Range ...int) BiSeq[int, T]
    MapBiInt(f func(T) int) BiSeq[int, T]
    MapBiString(f func(T) string) BiSeq[string, T]
    MapBiAny(f func(T) any) BiSeq[any, T]
    MapBiAnyRight(f func(T) any) BiSeq[T, any]
    MapParallel(f func(T) any, order ...int) Seq[any]
    Map(f func(T) any) Seq[any]
    MapString(f func(T) string) Seq[string]
    MapInt(f func(T) int) Seq[int]
    MapFlat(f func(T) Seq[any]) Seq[any]
    MapFlatInt(f func(T) Seq[int]) Seq[int]
    MapFlatString(f func(T) Seq[string]) Seq[string]
    MapSliceN(n int) Seq[any]
    MapSliceBy(f func(T, []T) bool) Seq[any]
    Join(seqs ...Seq[T]) Seq[T]
    JoinF(seq Seq[any], cast func(any) T) Seq[T]
    Add(ts ...T) Seq[T]
    AddF(cast func(any) T, ts ...any) Seq[T]
    Complete()
    ForEach(f func(T))
    AsyncEach(f func(T))
    First() *T
    FirstOr(d T) T
    FirstOrF(d func() T) T
    Last() *T
    LastOr(d T) T
    LastOrF(d func() T) T
    AnyMatch(f func(T) bool) bool
    AllMatch(f func(T) bool) bool
    GroupBy(f func(T) any) map[any][]T
    GroupByFirst(f func(T) any) map[any]T
    GroupByLast(f func(T) any) map[any]T
    Reduce(f func(T, any) any, init any) any
    ToSlice() []T
    Count() int
    Count64() int64
    SumBy(f func(T) int) int
    SumBy64(f func(T) int64) int64
    SumByFloat32(f func(T) float32) float32
    SumByFloat64(f func(T) float64) float64
    JoinStringBy(f func(T) string, delimiter ...string) string
    JoinString(delimiter ...string) string
    Stoppable() Seq[T]
    Catch(f func(any)) Seq[T]
    CatchWithValue(f func(T, any)) Seq[T]
    OnEach(f func(T)) Seq[T]
    OnEachN(step int, f func(T), skip ...int) Seq[T]
    OnEachNX(step int, f func(T), skip ...int) Seq[T]
    OnBefore(i int, f func(T)) Seq[T]
    OnAfter(i int, f func(T)) Seq[T]
    OnFirst(f func(T)) Seq[T]
    OnLast(f func(*T)) Seq[T]
    Sync() Seq[T]
    Parallel(concurrent ...int) Seq[T]
    Sort(less func(T, T) bool) Seq[T]
    Cache() Seq[T]
    Repeat(n ...int) Seq[T]
}

interface BiSeq[K,V]{
    Filter(f func(K, V) bool) BiSeq[K, V]
    Take(n int) BiSeq[K, V]
    Drop(n int) BiSeq[K, V]
    Distinct(equals func(K, V, K, V) bool) BiSeq[K, V]
    DistinctK(equals func(K, K) bool) BiSeq[K, V]
    DistinctV(equals func(V, V) bool) BiSeq[K, V]
    SeqK() Seq[K]
    SeqV() Seq[V]
    SeqBy(f func(K, V) any) Seq[any]
    SeqKBy(f func(K, V) K) Seq[K]
    SeqVBy(f func(K, V) V) Seq[V]
    MapStringBy(f func(K, V) string) Seq[string]
    MapIntBy(f func(K, V) int) Seq[int]
    MapSliceN(n int) Seq[any]
    MapSliceBy(f func(K, V, any) bool) Seq[any]
    Stoppable() BiSeq[K, V]
    Catch(f func(any)) BiSeq[K, V]
    CatchWithValue(f func(K, V, any)) BiSeq[K, V]
    OnEach(f func(K, V)) BiSeq[K, V]
    OnEachAfter(f func(K, V)) BiSeq[K, V]
    OnEachN(step int, f func(k K, v V), skip ...int) BiSeq[K, V]
    OnEachNX(step int, f func(k K, v V), skip ...int) BiSeq[K, V]
    OnBefore(i int, f func(K, V)) BiSeq[K, V]
    OnAfter(i int, f func(K, V)) BiSeq[K, V]
    OnFirst(f func(K, V)) BiSeq[K, V]
    OnLast(f func(*K, *V)) BiSeq[K, V]
    Cache() BiSeq[K, V]
    Sync() BiSeq[K, V]
    Parallel(concurrent ...int) BiSeq[K, V]
    Sort(less func(K, V, K, V) bool) BiSeq[K, V]
    SortK(less func(K, K) bool) BiSeq[K, V]
    SortV(less func(V, V) bool) BiSeq[K, V]
    Repeat(n ...int) BiSeq[K, V]
    Complete()
    ForEach(f func(K, V))
    AsyncEach(f func(K, V))
    First() (*K, *V)
    FirstOr(k K, v V) (K, V)
    FirstOrF(f func() (K, V)) (K, V)
    Last() (*K, *V)
    LastOr(k K, v V) (K, V)
    LastOrF(f func() (K, V)) (K, V)
    AnyMatch(f func(K, V) bool) bool
    AllMatch(f func(K, V) bool) bool
    Keys() []K
    Values() []V
    Count() int
    Count64() int64
    SumBy(f func(K, V) int) int
    JoinStringBy(f func(K, V) string, delimiter ...string) string
    Reduce(f func(K, V, any) any, init any) any
    MapKParallel(f func(k K, v V) any, order ...int) BiSeq[any, V]
    MapVParallel(f func(k K, v V) any, order ...int) BiSeq[K, any]
    ExchangeKV() BiSeq[V, K]
    Map(f func(K, V) (any, any)) BiSeq[any, any]
    MapK(f func(K, V) any) BiSeq[any, V]
    MapKInt(f func(K, V) int) BiSeq[int, V]
    MapKString(f func(K, V) string) BiSeq[string, V]
    MapV(f func(K, V) any) BiSeq[K, any]
    MapVInt(f func(K, V) int) BiSeq[K, int]
    MapVString(f func(K, V) string) BiSeq[K, string]
    MapFlat(f func(K, V) BiSeq[any, any]) BiSeq[any, any]
    MapFlatK(f func(K, V) Seq[any]) BiSeq[any, V]
    MapFlatV(f func(K, V) Seq[any]) BiSeq[K, any]
    MapFlatVInt(f func(K, V) Seq[int]) BiSeq[K, int]
    MapFlatVString(f func(K, V) Seq[string]) BiSeq[K, string]
    Join(seqs ...BiSeq[K, V]) BiSeq[K, V]
    JoinBy(seq BiSeq[any, any], cast func(any, any) (K, V)) BiSeq[K, V]
    Add(k K, v V) BiSeq[K, V]
    AddTuple(vs ...BiTuple[K, V]) BiSeq[K, V]
    AddBy(cast func(any, any) (K, V), es ...any) BiSeq[K, V]
}
```
