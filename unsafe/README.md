# 一些不安全的操作

## 获取go的hash方法

```go
hasher:=NewHasher[int]()
hasher.Hash(1)
```

## 获取goroutine id

```go
println(GoID())
```

目前仅兼容了amd64架构和arm64