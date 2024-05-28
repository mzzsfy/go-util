package helper

import (
    "errors"
    "math"
    "math/bits"
    "math/rand"
    "strconv"
    "strings"
    "time"
)

type everyTask time.Duration

func (e everyTask) NextTime(t time.Time) time.Time {
    return t.Add(time.Duration(e))
}

type randomTask struct {
    offset time.Duration
    rand   int64
}

func (e *randomTask) NextTime(t time.Time) time.Time {
    return t.Add(time.Duration(rand.Int63n(e.rand)) + e.offset)
}

var (
    cronRules = [][]uint64{
        {0, 59},
        {0, 59},
        {0, 23},
        {1, 31},
        {1, 12},
        {0, 7},
        {1990, 9999},
    }
    weekMap = map[string]string{
        "SUN": "1",
        "MON": "2",
        "TUE": "3",
        "WED": "4",
        "THU": "5",
        "FRI": "6",
        "SAT": "7",
    }
    monthMap = map[string]string{
        "JAN": "1",
        "FEB": "2",
        "MAR": "3",
        "APR": "4",
        "MAY": "5",
        "JUN": "6",
        "JUL": "7",
        "AUG": "8",
        "SEP": "9",
        "OCT": "10",
        "NOV": "11",
        "DEC": "12",
    }
)

//解析cron表达式中非年份部分,星期和月份部分需要特殊处理
func parseCronItem(cronItem string, min, max uint64) (s uint64, e error) {
    if strings.Contains(cronItem, "/") {
        //处理步长
        start, end, step, err := parseStep(cronItem, min, max)
        if err != nil {
            e = err
            return
        }
        for i := start; i <= end; i += step {
            s |= 1 << i
        }
        return
    }
    //处理范围
    if strings.Contains(cronItem, "-") {
        start, end, err := parseRange(cronItem, min, max)
        if err != nil {
            e = err
            return
        }
        for i := start; i <= end; i++ {
            s |= 1 << i
        }
        return
    }
    //处理*号
    if cronItem == "*" || cronItem == "?" {
        s = parseBlurRange(min, max)
        return
    }
    r, err := parseSingle(cronItem, min, max)
    if err != nil {
        e = err
        return
    }
    for _, i := range r {
        s |= 1 << i
    }
    return
}

