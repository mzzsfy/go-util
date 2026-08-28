# Changelog

本项目所有显著变更记录于此。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

注意:本项目可能会有破坏性的修改函数签名行为,不要轻易升级。

## [Unreleased]

### Breaking

以下变更不留兼容层,升级时需同步修改调用方:

- 删除 helper `OneOf` 系列:`OneOfL`/`OneOfR`/`OneOf3L`/`OneOf3M`/`OneOf3R`(util_select.go 整体移除)
- 删除 seq 恒等装箱 API:`AnyT`/`AnyBiT*`/`EqualsT`
- 删除 `DefaultParallelFunc`
- 重命名:`NewGopool` → `NewGoPool`;`WithMaxWorks` → `WithMaxWorkers`(字段同步更名为 `maxWorkers`)
- `NewParallel` 签名变更:返回 `(Parallel, error)`,新增错误哨兵 `ErrConcurrent`
- `NewSlidingWindow` 签名变更:返回 `(*SlidingWindow, error)`,新增错误哨兵 `ErrWindowNumber`/`ErrWindowTime`/`ErrAllowNumber`,以 `errors.Is` 判定
- `Distinct` 无参化:Seq 删除静默无效的 equals 参数;BiSeq 去重改为 `==` 比较语义,元素类型不可比较时 panic;需自定义 key 的场景改用新增的 `BiDistinctByKey`
- `ParseConfigs2Map` 合并方向反转:高 Order 配置覆盖低 Order(原方向相反)

### Fixed

- di:Shutdown 死锁;销毁顺序违背依赖承诺;并发首次加载误报;Transient 实例泄漏;ChildScope 竞态;半初始化实例可见
- script:常量池哈希碰撞;字节码槽越界;状态保存错乱;移位运算异常
- logger:go1.27 下 caller 定位错误;32 位平台 caller 指针宽度错误
- 32 位平台(386/arm):fastrand64 链接导致新版本 Go 下包初始化即崩溃(改链 runtime.fastrand);队列 slot 数组未对齐导致 64 位原子操作 panic;GoID 汇编偏移错误
- concurrent:GoPool Shutdown 丢任务;`Int64Adder.Reset` 不归零;segQueue 扩容;TimerWheel 任务取消竞争
- seq:`OnLast` 与 `Parallel` 组合存在数据竞争(已禁止该组合,以 `Finally` 替代);Parallel 停止语义(goroutine 泄漏、panic 死锁、锁重入死锁)
- config:非法路径、未知扩展名、空参数等输入不再 panic,统一返回 error 或 nil
- cron:2 月 29 跨年计数错误
- SlidingWindow:启动期约 2 倍超发(构造时初始化时间戳消除)
- seq:三参 `FromIntSeq`/`IteratorInt` 首元素不再偏移(原实现从 start+step 起始);`Repeat` + `Take` 组合不再死循环(停止哨兵识别修正)

### Changed

性能(数据来自本地基准,详见各包 README 与 benchmark):

- GoPool 取任务延迟降低约 81%(与裸 go 差距 8.8x → 1.66x)
- seq `Last` 系列提升 10.4 倍(消除每元素取地址逃逸)
- seq `Parallel` 重写为固定 worker 池,提升 1.4 倍;order=2 归并模型提升 1.25 倍
- seq `Sort` 乱序数据提升 1.38 倍,新增有序数据快速路径
- seq `Distinct` map 化,提升 6.0 倍
- 队列超时路径 timer 池化,分配次数 3 → 1
- segQueue 扩容垃圾减少约 78%;时间轮槽 slice 复用
- logger `appendDuration` 零分配;asyncWriter 断言缓存

### Added

- 测试补齐:新增约 181 个测试函数(seq 约 70 个算子、helper 27 个函数、config 解析、logger 新 API、concurrent `TryDequeue` 等),并治理约 30 处不稳定测试(无超时等待、吞错误、零断言)
- config 新增基准:TilingMap / UntilingMap / ResolveMap / ParseConfigs2Map
- CI 工程化:`.github/workflows/go_test.yml` 新增门禁
  - vet(排除 unsafe 包 3 处已备案的 runtime hack 告警)
  - gofmt
  - `-race`(go1.27 单轨)
  - linux/386 编译
  - 自定义构建标签编译验证:`json_parser`、`concurrent_fast`、`concurrent_memory`、`nosimd`,go1.27 与 go1.18 双轨
  - govulncheck(`continue-on-error`,先观察)
  - Actions 全部 SHA pin
- 覆盖率基线(go1.27,2026-08,不设阈值门禁):

  | 包 | 语句覆盖率 |
  |---|---|
  | di | 97.2% |
  | script | 96.0% |
  | pool | 94.3% |
  | helper | 91.0% |
  | unsafe | 90.0% |
  | seq | 89.3% |
  | concurrent | 88.3% |
  | config | 88.1% |
  | logger | 79.1% |
  | storage | 67.5% |
