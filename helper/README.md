# 一些简单的工具类

## 简便方法

```go
StringToBytes(str string) []byte
BytesToString(b []byte) string
Md5(str string) string
Base64(str string) string

// PaddingOrTruncate 填充空格或截取到指定长度
// leftOrRight 填充或者截取方向,与对齐方向逻辑相反,默认false,左对齐
PaddingOrTruncate(str string, toLen int, leftOrRight ...bool) string

// TruncateAndAppendSuffix 截断字符串到指定长度,如果截断了,则添加后缀
TruncateAndAppendSuffix(str string, toLen int, suffix string, leftOrRight ...bool) string

// Truncate 截断字符串,保证长度不超过指定长度
Truncate(str string, toLen int, leftOrRight ...bool) string

// Sub 截取字符串
// before 保留flag之前的字符串还是之后
// last 从前开始查找还是从后开始查找
Sub(src, flag string, before, last bool) string

// StringAllIsNumber 字符串是否全是数字
StringAllIsNumber(str string) bool

//三元运算
Ternary[T any](test bool, trueValue, falseValue T) T
Default[T any](test, defaultValue T) T
TryWithStack(f func(), callback func(recoverValue any, stack []Stack))
```

## cron任务,延迟执行

```go
s := NewScheduler()
s.AddIntervalTask(func() {code... }, time.Millisecond*200)
s.AddDelayTask(func() {code... }, time.Millisecond*200)
s.AddCustomizeTask(func() {code... }, func(t time.Time) time.Time { return t.Add(...) })
s.AddCronTask("0", func(){})
```

## 批量按序运行

```go
//注册启动后运行回调
AfterInit(name string, f func())

caller:=&FuncCaller{}
caller.AddFnOrder(func() {code...})
```

## 数字

```go
NumberToString[T Number](T) string
Min[T Number](a, b T) T
Abs[T Number](a, b T) T
```

## 时间日期

```go
// ParseLocalTimeAuto 自动匹配常见格式,只支持数字格式,支持常见中文
ParseLocalTimeAuto(str string) (LocalTime, error)

// FormatDuration 格式化time.Duration 使其长度尽量为7位
FormatDuration(duration time.Duration) time.Duration 
```

## 布隆过滤器

todo

## dfa查找

```go
dfa:=NewDfa(MakeNewDfsNode[bool](i))
dfa.Add([]byte("aaa"), true) //可以存储这个词对应的信息
dfa.Test([]byte("aaa"))
```

## 栈获取

```go
CallerStack(skip int, limit ...int) []Stack
```
