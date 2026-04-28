package helper

import (
    "fmt"
    "math"
    "strconv"
    "sync"
    "testing"
    "time"
)

// ==================== 辅助函数 ====================

// buildBitmask 构建 [from, to] 范围内按步长 step 设置的位掩码
func buildBitmask(from, to, step uint64) uint64 {
    expect := uint64(0)
    for j := from; j <= to; j += step {
        expect |= 1 << j
    }
    return expect
}

// assertNextTime 断言 cron 表达式从 start 开始的 NextTime 序列
func assertNextTime(t *testing.T, cron string, start time.Time, expects ...time.Time) {
    t.Helper()
    c, err := ParseCron(cron)
    if err != nil {
        t.Fatalf("ParseCron(%q) error: %v", cron, err)
    }
    cur := start
    for i, expect := range expects {
        next := c.NextTime(cur)
        if next != expect {
            t.Errorf("[%d] expected: %v, got: %v", i, expect.Format(DateTimeLayout), next.Format(DateTimeLayout))
        }
        cur = next
    }
}

// assertParseError 断言 ParseCron 应该返回错误
func assertParseError(t *testing.T, cron string) {
    t.Helper()
    _, err := ParseCron(cron)
    if err == nil {
        t.Errorf("ParseCron(%q): expected error, got nil", cron)
    }
}

// ==================== parseCronItem 步长解析测试 ====================

func Test_ParseCronItem_WithStep(t *testing.T) {
    cases := []struct {
        name   string
        input  string
        min    uint64
        max    uint64
        expect uint64
        err    bool
    }{
        {"*/1", "*/1", 0, 59, buildBitmask(0, 59, 1), false},
        {"*/5", "*/5", 0, 59, buildBitmask(0, 59, 5), false},
        {"1-20/5", "1-20/5", 0, 59, buildBitmask(1, 20, 5), false},
        {"20-20/5", "20-20/5", 0, 59, buildBitmask(20, 20, 5), false},
        {"*/0 步长为0", "*/0", 0, 59, 0, true},
        {"*/-1 负数", "*/-1", 0, 59, 0, true},
        {"*/60 超出范围", "*/60", 0, 59, 0, true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := parseCronItem(tc.input, tc.min, tc.max)
            if tc.err {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                return
            }
            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            if got != tc.expect {
                t.Errorf("expected: %b, got: %b", tc.expect, got)
            }
        })
    }
    // 全量遍历测试：*/x 对所有合法步长
    t.Run("全量步长遍历", func(t *testing.T) {
        for x := 1; x < 60; x++ {
            got, err := parseCronItem("*/"+strconv.Itoa(x), 0, 59)
            if err != nil {
                t.Errorf("*/%d unexpected error: %v", x, err)
                continue
            }
            expect := buildBitmask(0, 59, uint64(x))
            if got != expect {
                t.Errorf("*/%d expected: %b, got: %b", x, expect, got)
            }
        }
    })
}

// ==================== parseCronItem 范围解析测试 ====================

func Test_ParseCronItem_WithRange(t *testing.T) {
    cases := []struct {
        name   string
        input  string
        min    uint64
        max    uint64
        expect uint64
        err    bool
    }{
        {"10-20", "10-20", 0, 59, buildBitmask(10, 20, 1), false},
        {"60-6 开始大于结束", "60-6", 0, 59, 0, true},
        {"6-99 结束超出范围", "6-99", 0, 59, 0, true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := parseCronItem(tc.input, tc.min, tc.max)
            if tc.err {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                return
            }
            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            if got != tc.expect {
                t.Errorf("expected: %b, got: %b", tc.expect, got)
            }
        })
    }
    // 全量遍历测试：x-y 对所有合法范围
    t.Run("全量范围遍历", func(t *testing.T) {
        for x := 0; x < 60; x++ {
            for y := x; y < 60; y++ {
                input := strconv.Itoa(x) + "-" + strconv.Itoa(y)
                got, err := parseCronItem(input, 0, 59)
                if err != nil {
                    t.Errorf("%s unexpected error: %v", input, err)
                    continue
                }
                expect := buildBitmask(uint64(x), uint64(y), 1)
                if got != expect {
                    t.Errorf("%s expected: %b, got: %b", input, expect, got)
                }
            }
        }
    })
}

// ==================== parseCronItem 单值解析测试 ====================

