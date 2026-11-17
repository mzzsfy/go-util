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

获取当前 goroutine 的 id,返回类型为 `int64`。

```go
id := GoID()
println(id)
```

### 实现与版本兼容

`GoID` 的签名统一为 `func GoID() int64`。amd64/arm64/386/arm/amd64p32 由 `goid_s.go` 联合 `goid_*.s` 的汇编实现提供;各 Go 版本只在独立文件中声明 g 结构体 `goid` 字段的偏移量,由 build tag 选择:

| Go 版本        | amd64 / arm64      | 386 / arm / amd64p32    |
| -------------- | ------------------ | ----------------------- |
| go1.10 ~ go1.22 | `goid_s_1.10.go`   | `goid_s_1.10_32.go`     |
| go1.23 ~ go1.24 | `goid_s_1.23.go`   | `goid_s_1.23_32.go`     |
| go1.25+        | `goid_s_1.25.go`   | `goid_s_1.25_32.go`     |

表格中的文件只提供偏移量常量,汇编取值由 `goid_*.s` 的 getG 完成。其他架构回退到 `goid_stack.go` 中基于 `runtime.Stack` 解析的实现。`amd64p32` 属于已移除的 NaCl 目标,保留该 build tag 仅为历史兼容。
