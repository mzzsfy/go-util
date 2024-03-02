package logger

import (
    "io"
    "strconv"
    "time"
)

var (
    lastYear        int
    lastYearBytes   [4]byte
    lastMonth       int
    lastMonthBytes  [2]byte
    lastDay         int
    lastDayBytes    [2]byte
    lastHour        int
    lastHourBytes   [2]byte
    lastMinute      int
    lastMinuteBytes [2]byte
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
            lastHourBytes[1] = byte(lastHour)
        } else {
            copy(lastHourBytes[:], strconv.Itoa(lastHour))
        }
    }
    if lastMinute != min {
        lastMinute = min
        if lastMinute < 10 {
            lastMinuteBytes[0] = _0
            lastMinuteBytes[1] = byte(lastMinute)
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

func append60(sb interface {
    io.Writer
    WriteByte(byte) error
}, v int) {
    sb.WriteByte(_0 + byte(v/10))
    sb.WriteByte(_0 + byte(v%10))
}

func append999(sb interface {
    io.Writer
    WriteByte(byte) error
}, v int) {
    sb.WriteByte(_0 + byte(v/100))
    sb.WriteByte(_0 + byte(v/10%10))
    sb.WriteByte(_0 + byte(v%10))
}
