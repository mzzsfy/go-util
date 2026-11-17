# logger

高性能日志库, 零依赖(仅标准库), 要求 Go 1.18+。双版本 API + 可插拔格式化。

核心特性:
- **双版本 API**: 性能版链式 (零分配) + 便捷版命令式 (D/I/W/E/L/DF/IF/WF/EF/LF)
- **类型安全**: Str/Int/Bool 等链式方法避免 interface 装箱
- **Formatter 接口**: ConsoleFormatter (默认, 控制台易读) / JSONFormatter (JSON 行), 支持自定义
- **级别过滤零开销**: 不达级别返回 disabled Event, 所有方法 noop
- **命名继承**: 按点分隔自动建立父子级别继承
- **Hook/Caller**: 可扩展的行钩子和调用者追踪

## 快速开始

```go
package main

import "github.com/mzzsfy/go-util/logger"

func main() {
    log := logger.New("app")

    // 性能版: 链式 API, 零分配, 类型安全
    log.Info().Str("user", "moke").Int("id", 42).Msg("login")

    // 便捷版: 命令式, 参数按 {} 占位符填充, 有装箱开销, 适合非热路径
    log.I("login user={} id={}", "moke", 42)

    // 延迟参数: 级别不够时 f 不会被调用
    log.DF("debug detail", func() []any { return expensiveArgs() })

    // 格式化消息
    log.Info().Msgf("count=%d name=%s", 5, "test")

    // 被过滤的级别零开销 (Debug < Info)
    log.Debug().Str("detail", "...").Msg("filtered") // 不产生任何输出
}
```

控制台输出 (默认):
```
26-08-05 14:30:00.123[               app] user=moke id=42 login
26-08-05 14:30:00.123[               app] login user=moke id=42
```

控制台格式不含级别标识, 级别过滤在写入前完成。

## Formatter

通过 Formatter 接口切换输出格式。两种内置实现, 常规字段均零分配。

### ConsoleFormatter (默认)

控制台易读模式, 名称右对齐填充到固定宽度:

```
26-08-05 14:30:00.123[               app] user=moke id=42 login handler.go:42
```

caller 部分需先 `SetCaller(true)` 才会输出。

NameWidth 控制名称显示宽度 (默认 18, 0 或负数一律按 18 处理, 名称超宽时截断到该宽度):
```go
logger.SetFormatter(logger.ConsoleFormatter{NameWidth: 12}) // 全局; 单 Logger 用 WithFormatter
```

### JSONFormatter

JSON 行模式, 适合采集与聚合:

```go
logger.SetFormatter(logger.JSONFormatter{})
log.Info().Str("user", "moke").Msg("login")
// {"t":"2026-08-05T14:30:00.123","lv":"Info","logger":"app","user":"moke","msg":"login"}
```

字段顺序: t, lv, logger, [链式字段], msg, [caller]。注意: JSONFormatter 不输出 With 预设字段(context), 这些字段只在 ConsoleFormatter 下可见; Any 字段仅原生支持基础类型、[]byte、error 与带 String() 的类型, 其余类型在 JSON 下输出 ""。

### 作用域

- 全局: `logger.SetFormatter(...)` 立即作用于所有未用 `WithFormatter` 固定的 Logger, 含已创建的
- 单 Logger: `logger.New("app", logger.WithFormatter(logger.JSONFormatter{}))`
- 注意: With 预设字段在创建 With 时按当时的 Formatter 编码, 之后切换 Formatter 会错配或被丢弃

### 自定义 Formatter

实现 Formatter 接口即可。可嵌入 ConsoleFormatter 只覆盖需要的方法:

```go
type MyFmt struct{ logger.ConsoleFormatter }

func (m MyFmt) Msg(buf *[]byte, msg string) {
    // 自定义消息编码
}
```

## 日志级别

```go
const (
    TraceLevel Level = iota
    DebugLevel
    InfoLevel  // 默认
    WarnLevel
    ErrorLevel
    FatalLevel // 输出后 os.Exit(1)
)

// LevelUnset 清除本地级别, 继承父级
const LevelUnset Level = -1
```

