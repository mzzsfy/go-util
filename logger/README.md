# 日志

一个简单的日志框架,简单支持一些插件功能,方便扩展
**当前状态不够完善,需要后续继续添加一些api**


一些需要注意设置的参数
```go
// PrintBlankTag 是否打印空白tag
PrintBlankTag bool
// PrintYearInfo 打印年份信息,0打印4位,1,打印2位,2不打印
PrintYearInfo = 0
SetDefaultLogLevel()
SetDefaultWriterTarget()
```

tip: 在某些特殊场景(无法传递context)你可以使用 [storage.gls](../storage/README.md#gls) 来存储一些特殊信息,比如请求id,用户id等,然后在日志插件中获取这些信息,然后打印到日志中