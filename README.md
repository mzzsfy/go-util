### go的一些工具类,大量运用了泛型,最低要求 golang 1.18

- [seq](./seq)  
  一个golang的链式调用库

```go
// 生成一个从19到-10步进-3的整数序列,取偶数,并循环打印
FromIntSeq(19, -10, -3).Filter(func(i int) bool {
    return i%2 == 0
}).ForEach(func(i int) {
    fmt.Println(i)
})
```

> 参考了: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw
