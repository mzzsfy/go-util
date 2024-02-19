# 一些不安全的操作

## hasher

获取go的hash方法,目前兼容了1.18以后的版本

```go
hasher:=NewHasher[int]()
hasher.Hash(1)
```

## goroutine id

获取goroutine id,目前仅兼容了部分架构的汇编模式获取,其他架构暂时使用runtime.Stack

```go
println(GoID())
```
