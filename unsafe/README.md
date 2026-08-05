# 一些不安全的操作

## hasher

获取go的hash方法,目前兼容了1.18以后的版本

```go
type Hasher[K comparable] interface {
    Hash(key K) uint64
    NewSeed() Hasher[K]              // 使用新随机种子
    WithSeed(seed uintptr) Hasher[K] // 使用指定种子
}

hasher := NewHasher[int]()
hasher.Hash(1)

h2 := hasher.NewSeed()       // 新随机种子
h3 := hasher.WithSeed(123)   // 指定种子
```

## goroutine id

获取goroutine id,目前仅兼容了部分架构的汇编模式获取,其他架构暂时使用runtime.Stack

```go
println(GoID())
```
