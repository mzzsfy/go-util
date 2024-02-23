# 存储相关

## gls

goroutine local storage, 用于存储goroutine本地数据,类似java的threadlocal

```go
item := NewGlsItem[string]()
item.Set("testValue")
item.Set("newValue")
value, ok := item.Get()
item.GlsClean()
```

## SwissMap

基于 https://github.com/dolthub/swiss 的map,相比go自带map,性能更高,内存占用更低

```go
m := NewMap(MapTypeSwiss[int, int](uint32(len(keys))))
m.Put(key, i)
```

todo: 支持并发