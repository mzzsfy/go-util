package helper

import (
    "math/bits"
    "time"
)

type cronTimer struct {
    year  []int
    month uint16
    week  uint8
    //L,C,W,# 从高到低位含义为
    //日:L月末,W工作日,周:L最后一周,W最近的工作日(不支持),空2位,(3位)第几周#(可用1~7,实际为1~5)
    //lcw uint8

    day, hour      uint32
    second, minute uint64

    local *time.Location
}

//获取下一个合法值
func getNextNumber(b int, Range uint64) (i int, circulate bool) {
    //右移b位,判断是否需要进位
    t := Range >> b
    if t == 0 {
        return bits.TrailingZeros64(Range), true
    }
    //获取低位有多少个0,并加上基础值b
    return bits.TrailingZeros64(t) + b, false
}

//检查是否是合法值
func checkNumber(b int, Range uint64) bool {
    return Range>>b&1 != 0
}

func (s *cronTimer) NextTime(t time.Time) time.Time {
    year, month, day := t.Date()
    hour, minu, sec := t.Clock()
    var fixed bool

    //修正原始时间
    if !checkNumber(sec, s.second) {
        sec1, _ := getNextNumber(sec+1, s.second)
        fixed = sec != sec1
        if sec1 < sec {
            minu++
        }
        sec = sec1
    }
    if !checkNumber(minu, s.minute) {
        minu1, _ := getNextNumber(minu+1, s.minute)
        fixed = fixed || minu != minu1
        if minu1 < minu {
            hour++
            sec, _ = getNextNumber(0, s.second)
        }
        minu = minu1
    }
    if !checkNumber(hour, uint64(s.hour)) {
        hour1, _ := getNextNumber(hour+1, uint64(s.hour))
        fixed = fixed || hour != hour1
        if hour1 < hour {
            day++
            minu, _ = getNextNumber(0, s.minute)
            sec, _ = getNextNumber(0, s.second)
        }
        hour = hour1
    }
    if !checkNumber(day, uint64(s.day)) {
        day1, _ := getNextNumber(day+1, uint64(s.day))
        fixed = fixed || day != day1
        if day1 < day {
            month++
            hour, _ = getNextNumber(0, uint64(s.hour))
            minu, _ = getNextNumber(0, s.minute)
            sec, _ = getNextNumber(0, s.second)
        }
        day = day1
    }
    if !checkNumber(int(month), uint64(s.month)) {
        month1, _ := getNextNumber(int(month)+1, uint64(s.month))
        fixed = fixed || int(month) != month1
        if month1 < int(month) {
            year++
            month = time.Month(month1)
            day, _ = getNextNumber(0, uint64(s.day))
            sec, _ = getNextNumber(0, s.second)
            minu, _ = getNextNumber(0, s.minute)
            hour, _ = getNextNumber(0, uint64(s.hour))
        }
        month = time.Month(month1)
    }

    if len(s.year) > 0 {
        yearMatched := false
        for _, y := range s.year {
            if y > year {
                year = y
                day = 0
                month = time.Month(1)
                sec, _ = getNextNumber(0, s.second)
                minu, _ = getNextNumber(0, s.minute)
                hour, _ = getNextNumber(0, uint64(s.hour))
                fixed = true
                yearMatched = true
                break
            }
            if y == year {
                year = y
                yearMatched = true
                break
            }
        }
        // 所有指定年份都已过期, 无法调度
        if !yearMatched {
            return time.Time{}
        }
    }
    if !fixed {
        var circulate bool
        if sec, circulate = getNextNumber(sec+1, s.second); !circulate {
            goto ok
        }
        if minu, circulate = getNextNumber(minu+1, s.minute); !circulate {
            goto ok
        }
        if hour, circulate = getNextNumber(hour+1, uint64(s.hour)); circulate {
            var month1 int
            day, month1, year = s.nextDay(day, int(month), year)
            month = time.Month(month1)
            if year == 0 {
                return time.Time{}
            }
        }
    ok:
    }
    {
        if day == 0 {
            var month1 int
            day, month1, year = s.nextDay(day, int(month), year)
            month = time.Month(month1)
            if year == 0 {
                return time.Time{}
            }
        } else if s.week != 0x7f {
            t1 := weekdayOf(year, month, day)
            weekday := t1
            if (s.week>>uint8(weekday))&1 != 1 {
                var month1 int
                day, month1, year = s.nextDay(day, int(month), year)
                month = time.Month(month1)
                if year == 0 {
                    return time.Time{}
                }
            }
        }
    }

    // 校验 day 不超过该月实际天数
    if day > 0 && day > daysInMonth(year, month) {
        var month1 int
        day, month1, year = s.nextDay(day, int(month), year)
        month = time.Month(month1)
        if year == 0 {
            return time.Time{}
        }
    }

    if year == 0 {
        return time.Time{}
    }
    return time.Date(year, month, day, hour, minu, sec, 0, s.local)
}

