package logger

import (
    "strconv"
    "strings"
    "time"
)

var (
    lastYear        int
    lastYearBytes   []byte
    lastMonth       int
    lastMonthBytes  []byte
    lastDay         int
    lastDayBytes    []byte
    lastHour        int
    lastHourBytes   []byte
    lastMinute      int
    lastMinuteBytes []byte
)

// AppendNowTime yyyy-MM-dd HH:mm:ss.SSS格式添加现在的时间
func AppendNowTime(s *strings.Builder) {
    now := time.Now()
    year, month, day := now.Date()
    if lastYear != year {
        lastYear = year
        lastYearBytes = []byte(strconv.Itoa(year))
    }
    if lastMonth != int(month) {
        lastMonth = int(month)
        lastMonthBytes = []byte(strconv.Itoa(lastMonth))
        if lastMonth < 10 {
            lastMonthBytes = append([]byte("0"), lastMonthBytes[0])
        }
    }
    if lastDay != day {
        lastDay = day
        lastDayBytes = []byte(strconv.Itoa(day))
        if lastDay < 10 {
            lastDayBytes = append([]byte("0"), lastDayBytes[0])
        }
    }
    if PrintYearInfo < 2 {
        if PrintYearInfo == 1 {
            s.Write(lastYearBytes[2:])
        } else {
            s.Write(lastYearBytes)
        }
        s.WriteByte('-')
    }
    s.Write(lastMonthBytes)
    s.WriteByte('-')
    s.Write(lastDayBytes)
    s.WriteByte(' ')
    hour, min, sec := now.Clock()
    if lastHour != hour {
        lastHour = hour
        lastHourBytes = []byte(strconv.Itoa(hour))
        if lastHour < 10 {
            lastHourBytes = append([]byte("0"), lastHourBytes[0])
        }
    }
    if lastMinute != min {
        lastMinute = min
        lastMinuteBytes = []byte(strconv.Itoa(min))
        if lastMinute < 10 {
            lastMinuteBytes = append([]byte("0"), lastMinuteBytes[0])
        }
    }
    s.Write(lastHourBytes)
    s.WriteByte(':')
    s.Write(lastMinuteBytes)
    s.WriteByte(':')
    append60(s, sec)
    s.WriteByte('.')
    append999(s, now.Nanosecond()/1e6)
}

var _0 = byte('0')

func append60(sb *strings.Builder, v int) {
    sb.WriteByte(_0 + byte(v/10))
    sb.WriteByte(_0 + byte(v%10))
}

func append999(sb *strings.Builder, v int) {
    sb.WriteByte(_0 + byte(v/100))
    sb.WriteByte(_0 + byte(v/10%10))
    sb.WriteByte(_0 + byte(v%10))
}
