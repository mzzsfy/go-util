# logger

高性能日志库, 零依赖(仅标准库), 要求 Go 1.18+。双版本 API + 可插拔格式化。

核心特性:
- **双版本 API**: 性能版链式 (零分配) + 便捷版命令式 (D/I/L/DF/IF/LF)
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

    // 便捷版: 命令式, 有装箱开销, 适合非热路径
    log.I("login", "user", "moke", "id", 42)

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
26-08-05 14:30:00.123[  app]I: user=moke id=42 login
```

## Formatter

通过 Formatter 接口切换输出格式。两种内置实现, 均零分配。

### ConsoleFormatter (默认)

控制台易读模式, 名称左对齐填充到固定宽度:

```
26-08-05 14:30:00.123[  app]I: user=moke id=42 login handler.go:42
```

NameWidth 控制名称显示宽度 (默认 18, 0=自适应):
```go
logger.SetFormatter(logger.ConsoleFormatter{NameWidth: 12})
```

### JSONFormatter

JSON 行模式, 适合采集与聚合:

```go
logger.SetFormatter(logger.JSONFormatter{})
log.Info().Str("user", "moke").Msg("login")
// {"t":"2026-08-05T14:30:00.123","lv":"Info","logger":"app","user":"moke","msg":"login"}
```

字段顺序: t, lv, logger, [With 预设字段], [链式字段], msg, [caller]

### 作用域

- 全局: `logger.SetFormatter(...)` 影响后续 New/Get 创建的 Logger
- 单 Logger: `logger.New("app", logger.WithFormatter(logger.JSONFormatter{}))`

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
// 从全局默认快照级别、输出目标和 Formatter
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
// child 自动感知: EffectiveLevel 变为 ErrorLevel
```

### 全局默认

```go
logger.SetDefaultLevel(logger.DebugLevel)   // 影响 New 创建的 Logger
logger.SetDefaultWriter(os.Stderr)          // 传 nil 回退到 os.Stdout
logger.SetFormatter(logger.JSONFormatter{}) // 影响后续 New/Get
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

创建带预设字段的 Logger, 每次日志自动携带:

```go
reqLog := log.With().
    Str("traceId", traceID).
    Str("method", "GET").
    Logger()

reqLog.Info().Int("status", 200).Msg("request") // 自动带 traceId, method
```

With 不修改原 Logger, 返回新实例。级别独立, 不参与命名继承。
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

Hook 在写入前调用, 可追加内容到缓冲区:

```go
logger.AddHook(func(e *logger.Event) {
    e.AppendString(" [svc=api]")
})

logger.RemoveHook(myHook)
logger.CleanHooks()
```

## 调用者信息

```go
logger.SetCaller(true)      // 输出 file:line
logger.SetCallerFunc(true)  // 同时输出函数名
logger.SetCallerSkip(4)     // 跳过层数 (默认 4, 范围 0-255)
```

输出: `26-08-05 14:30:00.123[  app]I: login main.go:42`

## 时间格式

```go
logger.SetYearMode(logger.YearFull)  // YYYY-MM-DD (4 位年份)
logger.SetYearMode(logger.YearShort) // YY-MM-DD (2 位年份, 默认)
logger.SetYearMode(logger.YearNone)  // MM-DD (无年份)
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

i5-8500, io.Discard, Go 1.25:

```
NoArgs              81 ns/op    0 B/op    0 allocs/op
WithFields(3)      123 ns/op    0 B/op    0 allocs/op
FilteredOut          4 ns/op    0 B/op    0 allocs/op  <- 级别过滤零开销
ParallelFiltered     1 ns/op    0 B/op    0 allocs/op
NamedGet            24 ns/op    0 B/op    0 allocs/op
WithCaller         362 ns/op    0 B/op    0 allocs/op
```

级别过滤零分配: `log.Debug().Str("k","v").Msg("filtered")` 在 Info 级别下不产生任何分配。

## API 概览

### Logger 方法

| 分类 | 方法 | 说明 |
| --- | --- | --- |
| 级别 (链式) | Trace/Debug/Info/Warn/Error/Fatal | 返回 Event, 零分配链式 |
| 命令式 | D/I/L(lv, ...) | 直接输出, 有装箱开销 |
| 命令式 (延迟) | DF/IF/LF(lv, ...) | 级别不够时 f 不调用 |
| 派生 | With / WithKvs | 返回带预设字段的新 Logger |
| 状态 | SetLevel / Level / Name / Enabled | 级别与查询 |

### Event 方法

| 分类 | 方法 | 说明 |
| --- | --- | --- |
| 类型安全字段 | Str/Int/Int64/Uint64/Float64/Bool/Time/Dur/Err | 无装箱 |
| 通用字段 | Any | 有装箱, 热路径慎用 |
| 触发输出 | Msg/Msgf/Send | 必须调用其一, 否则事件泄露 |
| 派生 | Logger | 从 With 事件构建新 Logger |
| Hook 用 | AppendString/AppendBytes/Level | 写入缓冲区 |

### 配置函数 (包级)

| 分类 | 函数 |
| --- | --- |
| 全局默认 | SetDefaultLevel / DefaultLevel / SetDefaultWriter / DefaultWriter |
| Formatter | SetFormatter / DefaultFormatter |
| Caller | SetCaller / SetCallerFunc / SetCallerSkip |
| 时间 | SetYearMode |
| Hook | AddHook / RemoveHook / CleanHooks |
| 命名管理 | Get / SetLevelByName / SetLevelRecursive / RemoveLogger / AllLogger / RegistryCount |
| 默认 Logger | Default |
