# seq

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
.OnBefore(1, func(_ i, _ []byte) {
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
	FromBiK().JoinString(func(s string){retrun s}, ",")
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
    TakeWhile(f func(T) bool) Seq[T]
    Limit(n int) Seq[T]
    Drop(n int) Seq[T]
    DropWhile(f func(T) bool) Seq[T]
    Skip(n int) Seq[T]
    Distinct(equals func(T, T) bool) Seq[T]
    DistinctCustomize(contains func(T) bool) Seq[T]
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
    MapParallel(syncFn func(T) any, order ...int) Seq[any]
    MapParallelCustomize(asyncFn func(T, func(any))) Seq[any]
    Map(f func(T) any) Seq[any]
    MapString(f func(T) string) Seq[string]
    MapInt(f func(T) int) Seq[int]
    MapFlat(f func(T) Seq[any]) Seq[any]
    MapFlatInt(f func(T) Seq[int]) Seq[int]
    MapFlatString(f func(T) Seq[string]) Seq[string]
    MapSliceN(n int) Seq[any]
    MapSliceBy(f func(T, []T) bool) Seq[any]
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
    GroupBy(f func(T) any) map[any][]T
    GroupByFirst(f func(T) any) map[any]T
    GroupByLast(f func(T) any) map[any]T
    Reduce(f func(T, any) any, init any) any
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
    MapVParallel(f func(k K, v V) any, order ...int) BiSeq[K, any]
    Map(f func(K, V) (any, any)) BiSeq[any, any]
    MapV(f func(K, V) any) BiSeq[K, any]
    MapFlat(f func(K, V) BiSeq[any, any]) BiSeq[any, any]
    Join(seqs ...BiSeq[K, V]) BiSeq[K, V]
    JoinBy(seq BiSeq[any, any], cast func(any, any) (K, V)) BiSeq[K, V]
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
    Reduce(f func(K, V, any) any, init any) any
}
```