func parseSingle(item string, min uint64, max uint64) (r []uint64, e error) {
    items := strings.Split(item, ",")
    //处理单个时间
    for _, str := range items {
        i, err := strconv.ParseInt(str, 10, 64)
        if err != nil {
            e = joinErrs(errors.New(item+"无法解析为数字"), err)
            return
        }
        if i < int64(min) || i > int64(max) {
            e = errors.New("范围错误,范围必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(i)) + "超出范围")
            return
        }
        r = append(r, uint64(i))
    }
    return
}

func parseBlurRange(min, max uint64) uint64 {
    s := uint64(math.MaxUint64)
    s = s >> (63 - max)
    s = s >> min << min
    return s
}

func parseStep(item string, min, max uint64) (start, end, step uint64, e error) {
    stepItem := strings.Split(item, "/")
    if len(stepItem) != 2 {
        e = errors.New("解析错误,步长错误必须'a/b'格式,尝试解析的值为:" + item)
        return
    }
    start = min
    end = max
    if strings.Contains(stepItem[0], "-") {
        //最复杂的情况,包含/和-
        start1, end1, err := parseRange(stepItem[0], min, max)
        if err != nil {
            e = err
            return
        }
        start = uint64(start1)
        end = uint64(end1)
    } else if stepItem[0] != "*" {
        start1, err := strconv.ParseInt(stepItem[0], 10, 64)
        if err != nil {
            e = joinErrs(errors.New(stepItem[0]+"无法解析为数字"), err)
            return
        }
        start = uint64(start1)
    }
    if start > end {
        e = errors.New("范围错误,开始不能大于结束," + strconv.Itoa(int(start)) + ">" + strconv.Itoa(int(end)))
        return
    }
    if start < min {
        e = errors.New("范围错误,开始必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(start)) + "超出范围")
        return
    }
    if end > max {
        e = errors.New("范围错误,结束必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(end)) + "超出范围")
        return
    }
    step1, err := strconv.ParseInt(stepItem[1], 10, 64)
    if err != nil {
        e = joinErrs(errors.New(stepItem[1]+"无法解析为数字"), err)
        return
    }
    step = uint64(step1)
    if step == 0 {
        e = errors.New("步长错误,步长不能为0,尝试解析的值为:" + item)
        return
    }
    if step > max {
        e = errors.New("步长错误,步长不能大于" + strconv.Itoa(int(max)) + ",尝试解析的值为:" + item)
        return
    }
    return
}

func parseRange(item string, min, max uint64) (start int64, end int64, e error) {
    rangeItem := strings.Split(item, "-")
    if len(rangeItem) != 2 {
        e = errors.New("解析错误,范围错误必须'a-b'格式")
        return
    }
    start, err := strconv.ParseInt(rangeItem[0], 10, 64)
    if err != nil {
        e = joinErrs(errors.New(rangeItem[0]+"无法解析为数字"), err)
        return
    }
    end, err = strconv.ParseInt(rangeItem[1], 10, 64)
    if err != nil {
        e = joinErrs(errors.New(rangeItem[1]+"无法解析为数字"), err)
        return
    }
    if start > end {
        e = errors.New("范围错误,开始不能大于结束," + strconv.Itoa(int(start)) + ">" + strconv.Itoa(int(end)))
        return
    }
    if start < int64(min) {
        e = errors.New("范围错误,开始必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(start)) + "超出范围")
        return
    }
    if end > int64(max) {
        e = errors.New("范围错误,结束必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(end)) + "超出范围")
        return
    }
    return
}

type schedulerCron struct {
    corn  string
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
func getNextNumber(b int8, Range uint64) (i int, circulate bool) {
    //右移b位,判断是否需要进位
    t := Range >> b
    if t == 0 {
        return bits.TrailingZeros64(Range), true
    }
    //获取低位有多少个0,并加上基础值b
    return bits.TrailingZeros64(t) + int(b), false
}

//检查是否是合法值
func checkNumber(b int8, Range uint64) bool {
    return Range>>b&1 != 0
}

func (s *schedulerCron) NextTime(t time.Time) time.Time {
    year, month, day := t.Date()
    hour, minu, sec := t.Clock()
    var fixed bool

    //修正原始时间
    if !checkNumber(int8(sec), s.second) {
        sec1, _ := getNextNumber(int8(sec+1), s.second)
        if sec1 < sec {
            fixed = true
            minu++
        }
        sec = sec1
    }
    if !checkNumber(int8(minu), s.minute) {
        minu1, _ := getNextNumber(int8(minu+1), s.minute)
        if minu1 < minu {
            fixed = true
            hour++
            sec, _ = getNextNumber(int8(0), s.second)
        }
        minu = minu1
    }
    if !checkNumber(int8(hour), uint64(s.hour)) {
        hour1, _ := getNextNumber(int8(hour+1), uint64(s.hour))
        if hour1 < hour {
            fixed = true
            day++
            minu, _ = getNextNumber(int8(0), s.minute)
            sec, _ = getNextNumber(int8(0), s.second)
        }
        hour = hour1
    }
    if !checkNumber(int8(day), uint64(s.day)) {
        day1, _ := getNextNumber(int8(day+1), uint64(s.day))
        if day1 < day {
            fixed = true
            month++
            hour, _ = getNextNumber(int8(0), s.second)
            minu, _ = getNextNumber(int8(0), s.minute)
            sec, _ = getNextNumber(int8(0), s.second)
        }
        day = day1
    }
    if !checkNumber(int8(month), uint64(s.month)) {
        month1, _ := getNextNumber(int8(month+1), uint64(s.month))
        if month1 < int(month) {
            fixed = true
            year++
            month = time.Month(month1)
            day, _ = getNextNumber(int8(0), uint64(s.day))
            sec, _ = getNextNumber(int8(0), s.second)
            minu, _ = getNextNumber(int8(0), s.minute)
            hour, _ = getNextNumber(int8(0), s.second)
        }
        month = time.Month(month1)
    }

    if len(s.year) > 0 {
        for _, y := range s.year {
            if y > year {
                year = y
                day = 0
                month = time.Month(1)
                sec, _ = getNextNumber(int8(0), s.second)
                minu, _ = getNextNumber(int8(0), s.minute)
                hour, _ = getNextNumber(int8(0), s.second)
                fixed = true
                break
            }
            if y == year {
                year = y
                break
            }
        }
    }
    if !fixed {
        var circulate bool
        if sec, circulate = getNextNumber(int8(sec+1), s.second); !circulate {
            goto ok
        }
        if minu, circulate = getNextNumber(int8(minu+1), s.minute); !circulate {
            goto ok
        }
        if hour, circulate = getNextNumber(int8(hour+1), uint64(s.hour)); circulate {
            var month1 int
            day, month1, year = s.nextDay(day, int(month), year)
            month = time.Month(month1)
        }
    ok:
    }
    {
        if day == 0 {
            var month1 int
            day, month1, year = s.nextDay(day, int(month), year)
            month = time.Month(month1)
        } else {
            t1 := time.Date(year, month, day, 0, 0, 0, 0, s.local)
            weekday := t1.Weekday() + 1
            if (s.week>>uint8(weekday))&1 != 1 {
                var month1 int
                day, month1, year = s.nextDay(day, int(month), year)
                month = time.Month(month1)
            }
        }
    }

    if year == 0 {
        return time.Time{}
    }
    return time.Date(year, month, day, hour, minu, sec, 0, s.local)
}

func (s *schedulerCron) nextDay(day, month, year int) (int, int, int) {
    circulate := false
    nextYear := false
    for {
        if day, circulate = getNextNumber(int8(day+1), uint64(s.day)); !circulate {
            goto testWeek
        }
        if month, circulate = getNextNumber(int8(month+1), uint64(s.month)); circulate {
            if len(s.year) == 0 {
                //未设置年份,允许进位一次
                if !nextYear {
                    year++
                    day, _ = getNextNumber(int8(0), uint64(s.day))
                    month, _ = getNextNumber(int8(0), uint64(s.month))
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
        t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, s.local)
        weekday := t.Weekday() + 1
        if (s.week>>uint8(weekday))&1 == 1 {
            break
        }
    }
    return day, month, year
}

// ParseCron 解析cron表达式,支持5~7位cron表达式,支持设置时区 TZ=Asia/Shanghai 10 10 * * * *
// 每一位含义为 秒(可选) 分 时 每月第几天 月 每周第几天 年(可选),6位时为年不设置,day of week 0和7都为周日
// 内置表达式@yearly,@annually,@monthly,@weekly,@daily,@midnight,@hourly,@every 1h1s,@random 1m 1h
// 其中@random [最低时长] 最高时长 中最低时长默认为1s
func ParseCron(cron string) (CustomSchedulerTime, error) {
    tz := time.Local
    //时区
    if strings.HasPrefix(cron, "CRON_TZ=") || strings.HasPrefix(cron, "TZ=") {
        tzStr := cron[:strings.Index(cron, " ")]
        tzStr = strings.Split(tzStr, "TZ=")[1]
        var err error
        tz, err = time.LoadLocation(tzStr)
        if err != nil {
            return nil, errors.New("指定时区错误:" + err.Error() + ",你的表达式为:" + cron)
        }
        cron = cron[strings.Index(cron, " ")+1:]
    }
    if cron[0] == '@' {
        if strings.HasPrefix(cron, "@every ") {
            cron = cron[7:]
            task, err := time.ParseDuration(cron)
            if err != nil {
                return nil, joinErrs(errors.New("时间间隔格式不正确,支持格式为:@every 1h1s,你的表达式为:"+cron), err)
            }
            if task < time.Second {
                return nil, errors.New("时间间隔最少为1s,你的表达式为:" + cron)
            }
            t := everyTask(task)
            return &t, nil
        }
        if strings.HasPrefix(cron, "@random ") {
            cron = cron[7:]
            split := strings.Split(cron, " ")
            min := time.Second
            var err error
            maxStr := split[0]
            if len(split) > 1 {
                maxStr = split[1]
                min, err = time.ParseDuration(split[0])
                if err != nil {
                    return nil, joinErrs(errors.New("时间间隔格式不正确,支持格式为:@random 1m 1h1s,你的表达式为:"+cron), err)
                }
                if min < time.Second {
                    return nil, errors.New("时间间隔最少为1s,你的表达式为:" + cron)
                }
            }
            max, err := time.ParseDuration(maxStr)
            if err != nil {
                return nil, joinErrs(errors.New("时间间隔格式不正确,支持格式为:@random 1m 1h1s,你的表达式为:"+cron), err)
            }
            if max < min {
                max, min = min, max
            }
            if min < time.Second {
                return nil, errors.New("时间间隔最少为1s,你的表达式为:" + cron)
            }
            return &randomTask{
                offset: min,
                rand:   int64(max - min),
            }, nil
        }
        cron1 := fixedExpressions(cron)
        if cron1 == "" {
            return nil, errors.New("不支持的表达式:" + cron)
        }
        cron = cron1
    }
    items := strings.Split(cron, " ")
    if len(items) < 5 {
        return nil, errors.New("cron表达式错误,必须最少包含5个部分,秒(可选),分,时,日,月,周,年(可选),当前为:'" + cron + "'")
    }
    if len(items) > 7 {
        return nil, errors.New("cron表达式错误,最多包含7个部分,秒,分,时,日,月,周,年(可选),当前为:'" + cron + "'")
    }
    if len(items) == 5 {
        items = append([]string{"0"}, items...)
    }
    s := &schedulerCron{
        corn:  cron,
        local: tz,
    }
    for i, item := range items {
        //年
        if i == 6 {
            if item == "?" || item == "*" {
                continue
            } else if strings.Contains(item, "/") {
                start, end, step, err := parseStep(item, 2000, 2099)
                if err != nil {
                    return nil, joinErrs(errors.New("解析错误,"+item), err)
                }
                for i := start; i <= end; i += step {
                    s.year = append(s.year, int(i))
                }
            } else if strings.Contains(item, "-") {
                start, end, err := parseRange(item, 2000, 2099)
                if err != nil {
                    return nil, joinErrs(errors.New("解析错误,"+item), err)
                }
                for i := start; i <= end; i++ {
                    s.year = append(s.year, int(i))
                }
            } else {
                //手动指定允许指定到9999年
                r, err := parseSingle(item, 0, 9999)
                if err != nil {
                    return nil, joinErrs(errors.New("解析错误,"+item), err)
                }
                for _, i := range r {
                    s.year = append(s.year, int(i))
                }
            }
            continue
        }
        //周
        if i == 5 {
            if item == "?" || item == "*" {
                v, _ := parseCronItem("*", 1, 7)
                s.week = uint8(v)
                continue
            }
            if strings.Contains(item, "#") {
                return nil, errors.New("解析错误,不支持#语法,你的表达式为:" + cron)
            }
            if strings.Contains(item, "L") {
                return nil, errors.New("解析错误,不支持L语法,你的表达式为:" + cron)
            }
            if strings.Contains(item, "C") {
                return nil, errors.New("解析错误,不支持C语法,你的表达式为:" + cron)
            }
            v, err := parseCronItem(item, cronRules[i][0], cronRules[i][1])
            if err != nil {
                //处理英文
                item1 := item
                for k, v1 := range weekMap {
                    item1 = strings.ReplaceAll(item1, k, v1)
                }
                v, err = parseCronItem(item1, cronRules[i][0], cronRules[i][1])
                if err != nil {
                    return nil, joinErrs(errors.New("解析错误,"+item), err)
                }
            }
            //非*或者?号,则与day of month互斥
            if v != math.MaxUint8>>1<<1 && s.day != math.MaxUint32>>1<<1 {
                return nil, errors.New("解析错误,周几和每月第几天不能同时设置,你的表达式为:" + cron)
            }
            //0替换为7
            if v&1 != 0 {
                v = (v & ^uint64(1)) | 1<<7
            }
            s.week = uint8(v)
            continue
        }
        v, err := parseCronItem(item, cronRules[i][0], cronRules[i][1])
        if err != nil {
            //月
            if i == 4 {
                item1 := item
                for k, v1 := range monthMap {
                    item1 = strings.ReplaceAll(item1, k, v1)
                }
                v, err = parseCronItem(item1, cronRules[i][0], cronRules[i][1])
                if err != nil {
                    return nil, joinErrs(errors.New("解析错误,"+item), err)
                }
            } else {
                return nil, joinErrs(errors.New("解析错误,"+item), err)
            }
        }
        switch i {
        case 0:
            s.second = v
        case 1:
            s.minute = v
        case 2:
            s.hour = uint32(v)
        case 3:
            s.day = uint32(v)
        case 4:
            s.month = uint16(v)
        default:
            panic("不可能出现的情况")
        }
    }
    return s, nil
}

func fixedExpressions(cron string) string {
    switch cron {
    case "@yearly", "@annually":
        return "0 0 0 1 1 *"
    case "@monthly":
        return "0 0 0 1 * *"
    case "@weekly":
        return "0 0 0 * * 0"
    case "@daily", "@midnight":
        return "0 0 0 * * *"
    case "@hourly":
        return "0 0 * * * *"
    default:
        return ""
    }
}