func Test_ParseCronItem_WithSingle(t *testing.T) {
    cases := []struct {
        name   string
        input  string
        min    uint64
        max    uint64
        expect uint64
        err    bool
    }{
        {"0", "0", 0, 59, 1, false},
        {"10", "10", 0, 59, 1 << 10, false},
        {"60 超出范围", "60", 0, 59, 0, true},
        {"-1 负数", "-1", 0, 59, 0, true},
        {"1,2,3 多值", "1,2,3", 0, 59, 0b1110, false},
        {"1,11,47,59 多值", "1,11,47,59", 0, 59, (1 << 1) | (1 << 11) | (1 << 47) | (1 << 59), false},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := parseCronItem(tc.input, tc.min, tc.max)
            if tc.err {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                return
            }
            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            if got != tc.expect {
                t.Errorf("expected: %b, got: %b", tc.expect, got)
            }
        })
    }
    // 全量遍历测试：所有合法单值
    t.Run("全量单值遍历", func(t *testing.T) {
        for x := 0; x < 60; x++ {
            got, err := parseCronItem(strconv.Itoa(x), 0, 59)
            if err != nil {
                t.Errorf("%d unexpected error: %v", x, err)
                continue
            }
            if got != 1<<x {
                t.Errorf("%d expected: %b, got: %b", x, 1<<x, got)
            }
        }
    })
}

// ==================== parseCronItem 通配符解析测试 ====================

func Test_ParseCronItem_WithBlurRange(t *testing.T) {
    // 各种典型范围的通配符测试
    rangeCases := []struct {
        name string
        min  uint64
        max  uint64
    }{
        {"秒 0-59", 0, 59},
        {"日 1-31", 1, 31},
        {"时 0-23", 0, 23},
        {"月 1-12", 1, 12},
    }
    for _, rc := range rangeCases {
        t.Run(rc.name, func(t *testing.T) {
            got, err := parseCronItem("*", rc.min, rc.max)
            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            expect := buildBitmask(rc.min, rc.max, 1)
            if got != expect {
                t.Errorf("expected:\n %b, got:\n %b", expect, got)
            }
        })
    }
    // 全量遍历测试：* 对所有 min<=max 的范围
    t.Run("全量通配符范围遍历", func(t *testing.T) {
        for a := 0; a < 60; a++ {
            for b := a; b < 60; b++ {
                got, err := parseCronItem("*", uint64(a), uint64(b))
                if err != nil {
                    t.Errorf("* (%d-%d) unexpected error: %v", a, b, err)
                    continue
                }
                expect := buildBitmask(uint64(a), uint64(b), 1)
                if got != expect {
                    t.Errorf("* (%d-%d) expected:\n %b, got:\n %b", a, b, expect, got)
                }
            }
        }
    })
}

// ==================== ParseCron 完整解析测试 ====================

func Test_ParseCron(t *testing.T) {
    t.Run("6位表达式", func(t *testing.T) {
        v1, err := ParseCron("*/5 * * * * ?")
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        v := v1.(*cronTimer)
        second := buildBitmask(0, 59, 5)
        if v.second != second {
            t.Errorf("second expected:\n %b, got:\n %b", second, v.second)
        }
        if v.minute != math.MaxUint64>>4 {
            t.Errorf("minute expected:\n %b, got:\n %b", math.MaxUint64>>4, v.minute)
        }
        if v.hour != (math.MaxUint32 >> (32 - 24)) {
            t.Errorf("hour expected:\n %b, got:\n %b", math.MaxUint32>>(32-24), v.hour)
        }
        if v.day != (math.MaxUint32>>(32-31))<<1 {
            t.Errorf("day expected:\n %b, got:\n %b", (math.MaxUint32>>(32-31))<<1, v.day)
        }
        if v.month != (math.MaxUint16>>(16-12))<<1 {
            t.Errorf("month expected:\n %b, got:\n %b", (math.MaxUint16>>(16-12))<<1, v.month)
        }
        if v.week != math.MaxUint8>>1 {
            t.Errorf("week expected:\n %b, got:\n %b", math.MaxUint8>>1, v.week)
        }
    })
    t.Run("7位表达式", func(t *testing.T) {
        v1, err := ParseCron("0 0 0 * * ? *")
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        v := v1.(*cronTimer)
        if v.second != 1 {
            t.Errorf("second expected: %b, got: %b", 1, v.second)
        }
        if v.minute != 1 {
            t.Errorf("minute expected: %b, got: %b", 1, v.minute)
        }
        if v.hour != 1 {
            t.Errorf("hour expected: %b, got: %b", 1, v.hour)
        }
        if v.day != (math.MaxUint32>>(32-31))<<1 {
            t.Errorf("day expected:\n %b, got:\n %b", (math.MaxUint32>>(32-31))<<1, v.day)
        }
        if v.month != (math.MaxUint16>>(16-12))<<1 {
            t.Errorf("month expected:\n %b, got:\n %b", (math.MaxUint16>>(16-12))<<1, v.month)
        }
        if v.week != math.MaxUint8>>1 {
            t.Errorf("week expected:\n %b, got:\n %b", math.MaxUint8>>1, v.week)
        }
    })
    // 固定表达式和 @every 系列解析应成功
    validCases := []string{"@yearly", "@every 1s", "@every 1h", "@every 1h1m1s1ms"}
    for _, cron := range validCases {
        t.Run(cron, func(t *testing.T) {
            if _, err := ParseCron(cron); err != nil {
                t.Errorf("ParseCron(%q) unexpected error: %v", cron, err)
            }
        })
    }
}