```go
log.Trace().Msg("细节")
log.Debug().Msg("调试")
log.Info().Msg("信息")
log.Warn().Msg("警告")
log.Error().Msg("错误")
log.Fatal().Msg("致命") // 写入后调用 os.Exit(1)
```

## 创建 Logger

### 独立 Logger

```go
// 快照全局默认级别和输出目标; Formatter 不快照, 每行读取全局值
log := logger.New("app")

// 指定级别、输出目标、Formatter
log := logger.New("app",
    logger.WithLevel(logger.DebugLevel),
    logger.WithWriter(os.Stderr),
    logger.WithFormatter(logger.JSONFormatter{}),
)
```

### 命名 Logger (Get)

`Get` 按点分隔自动建立父子继承, 同名返回同一实例:

```go
parent := logger.Get("app")
child := logger.Get("app.user") // 继承 parent 的级别

parent.SetLevel(logger.ErrorLevel)
// child 自动感知: child.Level() 变为 ErrorLevel
```

### 全局默认

```go
logger.SetDefaultLevel(logger.DebugLevel)   // 影响之后 New 的 Logger 与继承父级的命名 Logger
logger.SetDefaultWriter(os.Stderr)          // 传 nil 回退到 os.Stdout
logger.SetFormatter(logger.JSONFormatter{}) // 对所有未用 WithFormatter 固定的 Logger 即时生效
```

## 结构化字段

链式方法, 类型安全, 无装箱开销:

```go
log.Info().
    Str("method", "GET").
    Int("status", 200).
    Float64("latency", 0.123).
    Bool("cached", true).
    Time("ts", time.Now()).
    Dur("elapsed", time.Since(start)).
    Err(err).
    Msg("request")

// 通用方法 (有装箱开销, 热路径慎用)
log.Info().Any("data", someStruct).Msg("debug")
```

## With 预设字段

创建带预设字段的 Logger, ConsoleFormatter 下每次日志自动携带 (JSONFormatter 会丢弃这些字段):

```go
reqLog := log.With().
    Str("traceId", traceID).
    Str("method", "GET").
    Logger()

reqLog.Info().Int("status", 200).Msg("request") // 自动带 traceId, method
```

With 不修改原 Logger, 返回新实例。级别独立, 不参与命名继承。
预设字段在创建 With 时按当时的 Formatter 编码, 之后切换 Formatter 会错配或被丢弃, 需要跨 Formatter 切换时改用链式字段。
便捷版等价: `log.WithKvs("traceId", traceID, "method", "GET")`。

## Logger 管理

```go
// 递归设置级别
logger.SetLevelRecursive("app", logger.WarnLevel)

// 按名称设置级别
logger.SetLevelByName("app.user", logger.DebugLevel)

// 遍历所有命名 Logger
logger.AllLogger(func(name string) {
    fmt.Println(name)
})

// 移除 Logger 及其子级
logger.RemoveLogger("app.user")

// 已注册数量
count := logger.RegistryCount()
```

## Hook 系统

AddHook 在级别判定后、字段编码前调用; AddPostHook 在消息之后、换行与 caller 之前调用。两者都接收 `*logger.Event`, 可追加内容, 返回 bool 表示是否添加成功, 每个列表上限 64:

```go
logger.AddHook(func(e *logger.Event) {
    e.AppendString(" [svc=api]")
})

logger.AddPostHook(func(e *logger.Event) {
    e.AppendString(" done")
})

logger.RemoveHook(myHook) // 同时从两类列表移除
logger.CleanHooks()       // 清空两类列表
```

## 调用者信息

```go
logger.SetCaller(true)      // 输出 file:line, 默认关闭
logger.SetCallerFunc(true)  // 同时输出函数名 (需先 SetCaller(true), 否则无任何 caller 输出)
logger.SetCallerSkip(1)     // 跳过 1 层调用 (默认 0, 范围 0-255, 0 即直接调用日志的用户帧)
```

