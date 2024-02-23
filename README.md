### go的一些工具类,大量运用了泛型,最低要求 golang 1.18

[![](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fmzzsfy%2Fgo-util&count_bg=%2379C83D&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://github.com/mzzsfy)

项目遵守:

- **不引入任何第三方库**
- 全力使用泛型
- 尽量优化性能
- 提供优雅可读的函数签名

**本项目可能会有破坏性的修改函数签名行为,不要轻易升级**

### 简单说明

- [seq](./seq)  
  一个高性能的golang的泛型链式调用库,实现了类似java stream逻辑,不依赖chan和goroutine,支持任意类型的链式调用,支持并行化(
  可限制并行数量),排序等

    ```go
    // 从19开始,每次减3,直到-10,额外添加100,10,0,1,过滤出偶数,再丢弃前5个,从小到大排序,打印到控制台
    FromIntSeq(19, -10, -3).Add(-100,100,10,0,1).Filter(func(i int) bool {
        return i%2 == 0
    }).Drop(5).Order(LessT[int]).ForEach(func(i int) {
        fmt.t.Log(i)
    })
    // a 1 b 2 c 3
	// "a,b,c"
    BiFromMap(map[string]int{"a": 1, "b": 2, "c": 3}).OnEach(func(k string, v int) {
        print(k, v)
    }).KSeq().JoinString(func(s string){retrun s}, ",")
    ```
  更多例子见: [seq_test.go](./seq/seq__test.go) [biSeq_test.go](./seq/biSeq_test.go)

  > 参考来源: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw

- [config](./config)  
  工作模式类似spring配置工具类

- [unsafe](./unsafe)  
  获取goroutine id,hash等不安全操作

- [concurrent](./concurrent)  
  一些并发相关的工具,包含可重入锁等

- [pool](./pool)  
  一些池化工具:携程池,对象池

- [queue](./queue)  
  简单的队列

- [storage](./storage)  
  map等存储工具,swissMap,gls等

- [logger](./logger)  
  日志工具

- [helper](./helper)  
  一些简便方法(工具函数)