// ==================== ParseCron 无效表达式错误测试 ====================

func Test_ParseCron_WithInvalidExpression(t *testing.T) {
    cases := []struct {
        name string
        cron string
    }{
        {"纯字符串", "abc"},
        {"空字符串", ""},
        {"步长为0", "*/0 * * * * ?"},
        {"步长为负数", "*/-1 * * * * ?"},
        {"步长超出范围", "*/60 * * * * ?"},
        {"范围超出上限", "0-60 * * * * ?"},
        {"范围结束超出", "6-99 * * * * ?"},
        {"单值超出范围", "60 * * * * ?"},
        {"单值为负数", "-1 * * * * ?"},
        {"位数不足4位", "* * * ?"},
        {"位数超过8位", "* * * * * * * *"},
        {"同时指定日和周", "0 0 0 1 6 5"},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            assertParseError(t, tc.cron)
        })
    }
}

func Test_ParseCron_WithRandomEqualDuration(t *testing.T) {
    c, err := ParseCron("@random 1s")
    if err != nil {
        t.Fatalf("ParseCron(@random 1s) unexpected error: %v", err)
    }
    start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
    next := c.NextTime(start)
    expect := start.Add(time.Second)
    if next != expect {
        t.Fatalf("NextTime expected %v, got %v", expect, next)
    }
}

// ==================== NextTime 计算测试 ====================

