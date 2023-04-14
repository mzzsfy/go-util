
一个golang的链式调用库

  ```go
  // print 0,1,2,3,4,5,6,7,8,9
  FromIntSeq().Take(10).forEach(func(i int) {
      fmt.Println(i)
  })
  
  // 生成一个无限长度随机序列,循环打印,过滤出偶数,丢弃前10个,然后取前5个,生成切片
  From(func(f func(i int)) {
      for {
          f(rand.Int())
      }
  }).OnEach(func(i int){
      fmt.Println("",i)
  }).Filter(func(i int) bool {
      return i%2 == 0
  }).Rrop(10).Take(5).ToSlice()
  
  //结果 "10,9,8 ... 3,2,1"
  FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Sort(func(i, j int) bool {
      return i < j
  }).JoinString(strconv.Itoa,",")
  ```

多元素,如map使用BiSeq

```go
// "a,b,c"
BiFromMap(map[string]int{"a": 1, "b": 2, "c": 3}).OnEach(func(k string, v int) {
    fmt.Println(k, v)
}).KSeq().JoinString(func(s string){retrun s}, ",")
```

>  参考了: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw
  
