package helper

import (
    "errors"
    "math"
    "math/rand"
    "sort"
    "strings"
    "sync"
    "time"
)

// 并发安全的全局随机数生成器，兼容 Go 1.18
var (
    globalRandMu sync.Mutex
    globalRand   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// safeRandInt63n 并发安全地生成 [0, n) 的随机 int64
func safeRandInt63n(n int64) int64 {
    globalRandMu.Lock()
    v := globalRand.Int63n(n)
    globalRandMu.Unlock()
    return v
}

// CustomSchedulerTime 自定义调度时间接口
type CustomSchedulerTime interface {
    NextTime(time.Time) time.Time
}

type everyTask time.Duration

func (e everyTask) NextTime(t time.Time) time.Time {
    return t.Add(time.Duration(e))
}

type randomTask struct {
    offset time.Duration
    rand   int64
}

func (e *randomTask) NextTime(t time.Time) time.Time {
    return t.Add(time.Duration(safeRandInt63n(e.rand)) + e.offset)
}

// joinError 多错误组合
type joinError struct {
    errs []error
}

func (e *joinError) Error() string {
    var b []byte
    for i, err := range e.errs {
        if i > 0 {
            b = append(b, '\n')
        }
        b = append(b, err.Error()...)
    }
    return string(b)
}

func (e *joinError) Unwrap() []error {
    return e.errs
}

func joinErrs(errs ...error) error {
    n := 0
    for _, err := range errs {
        if err != nil {
            n++
        }
    }
    if n == 0 {
        return nil
    }
    e := &joinError{
        errs: make([]error, 0, n),
    }
    for _, err := range errs {
        if err != nil {
            e.errs = append(e.errs, err)
        }
    }
    return e
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
    if len(cron) == 0 {
        return nil, errors.New("cron表达式不能为空")
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
            cron = strings.TrimSpace(cron[8:])
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
            // max==min 时 rand=0, rand.Int63n(0) 会 panic, 退化为固定间隔
            if max == min {
                t := everyTask(min)
                return &t, nil
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
    s := &cronTimer{
        local: tz,
    }
    for i, item := range items {
        //年: 使用 parseCronValues 统一解析, 避免与 parseCronItem 结构重复
        if i == 6 {
            if item == "?" || item == "*" {
                continue
            }
            vals, err := parseCronValues(item, cronRules[6][0], cronRules[6][1])
            if err != nil {
                return nil, joinErrs(errors.New("解析错误,"+item), err)
            }
            for _, v := range vals {
                s.year = append(s.year, int(v))
            }
            continue
        }
        //周
        if i == 5 {
            if item == "?" || item == "*" {
                v, _ := parseCronItem("*", 0, 6)
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
            if v != parseBlurRange(0, 6) && s.day != math.MaxUint32>>1<<1 {
                return nil, errors.New("解析错误,周几和每月第几天不能同时设置,你的表达式为:" + cron)
            }
            // 7替换为0（Go的Sunday在bit 0）
            if v&(1<<7) != 0 {
                v |= 1
                v &^= 1 << 7
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
    // 年份列表排序, 保证 NextTime 线性遍历能找到最小的大于当前年份的值
    if len(s.year) > 1 {
        sort.Ints(s.year)
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