// daysPerMonth 非闰年每月天数
var daysPerMonth = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// daysInMonth 返回指定年份和月份的实际天数
func daysInMonth(year int, month time.Month) int {
    m := int(month)
    if m < 1 || m > 12 {
        return 31
    }
    d := daysPerMonth[m-1]
    if m == 2 && isLeapYear(year) {
        d++
    }
    return d
}

// isLeapYear 判断是否为闰年
func isLeapYear(year int) bool {
    return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// sakamotoTable Sakamoto 算法查找表, 用于快速计算星期几
var sakamotoTable = [12]int{0, 3, 2, 5, 0, 3, 5, 1, 4, 6, 2, 4}

// weekdayOf 使用 Sakamoto 算法计算星期几, 避免 time.Date 的日历构造开销
// 返回 0=Sunday ... 6=Saturday, 与 time.Weekday 偏移一致
func weekdayOf(y int, m time.Month, d int) time.Weekday {
    mm := int(m)
    if mm < 3 {
        y--
    }
    return time.Weekday((y + y/4 - y/100 + y/400 + sakamotoTable[mm-1] + d) % 7)
}

func (s *cronTimer) nextDay(day, month, year int) (int, int, int) {
    circulate := false
    nextYear := false
    for {
        if day, circulate = getNextNumber(day+1, uint64(s.day)); !circulate {
            goto testWeek
        }
        if month, circulate = getNextNumber(month+1, uint64(s.month)); circulate {
            if len(s.year) == 0 {
                //未设置年份,允许进位一次
                if !nextYear {
                    year++
                    day, _ = getNextNumber(0, uint64(s.day))
                    month, _ = getNextNumber(0, uint64(s.month))
                } else {
                    return 0, 0, 0
                }
                nextYear = true
            } else {
                for _, y := range s.year {
                    if y > year {
                        year = y
                        month = 1
                        goto testWeek
                    }
                }
                return 0, 0, 0
            }
        }
    testWeek:
        // 校验 day 不超过该月实际天数, 避免 time.Date 自动溢出到下月
        maxDay := daysInMonth(year, time.Month(month))
        if day > maxDay {
            // 当前 day 在该月不存在, 进位到下月重新查找合法 day
            month, circulate = getNextNumber(month+1, uint64(s.month))
            if circulate {
                // 月溢出, 走年份进位逻辑 (与上面的月进位代码相同)
                if len(s.year) == 0 {
                    if !nextYear {
                        year++
                        day, _ = getNextNumber(0, uint64(s.day))
                        month, _ = getNextNumber(0, uint64(s.month))
                    } else {
                        return 0, 0, 0
                    }
                    nextYear = true
                } else {
                    for _, y := range s.year {
                        if y > year {
                            year = y
                            month = 1
                            day, _ = getNextNumber(0, uint64(s.day))
                            goto testWeek
                        }
                    }
                    return 0, 0, 0
                }
            } else {
                // 月份进位成功, 重新查找合法 day
                day, _ = getNextNumber(0, uint64(s.day))
            }
            goto testWeek
        }
        if s.week == 0x7f {
            break
        }
        weekday := weekdayOf(year, time.Month(month), day)
        if (s.week>>uint8(weekday))&1 == 1 {
            break
        }
    }
    return day, month, year
}