func Test_Cron_nextTime(t *testing.T) {
    // 基本 NextTime 单步断言
    t.Run("步长_秒", func(t *testing.T) {
        assertNextTime(t, "*/5 * * * * ?",
            time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
            time.Date(2020, 1, 1, 0, 0, 5, 0, time.Local))
    })
    t.Run("步长_分", func(t *testing.T) {
        assertNextTime(t, "*/5 * * * ?",
            time.Date(2020, 1, 1, 0, 1, 1, 0, time.Local),
            time.Date(2020, 1, 1, 0, 5, 0, 0, time.Local))
    })
    t.Run("步长_日", func(t *testing.T) {
        assertNextTime(t, "0 0 0 */5 * ?",
            time.Date(2020, 1, 1, 1, 1, 1, 0, time.Local),
            time.Date(2020, 1, 6, 0, 0, 0, 0, time.Local))
    })
    // 各粒度递进测试
    granCases := []struct {
        name   string
        cron   string
        start  time.Time
        expect time.Time
    }{
        {"每秒", "* * * * * ? *",
            time.Date(2020, 1, 1, 1, 1, 1, 0, time.Local),
            time.Date(2020, 1, 1, 1, 1, 2, 0, time.Local)},
        {"每分", "0 * * * * ? *",
            time.Date(2020, 1, 1, 1, 1, 1, 0, time.Local),
            time.Date(2020, 1, 1, 1, 2, 0, 0, time.Local)},
        {"每时", "0 0 * * * ? *",
            time.Date(2020, 1, 1, 1, 1, 1, 0, time.Local),
            time.Date(2020, 1, 1, 2, 0, 0, 0, time.Local)},
        {"每日", "0 0 0 * * ? *",
            time.Date(2020, 1, 1, 1, 1, 1, 0, time.Local),
            time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local)},
        {"每月", "0 0 0 1 * ? *",
            time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
            time.Date(2020, 2, 1, 0, 0, 0, 0, time.Local)},
        {"每年", "@yearly",
            time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
            time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)},
        {"每秒_every", "@every 1s",
            time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
            time.Date(2020, 1, 1, 0, 0, 1, 0, time.Local)},
        {"周3", "0 0 0 * * 3",
            time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
            time.Date(2020, 1, 8, 0, 0, 0, 0, time.Local)},
        {"周1", "0 0 0 * * 1",
            time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
            time.Date(2020, 1, 6, 0, 0, 0, 0, time.Local)},
    }
    for _, tc := range granCases {
        t.Run(tc.name, func(t *testing.T) {
            assertNextTime(t, tc.cron, tc.start, tc.expect)
        })
    }

    // 多步递进测试
    t.Run("周3连续两周", func(t *testing.T) {
        assertNextTime(t, "0 0 0 * * 3",
            time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
            time.Date(2020, 1, 8, 0, 0, 0, 0, time.Local),
            time.Date(2020, 1, 15, 0, 0, 0, 0, time.Local))
    })
    t.Run("月+周组合_3月周五", func(t *testing.T) {
        assertNextTime(t, "0 0 0 * 3 5",
            time.Date(2020, 1, 5, 1, 1, 1, 0, time.Local),
            time.Date(2020, 3, 6, 0, 0, 0, 0, time.Local))
    })
    t.Run("月+周组合_6月周日", func(t *testing.T) {
        assertNextTime(t, "0 0 0 * 6 7",
            time.Date(2020, 3, 5, 1, 1, 1, 1, time.Local),
            time.Date(2020, 6, 7, 0, 0, 0, 0, time.Local))
    })
    t.Run("跨年_月+周", func(t *testing.T) {
        assertNextTime(t, "0 0 0 * 6 7",
            time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local),
            time.Date(2021, 6, 6, 0, 0, 0, 0, time.Local))
    })
    t.Run("指定年份", func(t *testing.T) {
        c, err := ParseCron("0 0 0 * * * 2024")
        if err != nil {
            t.Fatal(err)
        }
        expect := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
        for _, start := range []time.Time{
            time.Date(2021, 12, 5, 0, 0, 0, 0, time.Local),
            time.Date(2022, 12, 5, 0, 0, 0, 0, time.Local),
            time.Date(2023, 12, 5, 0, 0, 0, 0, time.Local),
        } {
            next := c.NextTime(start)
            if next != expect {
                t.Errorf("from %v: expected %v, got %v",
                    start.Format(DateTimeLayout), expect.Format(DateTimeLayout), next.Format(DateTimeLayout))
            }
        }
    })
    t.Run("week_0与7等价", func(t *testing.T) {
        c0, _ := ParseCron("0 0 0 * * 0")
        c1, _ := ParseCron("0 0 0 * * 7")
        if c0.(*cronTimer).week != c1.(*cronTimer).week {
            t.Error("week 0,1,2 != week 7,1,2", c0.(*cronTimer).week, c1.(*cronTimer).week)
        }
    })
}

// ==================== 跨天/跨月/跨年 hour 重置测试 ====================

func Test_Cron_nextTime_hourResetOnRollover(t *testing.T) {
    cases := []struct {
        name   string
        cron   string
        start  time.Time
        expect time.Time
    }{
        {
            // 跨天后 hour 应从 hour 位图取最小值，而不是错误复用 second 位图
            name:   "跨天后 hour 取 hour 字段最小值",
            cron:   "15 20 10 * * ?",
            start:  time.Date(2020, 1, 1, 10, 20, 15, 0, time.Local),
            expect: time.Date(2020, 1, 2, 10, 20, 15, 0, time.Local),
        },
        {
            // 跨月时同样要将 hour 重置为 hour 位图中的最小值
            name:   "跨月后 hour 取 hour 字段最小值",
            cron:   "15 20 10 1 * ?",
            start:  time.Date(2020, 1, 1, 10, 20, 15, 0, time.Local),
            expect: time.Date(2020, 2, 1, 10, 20, 15, 0, time.Local),
        },
        {
            // 跨年时也应保持从 hour 位图取最小值
            name:   "跨年后 hour 取 hour 字段最小值",
            cron:   "15 20 10 1 1 ?",
            start:  time.Date(2020, 1, 1, 10, 20, 15, 0, time.Local),
            expect: time.Date(2021, 1, 1, 10, 20, 15, 0, time.Local),
        },
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            assertNextTime(t, tc.cron, tc.start, tc.expect)
        })
    }
}

// ==================== 星期相关测试 ====================

