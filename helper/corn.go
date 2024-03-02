package helper

import (
    "errors"
    "math"
    "math/bits"
    "strconv"
    "strings"
    "time"
)

var (
    cronRules = [][]uint64{
        {0, 59},
        {0, 59},
        {0, 23},
        {1, 31},
        {1, 12},
        {1, 7},
        {1990, 9999},
    }
    cronFix = map[string]string{
        "@yearly":   "0 0 0 1 1 *",
        "@annually": "0 0 0 1 1 *",
        "@monthly":  "0 0 0 1 * *",
        "@weekly":   "0 0 0 * * 0",
        "@daily":    "0 0 0 * * *",
        "@midnight": "0 0 0 * * *",
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

    day, hour      uint32
    second, minute uint64
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
    var circulate bool

    //1. 检测给的时间是否需要修正

    if len(s.year) > 0 {
        for _, y := range s.year {
            if y > year {
                year = y
                day = 0
                month = time.Month(1)
                sec, _ = getNextNumber(int8(0), s.second)
                minu, _ = getNextNumber(int8(0), s.minute)
                hour, _ = getNextNumber(int8(0), s.second)
                goto noCheck
            }
            if y == year {
                year = y
            }
        }
    }
    if !checkNumber(int8(minu), s.minute) {
        minu, _ = getNextNumber(int8(0), s.minute)
    }
    if !checkNumber(int8(hour), uint64(s.hour)) {
        hour, _ = getNextNumber(int8(0), uint64(s.hour))
    }
    if !checkNumber(int8(day), uint64(s.day)) {
        day1, _ := getNextNumber(int8(0), uint64(s.day))
        if day1 < day {
            month++
        }
        day = day1
    }
    if !checkNumber(int8(month), uint64(s.month)) {
        month1, _ := getNextNumber(int8(month), uint64(s.month))
        if month1 < int(month) {
            year++
            day, _ = getNextNumber(int8(0), uint64(s.day))
            month = time.Month(month1)
            sec, _ = getNextNumber(int8(0), s.second)
            minu, _ = getNextNumber(int8(0), s.minute)
            hour, _ = getNextNumber(int8(0), s.second)
            goto noCheck
        }
        month = time.Month(month1)
    }
    if sec, circulate = getNextNumber(int8(sec+1), s.second); !circulate {
        goto ok
    }
    if minu, circulate = getNextNumber(int8(minu+1), s.minute); !circulate {
        goto ok
    }
    if hour, circulate = getNextNumber(int8(hour+1), s.second); !circulate {
        goto ok
    }
noCheck:
    {
        var month1 int
        day, month1, year = s.nextDay(day, int(month), year)
        month = time.Month(month1)
    }

    if year == 0 {
        return time.Time{}
    }
ok:
    return time.Date(year, month, day, hour, minu, sec, 0, t.Location())
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
        t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
        weekday := t.Weekday() + 1
        if (s.week>>uint8(weekday))&1 == 1 {
            break
        }
    }
    return day, month, year
}

type everyTask time.Duration

func (e everyTask) NextTime(t time.Time) time.Time {
    return t.Add(time.Duration(e))
}

func parseCron(cron string) (CustomSchedulerTime, error) {
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
        cron1, ok := cronFix[cron]
        if !ok {
            return nil, errors.New("不支持的cron表达式:" + cron)
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
        corn: cron,
    }
    for i, item := range items {
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
                    strings.ReplaceAll(item1, k, v1)
                }
                v, err = parseCronItem(item1, cronRules[i][0], cronRules[i][1])
                if err != nil {
                    return nil, joinErrs(errors.New("解析错误,"+item), err)
                }
            }
            //非*或者?号,则与day of month互斥
            if v != math.MaxUint8>>1<<1 && s.day != math.MaxUint32>>1<<1 {
                return nil, errors.New("解析错误,周和日不能同时设置,你的表达式为:" + cron)
            }
            s.week = uint8(v)
            continue
        }
        v, err := parseCronItem(item, cronRules[i][0], cronRules[i][1])
        if err != nil {
            return nil, joinErrs(errors.New("解析错误,"+item), err)
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
