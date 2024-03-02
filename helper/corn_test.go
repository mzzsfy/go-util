package helper

import (
    "math"
    "strconv"
    "testing"
    "time"
)

func Test_ParseCronItem_WithStep(t *testing.T) {
    t.Run("parseCronItem_WithStep_1", func(t *testing.T) {
        i, err := parseCronItem("*/1", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if i != math.MaxUint64>>4 {
            t.Errorf("expected: %b, got: %b", math.MaxUint64>>4, i)
        }
    })
    t.Run("parseCronItem_WithStep_5", func(t *testing.T) {
        i, err := parseCronItem("*/5", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 0; j < 60; j += 5 {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected: %b, got: %b", expect, i)
        }
    })
    t.Run("parseCronItem_WithStep_x", func(t *testing.T) {
        for x := 1; x < 60; x++ {
            i, err := parseCronItem("*/"+strconv.Itoa(x), 0, 59)
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            expect := uint64(0)
            for j := 0; j < 60; j += x {
                expect |= 1 << uint64(j)
            }
            if i != expect {
                t.Errorf("expected: %b, got: %b", expect, i)
            }
        }
    })
    t.Run("parseCronItem_WithStep_0", func(t *testing.T) {
        _, err := parseCronItem("*/0", 0, 59)
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCronItem_WithStep_-1", func(t *testing.T) {
        _, err := parseCronItem("*/-1", 0, 59)
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCronItem_WithStep_60", func(t *testing.T) {
        _, err := parseCronItem("*/60", 0, 59)
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCronItem_WithStep_60", func(t *testing.T) {
        i, err := parseCronItem("1-20/5", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 1; j <= 20; j += 5 {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected: %b, got: %b", expect, i)
        }
    })
    t.Run("parseCronItem_WithStep_60", func(t *testing.T) {
        i, err := parseCronItem("20-20/5", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 20; j <= 20; j += 5 {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected: %b, got: %b", expect, i)
        }
    })
}

func Test_ParseCronItem_WithRange(t *testing.T) {
    t.Run("parseCronItem_WithRange_10-20", func(t *testing.T) {
        i, err := parseCronItem("10-20", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 10; j <= 20; j++ {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected: %b, got: %b", expect, i)
        }
    })
    t.Run("parseCronItem_WithRange_x-y", func(t *testing.T) {
        for x := 0; x < 60; x++ {
            for y := x; y < 60; y++ {
                i, err := parseCronItem(strconv.Itoa(x)+"-"+strconv.Itoa(y), 0, 59)
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                expect := uint64(0)
                for j := x; j <= y; j++ {
                    expect |= 1 << uint64(j)
                }
                if i != expect {
                    t.Errorf("expected: %b, got: %b", expect, i)
                }
            }
        }
    })
    t.Run("parseCronItem_WithRange_0-60", func(t *testing.T) {
        _, err := parseCronItem("60-6", 0, 59)
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCronItem_WithRange_6-99", func(t *testing.T) {
        _, err := parseCronItem("6-99", 0, 59)
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
}

func Test_ParseCronItem_WithSingle(t *testing.T) {
    t.Run("parseCronItem_WithSingle_10", func(t *testing.T) {
        i, err := parseCronItem("10", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if i != 1<<10 {
            t.Errorf("expected: %b, got: %b", 1<<10, i)
        }
    })
    t.Run("parseCronItem_WithSingle_60", func(t *testing.T) {
        _, err := parseCronItem("60", 0, 59)
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCronItem_WithSingle_-1", func(t *testing.T) {
        _, err := parseCronItem("-1", 0, 59)
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCronItem_WithSingle_x", func(t *testing.T) {
        for x := 0; x < 60; x++ {
            i, err := parseCronItem(strconv.Itoa(x), 0, 59)
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if i != 1<<x {
                t.Errorf("expected: %b, got: %b", 1<<x, i)
            }
        }
    })
    t.Run("parseCronItem_WithSingle_0", func(t *testing.T) {
        i, err := parseCronItem("0", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if i != 1 {
            t.Errorf("expected: %b, got: %b", 1, i)
        }
    })
    t.Run("parseCronItem_WithSingle_multiple", func(t *testing.T) {
        i, err := parseCronItem("1,2,3", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if i != 0b1110 {
            t.Errorf("expected: %b, got: %b", 1<<10, i)
        }
    })
    t.Run("parseCronItem_WithSingle_multiple", func(t *testing.T) {
        i, err := parseCronItem("1,11,47,59", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if i != (1<<1)|(1<<11)|(1<<47)|(1<<59) {
            t.Errorf("expected: %b, got: %b", 1<<10, i)
        }
    })
}

func Test_ParseCronItem_WithBlurRange(t *testing.T) {
    t.Run("parseCronItem_BlurRange_0-59", func(t *testing.T) {
        i, err := parseCronItem("*", 0, 59)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 0; j <= 59; j++ {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected:\n %b, got:\n %b", expect, i)
        }
    })
    t.Run("parseCronItem_BlurRange_1-31", func(t *testing.T) {
        i, err := parseCronItem("*", 1, 31)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 1; j <= 31; j++ {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected:\n %b, got:\n %b", expect, i)
        }
    })
    t.Run("parseCronItem_BlurRange_0-23", func(t *testing.T) {
        i, err := parseCronItem("*", 0, 23)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 0; j <= 23; j++ {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected:\n %b, got:\n %b", expect, i)
        }
    })
    t.Run("parseCronItem_BlurRange_1-12", func(t *testing.T) {
        i, err := parseCronItem("*", 1, 12)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        expect := uint64(0)
        for j := 1; j <= 12; j++ {
            expect |= 1 << uint64(j)
        }
        if i != expect {
            t.Errorf("expected:\n %b, got:\n %b", expect, i)
        }
    })
    t.Run("parseCronItem_BlurRange_a-b", func(t *testing.T) {
        for a := 0; a < 60; a++ {
            for b := a; b < 60; b++ {
                i, err := parseCronItem("*", uint64(a), uint64(b))
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                expect := uint64(0)
                for j := a; j <= b; j++ {
                    expect |= 1 << uint64(j)
                }
                if i != expect {
                    t.Errorf("expected:\n %b, got:\n %b", expect, i)
                }
            }
        }
    })
}

func Test_ParseCron(t *testing.T) {
    t.Run("parseCron_len_6", func(t *testing.T) {
        v1, err := parseCron("*/5 * * * * ?")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        v := v1.(*schedulerCron)
        second := uint64(0)
        for j := 0; j < 60; j += 5 {
            second |= 1 << uint64(j)
        }
        if v.second != second {
            t.Errorf("expected:\n %b, got:\n %b", second, v.second)
        }
        if v.minute != math.MaxUint64>>4 {
            t.Errorf("expected:\n %b, got:\n %b", math.MaxUint64>>4, v.minute)
        }
        if v.hour != (math.MaxUint32 >> (32 - 24)) {
            t.Errorf("expected:\n %b, got:\n %b", math.MaxUint32>>(32-24), v.hour)
        }
        if v.day != (math.MaxUint32>>(32-31))<<1 {
            t.Errorf("expected:\n %b, got:\n %b", (math.MaxUint32>>(32-31))<<1, v.day)
        }
        if v.month != (math.MaxUint16>>(16-12))<<1 {
            t.Errorf("expected:\n %b, got:\n %b", (math.MaxUint16>>(16-12))<<1, v.month)
        }
        if v.week != (math.MaxUint8>>1)<<1 {
            t.Errorf("expected:\n %b, got:\n %b", math.MaxUint8>>1, v.week)
        }
    })
    t.Run("parseCron_len_7", func(t *testing.T) {
        v1, err := parseCron("0 0 0 * * ? *")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        v := v1.(*schedulerCron)
        if v.second != 1 {
            t.Errorf("expected:\n %b, got:\n %b", 1, v.second)
        }
        if v.minute != 1 {
            t.Errorf("expected:\n %b, got:\n %b", 1, v.minute)
        }
        if v.hour != 1 {
            t.Errorf("expected:\n %b, got:\n %b", 1, v.hour)
        }
        if v.day != (math.MaxUint32>>(32-31))<<1 {
            t.Errorf("expected:\n %b, got:\n %b", (math.MaxUint32>>(32-31))<<1, v.day)
        }
        if v.month != (math.MaxUint16>>(16-12))<<1 {
            t.Errorf("expected:\n %b, got:\n %b", (math.MaxUint16>>(16-12))<<1, v.month)
        }
        if v.week != (math.MaxUint8>>1)<<1 {
            t.Errorf("expected:\n %b, got:\n %b", math.MaxUint8>>1, v.week)
        }
    })
    t.Run("parseCron_fixed_yearly", func(t *testing.T) {
        _, err := parseCron("@yearly")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
    })
    t.Run("parseCron_fixed_every_1s", func(t *testing.T) {
        _, err := parseCron("@every 1s")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
    })
    t.Run("parseCron_fixed_every_1h", func(t *testing.T) {
        _, err := parseCron("@every 1h")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
    })
    t.Run("parseCron_fixed_every_1h1m1s1ms", func(t *testing.T) {
        _, err := parseCron("@every 1h1m1s1ms")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
    })
}

func Test_ParseCron_WithInvalidExpression(t *testing.T) {
    t.Run("parseCron_WithInvalidExpression_string", func(t *testing.T) {
        _, err := parseCron("abc")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_step_0", func(t *testing.T) {
        _, err := parseCron("*/0 * * * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_step_-1", func(t *testing.T) {
        _, err := parseCron("*/-1 * * * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_step_60", func(t *testing.T) {
        _, err := parseCron("*/60 * * * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_range_0-60", func(t *testing.T) {
        _, err := parseCron("0-60 * * * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_range_6-99", func(t *testing.T) {
        _, err := parseCron("6-99 * * * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_single_60", func(t *testing.T) {
        _, err := parseCron("60 * * * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_single_-1", func(t *testing.T) {
        _, err := parseCron("-1 * * * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_len_4", func(t *testing.T) {
        _, err := parseCron("* * * ?")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_len_8", func(t *testing.T) {
        _, err := parseCron("* * * * * * * *")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
    })
    t.Run("parseCron_WithInvalidExpression_setDayAndWeek", func(t *testing.T) {
        _, err := parseCron("0 0 0 1 6 5")
        if err == nil {
            t.Errorf("expected error, got nil")
        }
        t.Log(err)
    })
}

func Test_Cron_nextTime(t *testing.T) {
    t.Run("Cron_nextTime_sec", func(t *testing.T) {
        c, _ := parseCron("*/5 * * * * ?")
        next := c.NextTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local))
        expect := time.Date(2020, 1, 1, 0, 0, 5, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_month", func(t *testing.T) {
        c, _ := parseCron("0 0 0 * * ? *")
        next := c.NextTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local))
        expect := time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_yearly", func(t *testing.T) {
        c, _ := parseCron("@yearly")
        next := c.NextTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local))
        expect := time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_every", func(t *testing.T) {
        c, err := parseCron("@every 1s")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
            return
        }
        next := c.NextTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local))
        expect := time.Date(2020, 1, 1, 0, 0, 1, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_week", func(t *testing.T) {
        c, _ := parseCron("0 0 0 * * 1")
        date := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
        expect := time.Date(2020, 1, 5, 0, 0, 0, 0, time.Local)
        next := c.NextTime(date)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_month_week", func(t *testing.T) {
        c, _ := parseCron("0 0 0 * 3 5")
        next := c.NextTime(time.Date(2020, 1, 5, 0, 0, 0, 0, time.Local))
        expect := time.Date(2020, 3, 12, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_month_week1", func(t *testing.T) {
        c, err := parseCron("0 0 0 * 6 7")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
            return
        }
        next := c.NextTime(time.Date(2020, 3, 5, 0, 0, 0, 0, time.Local))
        expect := time.Date(2020, 6, 6, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_year_month_week1", func(t *testing.T) {
        c, err := parseCron("0 0 0 * 6 7")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
            return
        }
        next := c.NextTime(time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local))
        expect := time.Date(2021, 6, 5, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
    t.Run("Cron_nextTime_year", func(t *testing.T) {
        c, err := parseCron("0 0 0 * * * 2024")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
            return
        }
        next := c.NextTime(time.Date(2021, 12, 5, 0, 0, 0, 0, time.Local))
        expect := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
        next = c.NextTime(time.Date(2022, 12, 5, 0, 0, 0, 0, time.Local))
        expect = time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
        next = c.NextTime(time.Date(2023, 12, 5, 0, 0, 0, 0, time.Local))
        expect = time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
        if next != expect {
            t.Errorf("expected: %v, got: %v", expect.Format(time.DateTime), next.Format(time.DateTime))
        }
    })
}
