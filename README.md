### go的一些工具类,大量运用了泛型,最低要求 golang 1.18

- [seq](./seq)  
  一个golang的链式调用库

```go
FromIntSeq(19, -10, -3).Add(-100,100,10).Filter(func(i int) bool {
    return i%2 == 0
}).Drop(5).Order(LessT[int]).ForEach(func(i int) {
    fmt.Println(i)
})
```
更多例子见: [seq_test.go](./seq/seq__test.go) [biSeq_test.go](./seq/biSeq_test.go)

> 参考了: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw
