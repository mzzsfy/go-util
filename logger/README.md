# 日志

一个简单的日志框架,简单支持一些插件功能,方便扩展
**当前状态不够完善,需要后续继续添加一些api**

## 基础示例

```go
package main

import "github.com/mzzsfy/go-util/logger"

func main() {
    // name 规则: xx.xx.xxx, 自动建立父子继承关系
    logger.Logger("app.user").I("用户登录")
    logger.Logger("app.order").I("订单创建")

    // 支持 {} 占位符, 自动按顺序填充
    logger.Logger("app").I("用户{}下单,金额{}", "moke", 100)

    // Trace/Warn/Error/Fatal 使用 L 方法指定级别
    logger.Logger("app").L(logger.WarnLevel, "库存不足")

    // 延迟求参: 仅当日志级别通过时才执行构造函数
    logger.Logger("app").DF("调试信息", func() []any {
        return []any{expensiveCall()}
    })
}

func expensiveCall() string { return "detail" }
```

## 全局配置

日志级别与输出目标的全局默认值,通常在程序启动时设置。

```go
// 设置默认日志级别, 不接受 LevelUnset 等负值
// 返回 bool 表示是否成功
logger.SetDefaultLogLevel(logger.InfoLevel)

// 读取当前默认级别
lv := logger.DefaultLogLevel()

// 设置默认输出目标, 传 nil 时回退到 os.Stdout
logger.SetDefaultWriterTarget(os.Stdout)

// 读取当前默认输出目标
w := logger.DefaultWriterTarget()
```

### 时间格式

```go
// 年份打印格式: 0=4位年份, 1=2位年份(默认), 2=不打印年份
logger.SetPrintYearInfo(0)
```

### 调用者信息

```go
// 启用调用者信息 (输出到日志行末尾, 格式: file:line)
logger.SetCallerInfo(true)

// 同时显示函数名 (需先开启 SetCallerInfo)
logger.SetCallerFunc(true)

// 调用者跳过层数 (默认 2, 范围 0-65535)
logger.SetCallerSkip(3)
```

### 日志名与数量

```go
// 日志名显示最大长度, 默认 18
logger.SetLogNameMaxLength(20)

// 压缩过长的日志名: 0=关闭压缩, 非0=开启(默认)
logger.SetCompressedLogName(1)

// 最大 logger 数量, 0=无限制(默认 10000)
// 超出限制后新 logger 不再全局缓存, 每次返回临时实例
logger.SetMaxLoggerCount(5000)
```

## 日志级别

```go
const (
    TraceLevel Level = iota  // 记录细节, 参数等信息
    DebugLevel               // 记录重要的参数等信息
    InfoLevel                // 记录用户易读信息
    WarnLevel                // 记录需要关注的信息
    ErrorLevel               // 记录需要处理的信息
    FatalLevel               // 致命错误, 影响系统正常运行

    // LevelUnset 哨兵值: 清除本地设置, 继承父级或默认级别
    LevelUnset Level = -1
)
```

级别支持序列化 (JSON/YAML/Binary/Text),短名和全名互转:

```go
// String 返回单字符短名: T/D/I/W/E/F
logger.InfoLevel.String()    // "I"
// FullName 返回完整名称
logger.WarnLevel.FullName()  // "Warn"

// 从字符串解析级别, 大小写不敏感, 空串返回 InfoLevel
// "-" 或 "Unset" 返回 LevelUnset
lv := logger.FromString("debug") // DebugLevel
```

## Logger 管理

### 创建与继承

`Logger(name string, options ...Option) *Log` 按点分隔的 name 自动建立父子关系, 子 logger 默认继承父级的级别。同名重复调用返回同一实例。

```go
// 创建 app、app.user、app.order 三个 logger, 后两者继承 app 的级别
user := logger.Logger("app.user")
order := logger.Logger("app.order")

// 通过 Option 传入额外配置
log := logger.Logger("app", logger.WithWriter(file))
```

### 级别控制

```go
// 设置单个 logger 的级别, 传递 LevelUnset 清除本地设置并继承父级
logger.SetLevel("app.user", logger.DebugLevel)

// 递归设置 logger 及其所有子 logger
logger.SetLevelRecursive("app", logger.WarnLevel)

// 也可以直接在实例上设置
log := logger.Logger("app")
log.SetLevel(logger.ErrorLevel)
currentLevel := log.Level()
```

