# 日志

一个简单的日志框架,简单支持一些插件功能,方便扩展


一些需要注意设置的参数
```go
// PrintBlankTag 是否打印空白tag
PrintBlankTag bool
// PrintYearInfo 打印年份信息,0打印4位,1,打印2位,2不打印
PrintYearInfo = 1
SetDefaultLogLevel()
SetDefaultWriterTarget()
```