输出: `26-08-05 14:30:00.123[               app] login main.go:42`

## 时间格式

只替换年份段, 其余固定为 `MM-DD HH:MM:SS.mmm`:

```go
logger.SetYearMode(logger.YearFull)  // 年份 4 位: 2026-08-05 14:30:00.123
logger.SetYearMode(logger.YearShort) // 年份 2 位 (默认): 26-08-05 14:30:00.123
logger.SetYearMode(logger.YearNone)  // 无年份: 08-05 14:30:00.123
```

## 异步写入

配合 `helper.AsyncWriter` 实现异步写入, flush 自动检测:

```go
import "github.com/mzzsfy/go-util/helper"

aw := helper.NewAsyncWriter(os.Stdout)
aw.SetFlushSize(4 * 1024)
logger.SetDefaultWriter(aw)
```

## 性能

i5-8500, Windows/amd64, io.Discard, Go 1.27:

```
NoArgs              81 ns/op    0 B/op    0 allocs/op
WithFields(3)      123 ns/op    0 B/op    0 allocs/op
FilteredOut          4 ns/op    0 B/op    0 allocs/op  <- 级别过滤零开销
ParallelFiltered     1 ns/op    0 B/op    0 allocs/op
NamedGet            24 ns/op    0 B/op    0 allocs/op
WithCaller        3040 ns/op  368 B/op    2 allocs/op  <- 调用栈采集必然分配
```

级别过滤零分配: `log.Debug().Str("k","v").Msg("filtered")` 在 Info 级别下不产生任何分配。

## API 概览

### Logger 方法

| 分类 | 方法 | 说明 |
| --- | --- | --- |
| 级别 (链式) | Trace/Debug/Info/Warn/Error/Fatal | 返回 Event, 零分配链式 |
| 命令式 | D/I/W/E(msg, args...) / L(lv, msg, args...) | 直接输出, 有装箱开销 |
| 命令式 (延迟) | DF/IF/WF/EF(msg, f) / LF(lv, msg, f) | 级别不够时 f 不调用 |
| 派生 | With / WithKvs / WithContext | With 返回字段构建器 Event, 其余返回带预设字段/上下文的新 Logger |
| 状态 | SetLevel / Level / Name / Enabled | 级别与查询 |

### Event 方法

| 分类 | 方法 | 说明 |
| --- | --- | --- |
| 类型安全字段 | Str/Int/Int64/Uint64/Float64/Bool/Time/Dur/Err | 无装箱 |
| 通用字段 | Any | 有装箱, 热路径慎用 |
| 触发输出 | Msg/Msgf/MsgFormat/Send | 必须调用其一, 否则事件泄露 |
| 派生 | Logger | 从 With 事件构建新 Logger |
| Hook 用 | AppendString/AppendBytes/Level/Context | 读写缓冲区与请求上下文 |

### 配置函数 (包级)

| 分类 | 函数 |
| --- | --- |
| 全局默认 | SetDefaultLevel / DefaultLevel / SetDefaultWriter / DefaultWriter |
| 创建/选项 | New / WithLevel / WithWriter / WithFormatter |
| 级别解析 | FromString (按首字母匹配, 未识别返回 InfoLevel) |
| Formatter | SetFormatter / DefaultFormatter |
| Caller | SetCaller / SetCallerFunc / SetCallerSkip |
| 时间 | SetYearMode |
| Hook | AddHook / AddPostHook / RemoveHook / CleanHooks |
| 命名管理 | Get / Logger / SetLevelByName(返回 bool) / SetLevelRecursive / RemoveLogger / AllLogger / RegistryCount |
| 默认 Logger | Default |
| 旧版兼容 | SetDefaultLogLevel(返回 bool) / DefaultLogLevel / SetDefaultWriterTarget / DefaultWriterTarget / SetCallerInfo / SetPrintYearInfo / SetLevel |
