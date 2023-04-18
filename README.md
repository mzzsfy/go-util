### go的一些工具类,大量运用了泛型,最低要求 golang 1.18

- [seq](./seq)  
    一个高性能的golang的泛型链式调用库,实现了类似java stream逻辑,不依赖chan和goroutine,支持任意类型的链式调用,支持并行化(可限制并行数量),排序等

    ```go
    FromIntSeq(19, -10, -3).Add(-100,100,10).Filter(func(i int) bool {
        return i%2 == 0
    }).Drop(5).Order(LessT[int]).ForEach(func(i int) {
        fmt.Println(i)
    })
    ```
    更多例子见: [seq_test.go](./seq/seq__test.go) [biSeq_test.go](./seq/biSeq_test.go)
    
    > 参考: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw
