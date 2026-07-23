# 存储相关

## gls

goroutine local storage, 用于存储goroutine本地数据,类似java的threadlocal。

使用前必须调用 `KnowHowToUseGls()` 确认你了解清理义务;每个使用了 gls 的 goroutine 退出前必须调用 `GlsClean()`,否则会触发泄露检测 panic。`GlsClean()` 和 `GlsCleanWithId(goid)` 是包级函数,不是 Key 的方法。

```go
// 初始化,确认你知道需要调用 GlsClean
storage.KnowHowToUseGls()

// 全局声明 key,不要每次声明
var item = storage.NewGlsItem[string]()

// 使用
item.Set("testValue")
item.Set("newValue")
value, ok := item.Get()

// 删除单个 key,传 true 表示删除后若当前 g 无其他 key 则自动清理整个 gls
item.Delete(true)

// goroutine 退出前清理本 g 的全部 gls 数据
storage.GlsClean()
```

其他构造器:

- `NewGlsItemWithDefault[T](def T) Key[T]`:带默认值,`Get` 在未 `Set` 时返回默认值。
- `NewGlsItemWithFunc[T](def func() T) Key[T]`:带延迟初始化,首次 `Get` 时调用工厂函数。

`Key[T]` 接口方法:

- `Get() (T, bool)`
- `GetById(goid int64) (T, bool)`:跨 goroutine 读取窗口。
- `Set(T)`
- `Delete(autoClean ...bool)`:删除当前 g 的该 key;`autoClean` 为 true 时,若清理后当前 g 无其他 key,则自动清理整个 gls 子表。

## SwissMap

基于 https://github.com/dolthub/swiss 的 map,相比go自带map,性能更高,内存占用更低。go1.24+ 运行时内置 map 已采用 swiss table,此时 `MapTypeSwiss` 直接转发到 `MapTypeGo`。

```go
m := storage.NewMap(storage.MapTypeSwiss[int, int](16))
m.Put(key, i)
```

## Map 类型与构造器

所有 map 实现统一实现 `Map[K comparable, V any]` 接口:`Has`、`Get`、`GetSimple`、`Put`、`Delete`、`Iter`、`Clean`、`Count`。

`NewMap[K, V](opt ...MakeMap[K, V]) Map[K, V]` 为可变参数构造器:不传参时默认使用 `MapTypeSwiss[K, V](16)`;传一个 `MapTypeXxx` 时使用指定类型。

可选类型:

| 构造器 | 说明 |
| --- | --- |
| `MapTypeGo[K, V](cap int)` | go 原生 `map` 封装 |
| `MapTypeArray[K, V](size int)` | 基于数组的线性扫描 map,适合小数据量(低于约50个元素),空间利用率高 |
| `MapTypeSwiss[K, V](size uint32)` | swiss table(go1.24+ 退化为 `MapTypeGo`) |
| `MapTypeSwissConcurrent[K, V]()` | 内置分片并发 swiss map(go1.24+ 退化为分片 go map) |
| `MapTypeConcurrentWrapper[K, V](m MakeMap[K, V])` | 对任意底层 map 做分片并发包装,分片数随 CPU 核数自适应 |
| `MapTypeConcurrentLockWrapper[K, V](m MakeMap[K, V])` | 对任意底层 map 做轻量 `sync.RWMutex` 整体加锁包装 |

```go
// 默认 swiss
m1 := storage.NewMap[int, string]()

// go 原生 map
m2 := storage.NewMap(storage.MapTypeGo[int, string](8))

// 小数据量用 array
m3 := storage.NewMap(storage.MapTypeArray[int, string](8))

// 并发安全:对 array 做分片包装
m4 := storage.NewMap(storage.MapTypeConcurrentWrapper(
    storage.MapTypeArray[int, string](4),
))

// 并发安全:对 swiss 做轻量锁包装
m5 := storage.NewMap(storage.MapTypeConcurrentLockWrapper(
    storage.MapTypeSwiss[int, string](8),
))
```

`IterDeleteMap[K, V]` 接口在 `Map` 之上扩展 `IterDelete(cb func(k K, v V) (del, stop bool)) bool`,支持遍历时安全删除。对不支持该方法的底层 map,可使用包级函数 `IterDelete(m, cb)` 兜底(内部先收集再删除)。

## 缓存

提供缓存接口与并发安全包装,不绑定具体实现,由调用方自行选择底层存储。

接口:

- `Cache[K comparable, V any]`:`Get`、`Set`、`Delete`、`Clear`、`Size`。
- `TimedCache[K comparable, V any]`:在 `Cache` 之上扩展 `SetWithTimeout(key, value, timeout)` 与 `TTL(key)`,支持单 key 过期。

`CacheWrap[K, V]` 对任意 `Cache` 做加锁包装,核心方法 `GetOr(key, def func() V) V` 采用 double-check:命中直接返回;未命中加锁后二次检查,仍未命中才调用 `def` 生成值并写入。

```go
// 用任意 Map 实现一个简单 Cache
type myCache[K comparable, V any] struct{ m storage.Map[K, V] }
// ... 实现 Cache 接口方法 ...

wrap := storage.NewCacheWrap[string, int](&myCache[string, int]{})
v := wrap.GetOr("key1", func() int {
    // 未命中时执行,结果会写入缓存
    return computeExpensive()
})
```

`NewCacheWrap[K, V](cache Cache[K, V]) *CacheWrap[K, V]` 为唯一构造器,内部使用 `sync.Mutex` 保护 `GetOr` 的加载路径,其余读写依赖传入底层 `Cache` 自身的并发安全性。