### 查询与清理

```go
// 遍历所有已创建的 logger 名称
logger.AllLogger()(func(name string) {
    fmt.Println(name)
})

// 移除指定 logger 及其子 logger
// 注意: 移除后已持有的引用仍可用, 但不再被全局管理, 最终被 GC 回收
logger.RemoveLogger("app.user")
```

## 结构化字段 (kvs)

通过 `With` 派生带结构化字段的 logger,kv 以 `key=value` 形式输出在消息前。奇数个参数时最后一个单独输出。

```go
log := logger.Logger("app")
log.With("userId", 123, "ip", "1.2.3.4").I("登录成功")
// 输出: ...[app]I:  userId=123 ip=1.2.3.4 登录成功

// 多次 With 会叠加父级 kvs
log.With("userId", 123).With("action", "login").I("done")
```

派生的 logger 使用完毕后可调用 `Unuse` 归还对象池。仅对 `With`/`WithWriter` 派生的实例有效,命名 logger 调用无效。

```go
derived := log.With("traceId", "abc")
defer derived.Unuse()
```

## KvPrefix 前缀

`SetKvPrefix` 注册指定 key 的格式化器。该 key 的 value 不再出现在 kvs 区域, 而是通过格式化器输出为日志行的固定前缀 (位于 level 之后、kvs 之前)。适合请求 id、用户 id 等需要突出的字段。

```go
// 注册 "traceId" 的前缀格式化器
logger.SetKvPrefix("traceId", func(value any) string {
    return "[" + fmt.Sprint(value) + "]"
})

log := logger.Logger("app").With("traceId", "req-001")
log.I("处理请求")
// 输出: ...[app]I: [req-001] 处理请求
//                          ^^^^^^^^ 来自 KvPrefix
// traceId 不会重复出现在 kvs 区域

// 当 kvs 中的值状态变化时, 调用 UpdatePrefix 刷新缓存的前缀
log.UpdatePrefix().I("再次处理")
```

## Hook 系统

Hook 在日志行写入输出前调用, 可直接操作缓冲区。每个 logger 实例共享全局 Hook 链。最多 64 个 Hook, 防止无限添加导致每次调用的开销线性增长。

```go
// Hook 签名: func(lv Level, buf *pool.Bytes, log *Log)
// 注意: pool.Bytes 来自 github.com/mzzsfy/go-util/pool

// 添加 Hook, 返回是否成功
logger.AddHook(func(lv logger.Level, buf *pool.Bytes, log *logger.Log) {
    // 在行末追加自定义标记, 注意此时末尾还没有 '\n'
    buf.WriteString(" [hook]")
})

// 移除指定 Hook (基于函数指针比较)
logger.RemoveHook(myHook)

// 清理所有全局 Hook
logger.CleanHooks()
```

## 独立 Writer

### WithWriter

通过 `WithWriter(w io.Writer) Option` 派生使用独立输出目标的 logger, 不影响全局 writer。

```go
file, _ := os.Create("app.log")
defer file.Close()

// 方式一: 创建 logger 时通过 Option 传入
log := logger.Logger("app", logger.WithWriter(file))

// 方式二: 对已有 logger 应用 Option 派生新实例
base := logger.Logger("app")
log2 := logger.WithWriter(file)(base)
defer log2.Unuse()
```

### 异步写入

配合 `helper.AsyncWriter` 可以实现异步写入, `logger.write` 热路径会自动检测 writer 是否实现了 `helper.AsyncWriter` 接口, 走异步路径避免阻塞调用方。

```go
import "github.com/mzzsfy/go-util/helper"

// 使用默认异步控制台写入器
logger.SetDefaultWriterTarget(helper.AsyncConsole())

// 或基于任意 io.Writer 创建异步写入器
aw := helper.NewAsyncWriter(file)
aw.SetFlushSize(4 * 1024)   // 刷写阈值(字节)
aw.SetCacheSize(256)        // 缓存通道大小
logger.SetDefaultWriterTarget(aw)
```

tip: 在某些特殊场景(无法传递context)你可以使用 [storage.gls](../storage/README.md#gls)
来存储一些特殊信息,比如请求id,用户id等,然后在日志插件中获取这些信息,然后打印到日志中