// 验证每个 cron 星期值都能正确匹配对应的 Go Weekday
func Test_Cron_Week_AllIndividualDays(t *testing.T) {
    cases := []struct {
        cronDay int
        weekday time.Weekday
    }{
        {0, time.Sunday},
        {1, time.Monday},
        {2, time.Tuesday},
        {3, time.Wednesday},
        {4, time.Thursday},
        {5, time.Friday},
        {6, time.Saturday},
        {7, time.Sunday},
    }
    start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
    for _, tc := range cases {
        t.Run(fmt.Sprintf("cron_%d_%s", tc.cronDay, tc.weekday), func(t *testing.T) {
            c, err := ParseCron(fmt.Sprintf("0 0 0 * * %d", tc.cronDay))
            if err != nil {
                t.Fatal(err)
            }
            next := c.NextTime(start)
            if next.Weekday() != tc.weekday {
                t.Errorf("cron day %d: expected weekday %s, got %s (date: %v)",
                    tc.cronDay, tc.weekday, next.Weekday(), next.Format("2006-01-02"))
            }
        })
    }
}

// 验证 week 0 和 7 产生相同的内部表示
func Test_Cron_Week_0Equals7(t *testing.T) {
    c0, _ := ParseCron("0 0 0 * * 0")
    c7, _ := ParseCron("0 0 0 * * 7")
    w0 := c0.(*cronTimer).week
    w7 := c7.(*cronTimer).week
    if w0 != w7 {
        t.Errorf("week 0 (%08b) != week 7 (%08b)", w0, w7)
    }
}

// 验证 cron "*" 星期匹配（应该匹配所有天）
func Test_Cron_Week_AllDays(t *testing.T) {
    assertNextTime(t, "0 0 0 * * *",
        time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
        time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local),
        time.Date(2020, 1, 3, 0, 0, 0, 0, time.Local))
}

// 验证文本星期名称解析正确
func Test_Cron_Week_TextNames(t *testing.T) {
    cases := []struct {
        name    string
        cronDay string
        weekday time.Weekday
    }{
        {"SUN", "SUN", time.Sunday},
        {"MON", "MON", time.Monday},
        {"TUE", "TUE", time.Tuesday},
        {"WED", "WED", time.Wednesday},
        {"THU", "THU", time.Thursday},
        {"FRI", "FRI", time.Friday},
        {"SAT", "SAT", time.Saturday},
    }
    start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            c, err := ParseCron(fmt.Sprintf("0 0 0 * * %s", tc.cronDay))
            if err != nil {
                t.Fatal(err)
            }
            next := c.NextTime(start)
            if next.Weekday() != tc.weekday {
                t.Errorf("%s: expected weekday %s, got %s (date: %v)",
                    tc.cronDay, tc.weekday, next.Weekday(), next.Format("2006-01-02"))
            }
        })
    }
}

// 验证 @random 在并发调用下不会 panic
func Test_ParseCron_RandomTaskConcurrent(t *testing.T) {
    c, err := ParseCron("@random 1s 2s")
    if err != nil {
        t.Fatalf("ParseCron(@random 1s 2s) unexpected error: %v", err)
    }

    start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            next := c.NextTime(start)
            // 结果应在 [start+1s, start+2s) 范围内
            if next.Before(start.Add(time.Second)) || !next.Before(start.Add(2*time.Second)) {
                t.Errorf("NextTime out of range: %v", next)
            }
        }()
    }
    wg.Wait()
}

// H-12 回归测试: 当指定年份已全部过期时, NextTime 应返回零值
func Test_Cron_YearExhaustion_ReturnsZeroTime(t *testing.T) {
    // 指定一个明显过去的年份
    c, err := ParseCron("0 0 0 1 1 * 1999")
    if err != nil {
        t.Fatalf("ParseCron unexpected error: %v", err)
    }
    // 从 2020 年开始查找, 1999 年早已过去
    start := time.Date(2020, 6, 15, 12, 0, 0, 0, time.Local)
    next := c.NextTime(start)
    if !next.IsZero() {
        t.Errorf("expected zero time for exhausted year, got: %v", next)
    }
}

// 补充: 指定年份范围全部过期
func Test_Cron_YearRangeExhaustion_ReturnsZeroTime(t *testing.T) {
    c, err := ParseCron("0 0 0 1 1 * 2000-2005")
    if err != nil {
        t.Fatalf("ParseCron unexpected error: %v", err)
    }
    start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
    next := c.NextTime(start)
    if !next.IsZero() {
        t.Errorf("expected zero time for exhausted year range, got: %v", next)
    }
}

