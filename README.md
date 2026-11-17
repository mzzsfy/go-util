### go的一些工具类,大量运用了泛型,最低要求 golang 1.18

[![](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fmzzsfy%2Fgo-util&count_bg=%2379C83D&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://github.com/mzzsfy)

```bash
go get github.com/mzzsfy/go-util
```

项目遵守:

- **不引入任何第三方库**
- 全力使用泛型
- 尽量优化性能
- 提供优雅可读的函数签名

**本项目可能会有破坏性的修改函数签名行为,不要轻易升级**

### 简单说明

- [seq](./seq)  
  一个高性能的golang的泛型链式调用库,实现了类似java stream逻辑,串行链路不依赖chan和goroutine,
  并行能力基于固定worker池(goroutine+channel)实现,支持任意类型的链式调用,支持并行化(
  可限制并行数量),排序等

    ```go
    // 从19开始,每次减3,直到-10,额外添加-100,100,10,0,1,过滤出偶数,再丢弃前5个,从小到大排序,打印到控制台
    FromIntSeq(19, -10, -3).Add(-100,100,10,0,1).Filter(func(i int) bool {
        return i%2 == 0
    }).Drop(5).Sort(LessT[int]).ForEach(func(i int) {
        fmt.Println(i)
    })
    // a1b2c3(内置 print 不插分隔符;map 迭代顺序不固定)
    // "a,b,c"(顺序不固定)
    BiFromMap(map[string]int{"a": 1, "b": 2, "c": 3}).OnEach(func(k string, v int) {
        print(k, v)
    }).JoinStringBy(func(k string, v int) string {
        return k
    }, ",")
    ```
  更多例子见: [seq__test.go](./seq/seq__test.go) [bi_seq_test.go](./seq/bi_seq_test.go)(旧轨 go1.18~1.26) [seq_method_go127_test.go](./seq/seq_method_go127_test.go)(go1.27+)
  **如果使用seq处理大interface{},可能会导致编译速度下降,编译缓存磁盘占用大**

  > 泛型方法双轨: go1.27+ 提供强类型泛型方法(如 `Map[E]`/`GroupBy[K]`/`Cast[N]`);go1.18~1.26 的同名方法为
  > any 形态(`Map`/`GroupBy` 签名不同,`Cast` 需改用 `CastAny`,另有 `MapFlat`/`Reduce`/`GroupByFirst` 等差异,非穷举)。
  > 两轨源码不兼容,跨版本使用须按轨改写,详见 [seq/README.md](./seq/README.md)

  > 参考来源: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw

- [config](./config)  
  工作模式类似spring配置工具类

    ```go
    // 构造嵌套配置map
    cfg := map[string]any{
        "database": map[string]any{
            "host": "localhost",
            "port": 3306,
        },
    }
    // 通过路径读取嵌套值,支持自动类型转换和默认值
    host := config.GetByPath(cfg, "database.host")           // "localhost"
    port := config.ValueFromPath(cfg, "database.port").Int()  // 3306
    timeout := config.ValueFromPath(cfg, "database.timeout").IntD(5) // 5 (默认值)
    fmt.Println(host, port, timeout)
    ```
  更多例子见: [config/README.md](./config/README.md) [config_test.go](./config/config_test.go)

- [unsafe](./unsafe)  
  获取goroutine id,hash等不安全操作

    ```go
    // 获取当前 goroutine 的 id
    id := unsafe.GoID()
    println(id)
    ```
  更多例子见: [unsafe/README.md](./unsafe/README.md)

- [concurrent](./concurrent)  
  一些并发相关的工具,包含可重入锁,MPMC队列,滑动窗口限流,原子计数器等

    ```go
    // Int64Adder: 高并发原子计数器,类似Java LongAdder
    adder := &concurrent.Int64Adder{}
    adder.AddSimple(10) // 原子递增
    adder.AddSimple(20)
    val := adder.Sum() // 30
    println(val)

    // 滑动窗口限流: 1秒滑动窗口内最多允许100次请求
    sw, err := concurrent.NewSlidingWindow(1000, 100)
    if err != nil {
        panic(err) // 参数非法返回哨兵错误
    }
    if sw.CanDo() {
        // 允许执行
    } else {
        // 被限流,拒绝
    }
    ```
  更多例子见: [README.md](./concurrent/README.md) [int64_adder_test.go](./concurrent/int64_adder_test.go) [sliding_window_test.go](./concurrent/sliding_window_test.go)

- [pool](./pool)  
  一些池化工具:协程池,对象池

    ```go
    // 对象池:复用对象减少GC压力
    op := pool.NewObjectPool[bytes.Buffer](
        func() *bytes.Buffer { return &bytes.Buffer{} },
        func(b *bytes.Buffer) { b.Reset() }, // 归还时重置状态
    )
    buf := op.Get()     // 获取对象
    buf.WriteString("hello")
    op.Put(buf)         // 归还对象

    // 字符串池:用数字ID代替长字符串,适合做Map的Key
    sp := pool.NewStringPool()
    id := sp.Use("long-long-key-string")  // 分配ID,引用计数+1
    println(id)
    sp.UnUse("long-long-key-string")      // 释放引用,归零时自动清理
    ```
  更多例子见: [pool/README.md](./pool/README.md)

- [storage](./storage)  
  map等存储工具,swissMap,gls等

    ```go
    // 泛型Map:默认使用高性能Swiss Map实现
    m := storage.NewMap[string, int]() // 等价于 storage.NewMap(storage.MapTypeSwiss[string, int](16))
    m.Put("hello", 1)
    v, ok := m.Get("hello")  // 1, true
    println(v, ok)
    m.Iter(func(k string, v int) bool {
        println(k, v)
        return false
    })
    m.Delete("hello")

    // 并发安全Map
    cm := storage.NewMap(storage.MapTypeSwissConcurrent[string, int]())
    cm.Put("key", 1)

    // Goroutine Local Storage:类似Java ThreadLocal
    storage.KnowHowToUseGls() // 使用前必须声明已知用法
    var traceId = storage.NewGlsItem[string]() // 全局定义key
    func() {
        traceId.Set("abc-123")            // 在当前goroutine存储
        val, _ := traceId.Get()           // "abc-123"
        println(val)
        traceId.Delete(true)              // 使用完毕必须清理,参数true表示自动清理无其他key的goroutine
    }()
    ```
  更多例子见: [storage/README.md](./storage/README.md) [map_test.go](./storage/map_test.go) [gls_test.go](./storage/gls_test.go)

- [logger](./logger)
  日志工具,支持可配置日志级别与格式,面向性能设计

    ```go
    log := logger.New("app", logger.WithLevel(logger.DebugLevel)) // 默认级别 Info, 此处放开 Debug
    log.I("hello {}", "world")                                    // {} 占位
    log.D("debug {}", 1)                                          // 级别放开后才会输出
    log.Info().Str("k", "v").Msgf("count=%d", 1)                  // 事件式链式调用, fmt 格式
    ```
  更多例子见: [logger/README.md](./logger/README.md)

- [helper](./helper)
  一些工具,如: 字符串处理,时间日期处理,cron任务,延时任务 等

    ```go
    // 解析 cron 表达式,非法时返回错误
    scheduler, err := helper.ParseCron("0 * * * *")
    if err != nil {
        panic(err)
    }
    next := scheduler.NextTime(time.Now()) // 计算下一次触发时间
    fmt.Println(next)
    ```
  更多例子见: [helper/README.md](./helper/README.md)

- [di](./di)
  高性能依赖注入容器,支持完整的生命周期管理、配置注入和钩子系统

    ```go
    type Service struct{} // 业务类型

    c := di.New()
    // 提供构造函数,容器按参数类型自动解析依赖并缓存实例
    if err := di.Provide(c, func(c di.Container) (*Service, error) { return &Service{}, nil }); err != nil {
        panic(err)
    }
    // 获取实例(依赖未提供、循环依赖等返回错误)
    svc, err := di.Get[*Service](c)
    if err != nil {
        panic(err)
    }
    fmt.Println(svc)
    ```
  更多例子见: [di/README.md](./di/README.md)

- [script](./script)
  嵌入式脚本语言引擎,支持动态脚本执行、变量绑定、外部函数调用和完整控制流

    ```go
    // 单行脚本执行
    result, err := script.Eval("10 + 20")
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Int()) // 30

    // 带变量绑定执行
    result, _ = script.EvalWithBindings(`
        x :=>int getBindValue("x")
        y :=>int getBindValue("y")
        x + y
    `, map[string]any{"x": 10, "y": 20})

    // 编译一次,多次执行
    parser := script.NewParser()
    engine := script.NewEngine()
    ctx := script.NewContext()
    ctx.BindFunc("double", func(x int) int { return x * 2 })
    compiled, _ := parser.Compile(`
        #fn double(int)=>int
        double(21)
    `)
    result, _ = engine.Run(ctx, compiled) // 42
    ```
  更多例子见: [README.md](./script/README.md)

## License

[Apache-2.0](./LICENSE)
