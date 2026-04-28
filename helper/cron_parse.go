package helper

import (
    "errors"
    "math"
    "strconv"
    "strings"
)

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
        "SUN": "0",
        "MON": "1",
        "TUE": "2",
        "WED": "3",
        "THU": "4",
        "FRI": "5",
        "SAT": "6",
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

// parseCronValues 解析 cron 表达式项, 返回所有合法值列表
// 通配符 (* 或 ?) 返回 nil, nil, 由调用方处理
// 支持逗号分隔的复合表达式, 如 "1-5,10-15", "1,3,5", "1-5/2,10"
func parseCronValues(item string, min, max uint64) ([]uint64, error) {
    // 逗号分隔的复合表达式先拆分再合并
    if strings.Contains(item, ",") {
        parts := strings.Split(item, ",")
        var result []uint64
        for _, part := range parts {
            vals, err := parseCronValues(part, min, max)
            if err != nil {
                return nil, err
            }
            result = append(result, vals...)
        }
        return result, nil
    }
    if strings.Contains(item, "/") {
        start, end, step, err := parseStep(item, min, max)
        if err != nil {
            return nil, err
        }
        vals := make([]uint64, 0, (end-start)/step+1)
        for i := start; i <= end; i += step {
            vals = append(vals, i)
        }
        return vals, nil
    }
    if strings.Contains(item, "-") {
        start, end, err := parseRange(item, min, max)
        if err != nil {
            return nil, err
        }
        vals := make([]uint64, 0, end-start+1)
        for i := start; i <= end; i++ {
            vals = append(vals, uint64(i))
        }
        return vals, nil
    }
    if item == "*" || item == "?" {
        return nil, nil
    }
    return parseSingle(item, min, max)
}

//解析cron表达式中非年份部分,星期和月份部分需要特殊处理
func parseCronItem(cronItem string, min, max uint64) (s uint64, e error) {
    // 处理通配符, 转为完整位掩码
    if cronItem == "*" || cronItem == "?" {
        return parseBlurRange(min, max), nil
    }
    vals, err := parseCronValues(cronItem, min, max)
    if err != nil {
        return 0, err
    }
    for _, i := range vals {
        s |= 1 << i
    }
    return
}

func parseSingle(item string, min uint64, max uint64) (r []uint64, e error) {
    // parseCronValues 已按逗号拆分并递归调用, 此处 item 不含逗号
    i, err := strconv.ParseInt(item, 10, 64)
    if err != nil {
        e = joinErrs(errors.New(item+"无法解析为数字"), err)
        return
    }
    if i < int64(min) || i > int64(max) {
        e = errors.New("范围错误,范围必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(i)) + "超出范围")
        return
    }
    r = append(r, uint64(i))
    return
}

func parseBlurRange(min, max uint64) uint64 {
    // 防御性检查: max 超过 63 会导致移位下溢
    if max > 63 {
        max = 63
    }
    s := uint64(math.MaxUint64)
    s = s >> (63 - max)
    s = s >> min << min
    return s
}

// validateRange 校验范围合法性: start <= end, min <= start, end <= max
func validateRange(start, end, min, max int64) error {
    if start > end {
        return errors.New("范围错误,开始不能大于结束," + strconv.Itoa(int(start)) + ">" + strconv.Itoa(int(end)))
    }
    if start < min {
        return errors.New("范围错误,开始必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(start)) + "超出范围")
    }
    if end > max {
        return errors.New("范围错误,结束必须在" + strconv.Itoa(int(min)) + "-" + strconv.Itoa(int(max)) + "之间," + strconv.Itoa(int(end)) + "超出范围")
    }
    return nil
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
    if e = validateRange(int64(start), int64(end), int64(min), int64(max)); e != nil {
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
    if e = validateRange(start, end, int64(min), int64(max)); e != nil {
        return
    }
    return
}
