# 池化工具

## 对象池

简单的sync.Pool泛型封装

```go
pool:=NewObjpool[user](func() *user { return &user{} }, func(u *user) { u.id = nil;u.name="" })
u:=pool.get()
pool.put(u)
```
## 携程池

一个简单实现的携程池,参考了字节跳动的方案,代码有待优化,建议不要使用

```go
//使用默认携程池运行任务
Go(func() {
	//code
})

//自定义携程池运行任务
pool:=NewGopool(withName("pool"), withMaxSize(1000))
pool.Go(func() {
//code
}))

//停止
pool.Shutdown()
//重启
pool.Restart()
```
