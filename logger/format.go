package logger

import (
    "fmt"
    "github.com/mzzsfy/go-util/helper"
    "strconv"
    "sync"
    "time"
)

var (
    lastYear        int
    lastYearBytes   = [4]byte{'0', '0', '0', '0'}
    lastMonth       int
    lastMonthBytes  = [2]byte{'0', '0'}
    lastDay         int
    lastDayBytes    = [2]byte{'0', '0'}
    lastHour        int
    lastHourBytes   = [2]byte{'0', '0'}
    lastMinute      int
    lastMinuteBytes = [2]byte{'0', '0'}
)

// AppendNowTime yyyy-MM-dd HH:mm:ss.SSS格式添加现在的时间
func AppendNowTime(s Buffer) {
    now := time.Now()
    year, month, day := now.Date()
    if lastYear != year {
        lastYear = year
        y := strconv.Itoa(year)
        copy(lastYearBytes[:], y[len(y)-4:])
    }
    if lastMonth != int(month) {
        lastMonth = int(month)
        if lastMonth < 10 {
            lastMonthBytes[0] = _0
            lastMonthBytes[1] = _0 + byte(lastMonth)
        } else {
            copy(lastMonthBytes[:], strconv.Itoa(lastMonth))
        }
    }
    if lastDay != day {
        lastDay = day
        if lastDay < 10 {
            lastDayBytes[0] = _0
            lastDayBytes[1] = _0 + byte(lastDay)
        } else {
            copy(lastDayBytes[:], strconv.Itoa(lastDay))
        }
    }
    if PrintYearInfo < 2 {
        if PrintYearInfo == 1 {
            s.Write(lastYearBytes[2:])
        } else {
            s.Write(lastYearBytes[:])
        }
        s.WriteByte('-')
    }
    s.Write(lastMonthBytes[:])
    s.WriteByte('-')
    s.Write(lastDayBytes[:])
    s.WriteByte(' ')
    hour, min, sec := now.Clock()
    if lastHour != hour {
        lastHour = hour
        if lastHour < 10 {
            lastHourBytes[0] = _0
            lastHourBytes[1] = _0 + byte(lastHour)
        } else {
            copy(lastHourBytes[:], strconv.Itoa(lastHour))
        }
    }
    if lastMinute != min {
        lastMinute = min
        if lastMinute < 10 {
            lastMinuteBytes[0] = _0
            lastMinuteBytes[1] = _0 + byte(lastMinute)
        } else {
            copy(lastMinuteBytes[:], strconv.Itoa(lastMinute))
        }
    }
    s.Write(lastHourBytes[:])
    s.WriteByte(':')
    s.Write(lastMinuteBytes[:])
    s.WriteByte(':')
    append60(s, sec)
    s.WriteByte('.')
    append999(s, now.Nanosecond()/1e6)
}

var _0 = byte('0')

func append60(sb Buffer, v int) {
    sb.WriteByte(_0 + byte(v/10))
    sb.WriteByte(_0 + byte(v%10))
}

func append999(sb Buffer, v int) {
    sb.WriteByte(_0 + byte(v/100))
    sb.WriteByte(_0 + byte(v/10%10))
    sb.WriteByte(_0 + byte(v%10))
}

func appendAny(sb Buffer, arg any) {
    if arg == nil {
        sb.Write(helper.StringToBytes("<nil>"))
        return
    }
    switch v := arg.(type) {
    case []byte:
        sb.Write(v)
    case string:
        sb.Write(helper.StringToBytes(v))
    case bool:
        if v {
            sb.Write(helper.StringToBytes("true"))
        }
        sb.Write(helper.StringToBytes("false"))
    case int:
        appendInteger(sb, v)
    case int8:
        appendInteger(sb, v)
    case int16:
        appendInteger(sb, v)
    case int32:
        appendInteger(sb, v)
    case int64:
        appendInteger(sb, v)
    case uint:
        appendInteger(sb, v)
    case uint8:
        appendInteger(sb, v)
    case uint16:
        appendInteger(sb, v)
    case uint32:
        appendInteger(sb, v)
    case uint64:
        appendInteger(sb, v)
    case uintptr:
        appendInteger(sb, v)
    default:
        sb.Write(helper.StringToBytes(fmt.Sprint(v)))
    }

}

//18446744073709551615 uint64
//-9223372036854775808 int64
var buf = sync.Pool{
    New: func() any {
        return &[20]byte{}
    },
}

func appendInteger[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~int | ~int8 | ~int16 | ~int32 | ~int64](sb Buffer, n T) {
    bs := buf.Get().(*[20]byte)
    defer buf.Put(bs)
    r := bs[:]
    negative := false
    i := len(r) - 1
    if n < 0 {
        n = -n
        //溢出,特殊处理
        // math.MinInt8
        // math.MinInt16
        // math.MinInt32
        // math.MinInt64
        if n < 0 {
            next := n / 10
            r[i] = byte('0' + -(n - next*10))
            n = -next
            i--
        }
        negative = true
    }
    for n > 9 {
        next := n / 10
        r[i] = byte('0' + n - next*10)
        n = next
        i--
    }
    r[i] = byte('0' + n)
    if negative {
        i--
        r[i] = '-'
    }
    sb.Write(r[i:])
}