// 补充: 指定年份列表全部过期
func Test_Cron_YearListExhaustion_ReturnsZeroTime(t *testing.T) {
    c, err := ParseCron("0 0 0 1 6 * 2010,2015,2018")
    if err != nil {
        t.Fatalf("ParseCron unexpected error: %v", err)
    }
    start := time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)
    next := c.NextTime(start)
    if !next.IsZero() {
        t.Errorf("expected zero time for exhausted year list, got: %v", next)
    }
}

// H-14 回归测试: 年份列表无序输入应被正确排序
func Test_Cron_YearListUnsorted(t *testing.T) {
    // 无序年份列表: 2025,2020,2023
    c, err := ParseCron("0 0 0 1 1 * 2025,2020,2023")
    if err != nil {
        t.Fatalf("ParseCron unexpected error: %v", err)
    }
    start := time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)
    next := c.NextTime(start)
    // 应该返回 2020 年, 而非跳过到 2023 或 2025
    want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
    if !next.Equal(want) {
        t.Errorf("expected %v for unsorted year list, got: %v", want, next)
    }
}

// H-13 回归测试: 月份天数校验, 避免 time.Date 溢出到下月
func Test_Cron_InvalidDayInMonth(t *testing.T) {
    // 每月 31 日, 但 2 月没有 31 日
    c, err := ParseCron("0 0 0 31 * *")
    if err != nil {
        t.Fatalf("ParseCron unexpected error: %v", err)
    }
    // 从 2 月 1 日开始, 应该返回 3 月 31 日(跳过 2 月)
    start := time.Date(2020, 2, 1, 0, 0, 0, 0, time.Local)
    next := c.NextTime(start)
    want := time.Date(2020, 3, 31, 0, 0, 0, 0, time.Local)
    if !next.Equal(want) {
        t.Errorf("expected %v (skipping Feb), got: %v", want, next)
    }
}

// H-28 回归测试: 复合表达式解析, 支持逗号分隔的范围和步长
func Test_Cron_CompositeExpression(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name string
        cron string
        // 验证解析成功且能产生有效时间
    }{
        // 逗号分隔范围: "1-5,10-15" 应解析为 [1,2,3,4,5,10,11,12,13,14,15]
        {
            name: "minute_composite_range",
            cron: "0 1-5,10-15 0 * * *",
        },
        // 逗号分隔单值: "1,3,5"
        {
            name: "minute_composite_single",
            cron: "0 1,3,5 0 * * *",
        },
        // 逗号分隔混合: "1-5/2,10" -> [1,3,5,10]
        {
            name: "minute_composite_mixed",
            cron: "0 1-5/2,10 0 * * *",
        },
        // 秒级复合表达式
        {
            name: "second_composite",
            cron: "0,30 0 * * * *",
        },
        // 小时复合表达式
        {
            name: "hour_composite",
            cron: "0 0 9-17,20 * * *",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c, err := ParseCron(tt.cron)
            if err != nil {
                t.Fatalf("ParseCron(%q) error: %v", tt.cron, err)
            }
            // 简单验证: 解析成功且能产生有效时间
            now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
            next := c.NextTime(now)
            if next.IsZero() {
                t.Errorf("NextTime should not be zero for %q", tt.cron)
            }
        })
    }
}

// TestWeekdayOf_Sakamoto 验证 Sakamoto 算法与 time.Date 结果一致
func TestWeekdayOf_Sakamoto(t *testing.T) {
    t.Parallel()
    tests := []struct {
        y int
        m time.Month
        d int
    }{
        {2000, 1, 1},   // 周六
        {2024, 1, 1},   // 周一
        {2024, 2, 29},  // 闰年周四
        {2023, 12, 25}, // 周一
        {1999, 7, 4},   // 周日
        {2025, 3, 30},  // 周日
        {1, 1, 1},      // 极远过去
        {9999, 12, 31}, // 极远未来
    }
    for _, tt := range tests {
        tt := tt
        t.Run(fmt.Sprintf("%d-%02d-%02d", tt.y, tt.m, tt.d), func(t *testing.T) {
            t.Parallel()
            expected := time.Date(tt.y, tt.m, tt.d, 0, 0, 0, 0, time.UTC).Weekday()
            got := weekdayOf(tt.y, tt.m, tt.d)
            if got != expected {
                t.Fatalf("weekdayOf(%d, %d, %d) = %v, want %v", tt.y, tt.m, tt.d, got, expected)
            }
        })
    }
}
