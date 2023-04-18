一个高性能的golang的泛型链式调用库,实现了类似java stream逻辑,不依赖chan和goroutine,支持任意类型的链式调用,支持并行化(可限制并行数量),排序等

```go
// print 0,1,2,3,4,5,6,7,8,9
FromIntSeq().Take(10).ForEach(func(i int) {
  fmt.Println(i)
})

// 自定义生产者,生成无限长度随机序列,循环打印,过滤出偶数,丢弃前10个,然后取前5个,生成切片
From(func(f func(i int)) {
  for {
      f(rand.Int())
  }
}).OnEach(func(i int) {
  fmt.Println("", i)
}).Filter(func(i int) bool {
  return i%2 == 0
}).Drop(10).Take(5).ToSlice()

//结果 "10,9,8 ... 3,2,1"
FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Sort(func(i, j int) bool {
  return i > j
}).JoinString(strconv.Itoa,",")


// 远端下载多个文件,限制并发,限制顺序, 测试代码在 seq_download_test.go
BiCastAnyT(FromSlice(urls).MapBiSerialNumber(1).OnEach(func(i int, s string) {
    t.Logf("开始下载第%d个文件:%s\n", i, s)
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
    t.Logf("第%d个文件已写入\n", i)
}).
ForEach(func(i int, bytes []byte) {
    writer.Write(bytes)
})
```

更多例子见: [seq_test.go](./seq__test.go) [biSeq_test.go](./biSeq_test.go)

多元素,如map使用BiSeq

```go
// "a,b,c"
BiFromMap(map[string]int{"a": 1, "b": 2, "c": 3}).OnEach(func(k string, v int) {
    fmt.Println(k, v)
}).KSeq().JoinString(func(s string){retrun s}, ",")
```

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

> 参考了: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw
