# 一些简单的工具类

## 简便方法

```go
// 零拷贝转换,基于 unsafe 实现(go1.21+ 使用官方 API)
StringToBytes(s string) []byte
BytesToString(b []byte) string

// 编码与哈希
Md5(str string) string
Md5Base64(str string) string
Base64(str string) string
DeBase64(str string) string
DeBase64Byte(str string) []byte  // 解码失败返回 nil

// 运行时字符串哈希,基于 runtime.strhash
Hash(text string) uint64

// 三元运算
Ternary[T any](test bool, trueValue, falseValue T) T
TernaryF[T any, F func() T](test bool, trueValue, falseValue F) T  // 延迟求值
TernaryVF[T any, F func() T](test bool, trueValue T, falseValue F) T

// 默认值
Default[T any](test, defaultValue T) T       // test 为零值时返回 defaultValue
Defaults[T any](defaultValue T, tests ...T) T // 返回第一个非零值
NotZero(test any) bool

// 空接口与错误处理
AnyArray(vars ...any) []any
AnyArrayT[T any](vars ...T) []T
Must[T any](data T, err error) T   // err 非 nil 时 panic
MustR[T any](err error, data T) T  // 同上,参数顺序相反

// 值选择器,从多值中取指定位置
OneOfL[L, R any](data L, _ R) L
OneOfR[L, R any](_ L, data R) R
OneOf3L[L, M, R any](data L, _ M, _ R) L
OneOf3M[L, M, R any](_ L, data M, _ R) M
OneOf3R[L, M, R any](_ L, _ M, data R) R
```

## 字符串处理

```go
// PaddingOrTruncate 填充空格或截取到指定长度
// leftOrRight 填充或者截取方向,与对齐方向逻辑相反,默认 false 左对齐
PaddingOrTruncate(str string, toLen int, leftOrRight ...bool) string

// Truncate 截断字符串,保证 rune 长度不超过指定长度
Truncate(str string, toLen int, leftOrRight ...bool) string

// TruncateAndAppendSuffix 截断后若实际截断了则追加后缀
TruncateAndAppendSuffix(str string, toLen int, suffix string, leftOrRight ...bool) string

// Padding 填充空格到指定 rune 长度
Padding(str string, toLen int, leftOrRight ...bool) string

// Sub 截取字符串
// before: 保留 flag 之前还是之后
// last:   从前查找还是从后查找
Sub(src, flag string, before, last bool) string
SubBefore(src, flag string) string  // 等价 Sub(src, flag, true, false)
SubAfter(src, flag string)   // 等价 Sub(src, flag, false, true)

// SubByte 通过 byte 截取字符串
SubByte(src string, flag byte, before, last bool) string
SubByteBefore(src string, flag byte) string
SubByteAfter(src string, flag byte) string

// 判断
StringAllIsNumber(str string) bool  // 全为数字字符
StringIsInteger(str string) bool    // 整数(允许正负号前缀)

// StringBuilder 链式调用的 strings.Builder 封装
sb := &StringBuilder{}
sb.Append("a").AppendByte('b').AppendBytes([]byte("c"))
```

## 数字

```go
// 约束类型
type Signed   interface{ ~int | ~int8 | ~int16 | ~int32 | ~int64 }
type Unsigned interface{ ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr }
type Integer  interface{ Signed | Unsigned }
type Float    interface{ ~float32 | ~float64 }
type Number   interface{ Integer | Float }
type Ordered  interface{ Number | ~string }

Max[N Number](n1, n2 N) N
MaxN[N Number](ns ...N) N
Min[N Number](n1, n2 N) N
MinN[N Number](ns ...N) N
Abs[N Number](n N) N  // 单参数取绝对值,有符号整数最小值做饱和处理

// 整数转字符串,使用 sync.Pool 复用缓冲区
NumberToString[T Integer](n T) string

// 字符串解析,失败返回 defaultValue
ParseStringToInt(intStr string, defaultValue int) int
ParseStringToFloat(intStr string, defaultValue float64) float64
```

## 时间日期

```go
// Duration 常量
Duration1s, Duration10s, Duration1m, Duration10m, Duration1ms, ...

// DateTimeLayout 标准日期时间格式
DateTimeLayout = "2006-01-02 15:04:05"

// FormatDuration 格式化 time.Duration 使其长度尽量为 7 位
FormatDuration(duration time.Duration) time.Duration

// LocalTime 便于 JSON 序列化的本地时间类型
type LocalTime time.Time
// 支持 UnmarshalJSON/MarshalText/Scan/Parse 等

// ParseLocalTime 使用 DateTimeLayout 解析,长度不匹配时回退到 Auto
ParseLocalTime(str string) (LocalTime, error)
// ParseLocalTimeWithLayout 使用指定 layout 解析
ParseLocalTimeWithLayout(layout, str string) (LocalTime, error)
// ParseLocalTimeAuto 自动匹配常见格式,只支持数字格式,支持中文时间
ParseLocalTimeAuto(str string) (LocalTime, error)
```

## cron 与定时任务

```go
// ParseCron 解析 cron 表达式,支持 5~7 位,支持时区(TZ=Asia/Shanghai)和内置表达式
// @yearly @annually @monthly @weekly @daily @midnight @hourly @every 1h1s @random 1m 1h
scheduler, err := ParseCron("*/5 * * * * ?")
next := scheduler.NextTime(time.Now()) // 返回下次执行时间

// TimerWheel 分层时间轮调度器,单 goroutine 驱动,O(1) 添加/取消
tw := NewTimerWheel() // 默认 tick 间隔

// 可选配置项
tw = NewTimerWheel(
    WithTickInterval(50*time.Millisecond), // 设置 tick 间隔
    WithExecutor(func(t Task) { t.Run() }), // 自定义执行器(同步/异步)
    WithPanicHandler(func(task Task, r any) { /* ... */ }),
)

// FuncTask 将 func() 包装为 Task
tw.Schedule(time.Second*5, FuncTask(func() { /* 单次延迟 */ }))
tw.ScheduleRepeating(time.Second*10, FuncTask(func() { /* 固定间隔重复 */ }))

// ScheduleCustom 自定义调度,返回下次执行时间,零值终止
tw.ScheduleCustom(func(now time.Time) time.Time {
    return now.Add(time.Second)
}, FuncTask(func() { /* ... */ }))

// 与 cron 结合
cron, _ := ParseCron("0 * * * * ?")
tw.ScheduleCustom(cron.NextTime, FuncTask(func() { /* ... */ }))

// ScheduleRepeating/Schedule/ScheduleCustom 均返回 TaskHandle
handle := tw.Schedule(time.Second, FuncTask(func() { /* ... */ }))
handle.Cancel() // 取消任务

tw.Stop() // 停止调度器
```

## 调用栈与异常恢复

```go
// CallerStack 获取调用栈
// skip:  跳过的栈帧数(不含本函数)
// limit: 最大返回帧数,0 表示无限制
CallerStack(skip int, limit ...int) Stacks
CallerStackString(skip int, limit ...int) string
FormatStack(stack []Stack) string
SimpleFunctionName(pc uintptr) string // 去除包路径,只留函数名
FunctionName(pc uintptr) string

// TryWithStack 执行函数 f,panic 时捕获错误和调用栈
// 自动移除自身帧,首次 panic 后缓存 boundary PC 加速后续比较
TryWithStack(f func(), callback func(recoverValue any, stack Stacks))
```

## 其他工具

```go
// 反射
IsZero(test any) bool       // 判断零值,支持指针递归
New[T any](a T) T           // 创建同类型新实例
Ptr[T any](t T) *T          // 取地址

// 防抖
Debounce(call func(), duration time.Duration) func()
DebounceConcurrent(call func(), duration time.Duration) func() // 并发安全

// 并发
NewWaitGroup(init int) *sync.WaitGroup // 带初始计数的 WaitGroup

// 错误
type StringError string
NewError(s string) error

// Symbol 唯一标识符
NewSymbols(name string) NamedSymbol      // 带名称的全局唯一 Symbol
NewAnonymousSymbols() Symbol            // 匿名 Symbol

// StringBuilder 链式 strings.Builder
type StringBuilder struct{ strings.Builder }
```
