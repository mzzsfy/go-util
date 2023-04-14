
- seq  
  一个golang的链式调用库

```go
// 生成一个从19到-10步进-2的整数序列,取偶数,并循环打印
FromIntSeq(19, -10, -3).Filter(func(i int) bool {
    return i%2 == 0
}).forEach(func(i int) {
    fmt.Println(i)
})
```

> 参考了: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw
