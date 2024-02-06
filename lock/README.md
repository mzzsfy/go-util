# 锁工具

## 可重入锁

依赖goid,可修改GoID为自定义的获取goroutine id的函数

```go
lock:=NewReentrantLock()
lock.Lock()
lock.Lock()
lock.Unlock()
lock.Unlock()
```
