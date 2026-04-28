package logger

import (
    "encoding"
    "fmt"
    "strconv"
    "sync/atomic"
    "time"

    "github.com/mzzsfy/go-util/pool"
)

// timeCache 缓存已格式化的日期时间字符串，避免重复计算
type timeCache struct {
    unix   int64  // 秒级时间戳，用于判断缓存是否过期
    prefix []byte // "YYYY-MM-DD HH:MM:SS" 合并为单个切片, 减少 writeCachedTime 中的 Write 调用次数
}

var cachedTime atomic.Value // 存储 *timeCache

func init() {
    cachedTime.Store(&timeCache{})
}

var digitBase = byte('0')

// AppendNowTime yyyy-MM-dd HH:mm:ss.SSS格式添加现在的时间
func AppendNowTime(s *pool.Bytes) {
    now := time.Now()
    unix := now.Unix()

    // 快速路径: 检查秒级缓存是否命中
    cached := cachedTime.Load().(*timeCache)
    if cached.unix == unix {
        writeCachedTime(s, cached, now)
        return
    }

    // 慢路径: CAS 确保只有一个 goroutine 更新缓存
    newCache := formatTimeCache(now, unix)
    if !cachedTime.CompareAndSwap(cached, newCache) {
        newCache = cachedTime.Load().(*timeCache)
    }
    writeCachedTime(s, newCache, now)
}

// formatTimeCache 格式化日期时间并缓存
// 将 "date HH:MM:SS" 合并为单个 prefix []byte, 热路径仅需一次 Write
func formatTimeCache(now time.Time, unix int64) *timeCache {
    year, month, day := now.Date()
    hour, min, sec := now.Clock()

    // 计算前缀总长度并一次性构建 (含末尾 '.')
    // 无年份: "MM-DD HH:MM:SS." = 15 字节
    // 2位年份: "YY-MM-DD HH:MM:SS." = 18 字节
    // 4位年份: "YYYY-MM-DD HH:MM:SS." = 20 字节
    yearInfo := atomic.LoadInt32(&printYearInfo)
    var cap_ int
    if yearInfo < 2 {
        if yearInfo == 1 {
            cap_ = 18
        } else {
            cap_ = 20
        }
    } else {
        cap_ = 15
    }

    buf := make([]byte, 0, cap_)
    // 年份部分
    if yearInfo == 0 {
        // 4位年份, 手动转换避免 strconv.Itoa 分配
        buf = append(buf, '0'+byte(year/1000))
        year %= 1000
        buf = append(buf, '0'+byte(year/100))
        year %= 100
        buf = append(buf, '0'+byte(year/10))
        buf = append(buf, '0'+byte(year%10))
        buf = append(buf, '-')
    } else if yearInfo == 1 {
        // 2位年份, year%100 保证 year<100 时也不会越界
        yy := year % 100
        buf = append(buf, digitBase+byte(yy/10), digitBase+byte(yy%10))
        buf = append(buf, '-')
    }
    // 月-日
    buf = append(buf, digitBase+byte(month/10), digitBase+byte(month%10))
    buf = append(buf, '-')
    buf = append(buf, digitBase+byte(day/10), digitBase+byte(day%10))
    buf = append(buf, ' ')
    // 时:分:秒
    buf = append(buf, digitBase+byte(hour/10), digitBase+byte(hour%10))
    buf = append(buf, ':')
    buf = append(buf, digitBase+byte(min/10), digitBase+byte(min%10))
    buf = append(buf, ':')
    buf = append(buf, digitBase+byte(sec/10), digitBase+byte(sec%10))
    buf = append(buf, '.')

    return &timeCache{unix: unix, prefix: buf}
}

// writeCachedTime 将缓存的日期时间和实时毫秒写入 buffer
// prefix 已包含 "date HH:MM:SS.", 仅需追加毫秒
func writeCachedTime(s *pool.Bytes, c *timeCache, now time.Time) {
    s.Write(c.prefix)
    append999(s, now.Nanosecond()/1e6)
}

func append999(sb *pool.Bytes, v int) {
    sb.WriteByte(digitBase + byte(v/100))
    sb.WriteByte(digitBase + byte(v/10%10))
    sb.WriteByte(digitBase + byte(v%10))
}

func appendAny(sb *pool.Bytes, arg any) {
    if arg == nil {
        sb.WriteString("<nil>")
        return
    }
    switch v := arg.(type) {
    case []byte:
        sb.Write(v)
    case string:
        sb.WriteString(v)
    case bool:
        if v {
            sb.WriteString("true")
            return
        }
        sb.WriteString("false")
    case int:
        appendInteger(sb, int64(v))
    case int8:
        appendInteger(sb, int64(v))
    case int16:
        appendInteger(sb, int64(v))
    case int32:
        appendInteger(sb, int64(v))
    case int64:
        appendInteger(sb, v)
    case uint:
        appendUInteger(sb, uint64(v))
    case uint8:
        appendUInteger(sb, uint64(v))
    case uint16:
        appendUInteger(sb, uint64(v))
    case uint32:
        appendUInteger(sb, uint64(v))
    case uint64:
        appendUInteger(sb, v)
    case uintptr:
        appendUInteger(sb, uint64(v))
    case float64:
        var buf [32]byte
        sb.Write(strconv.AppendFloat(buf[:0], v, 'f', -1, 64))
    case float32:
        var buf [32]byte
        sb.Write(strconv.AppendFloat(buf[:0], float64(v), 'f', -1, 32))
    case time.Time:
        var buf [64]byte
        sb.Write(v.AppendFormat(buf[:0], time.RFC3339Nano))
    case error:
        sb.WriteString(v.Error())
    case fmt.Stringer:
        sb.WriteString(v.String())
    case encoding.TextMarshaler:
        if b, err := v.MarshalText(); err == nil {
            sb.Write(b)
        }
    default:
        // 直接写入 buffer, 避免 fmt.Sprint 产生临时字符串分配
        fmt.Fprint(sb, v)
    }
}

// 有符号整数转字符串, 写入 buffer
// 逐位 WriteByte 避免栈上数组通过接口逃逸到堆
func appendInteger(sb *pool.Bytes, n int64) {
    if n < 0 {
        sb.WriteByte('-')
        n = -n
        // 溢出特殊处理 (math.MinInt64)
        if n < 0 {
            sb.WriteString("9223372036854775808")
            return
        }
    }
    writeUint64(sb, uint64(n))
}

// 无符号整数转字符串, 写入 buffer
func appendUInteger(sb *pool.Bytes, n uint64) {
    writeUint64(sb, n)
}

// writeUint64 将无符号整数转换为字符串写入 buffer
// pool.Bytes 是具体类型而非接口, 栈数组切片不会逃逸
func writeUint64(sb *pool.Bytes, n uint64) {
    // 快速路径: 单个数字直接写入
    if n < 10 {
        sb.WriteByte(byte('0' + n))
        return
    }
    var buf [20]byte
    pos := len(buf)
    for n >= 10 {
        pos--
        buf[pos] = byte('0' + n%10)
        n /= 10
    }
    pos--
    buf[pos] = byte('0' + n)
    sb.Write(buf[pos:])
}
