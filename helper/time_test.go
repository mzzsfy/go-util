package helper

import (
    "math"
    "strings"
    "testing"
    "time"
)

func TestLocalTime_UnmarshalJSON(t *testing.T) {
    t.Run("ValidTimeFormat", func(t *testing.T) {
        var lt LocalTime
        err := lt.UnmarshalJSON([]byte("\"2022-12-31 23:59:59\""))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
    })

    t.Run("InvalidTimeFormat", func(t *testing.T) {
        var lt LocalTime
        err := lt.UnmarshalJSON([]byte("\"invalid time\""))
        if err == nil {
            t.Errorf("Expected error, got nil")
        }
    })

    t.Run("EmptyInput", func(t *testing.T) {
        var lt LocalTime
        err := lt.UnmarshalJSON([]byte{})
        if err == nil {
            t.Errorf("Expected error for empty input, got nil")
        }
    })

    t.Run("NullJSON", func(t *testing.T) {
        var lt LocalTime
        err := lt.UnmarshalJSON([]byte("null"))
        // null 不是合法时间格式, 应返回错误
        if err == nil {
            t.Errorf("Expected error for 'null' input, got nil")
        }
    })
}

func TestLocalTime_MarshalJSON(t *testing.T) {
    t.Run("MarshalTime", func(t *testing.T) {
        lt := LocalTime(time.Date(2022, 12, 31, 23, 59, 59, 0, time.Local))
        b, err := lt.MarshalBinary()
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        if string(b) != "2022-12-31 23:59:59" {
            t.Errorf("Expected 2022-12-31 23:59:59, got %s", string(b))
        }
    })
}

func TestParseLocalTime(t *testing.T) {
    t.Run("ValidTimeFormat", func(t *testing.T) {
        _, err := ParseLocalTime("2022-12-31 23:59:59")
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
    })

    t.Run("InvalidTimeFormat", func(t *testing.T) {
        _, err := ParseLocalTime("invalid time")
        if err == nil {
            t.Errorf("Expected error, got nil")
        }
    })
}

func TestParseLocalTimeAuto(t *testing.T) {
    t.Run("ValidTimeFormat", func(t *testing.T) {
        _, err := ParseLocalTimeAuto("20221231235959")
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
    })

    t.Run("InvalidTimeFormat", func(t *testing.T) {
        _, err := ParseLocalTimeAuto("invalid time")
        if err == nil {
            t.Errorf("Expected error, got nil")
        }
    })
}
func Test_ParseLocalTimeAuto(t *testing.T) {
    t.Run("ValidTimeFormat", func(t *testing.T) {
        _, err := ParseLocalTime("2022-12-31 23:59:59")
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
    })

    t.Run("InvalidTimeFormat", func(t *testing.T) {
        _, err := ParseLocalTime("invalid time")
        if err == nil {
            t.Errorf("Expected error, got nil")
        }
    })
    t.Run("ValidTimeFormatAuto", func(t *testing.T) {
        localTime, err := ParseLocalTimeAuto(time.Now().String())
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        if localTime.Time().IsZero() {
            t.Errorf("Expected non-zero time, got zero")
        }
    })
}

func Test_ParseLocalTimeAuto1(t *testing.T) {
    t.Run("ParseLocalTimeAuto_Number", func(t *testing.T) {
        _, err := ParseLocalTimeAuto("20221231235959")
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(strings.Replace(time.Now().Format("20060102150405.000000000"), ".", "", 1))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(strings.Replace(time.Now().Format("20060102150405.999999999"), ".", "", 1))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("20060102150405"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("20060102"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("150405"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
    })
    t.Run("ParseLocalTimeAuto_Str", func(t *testing.T) {
        _, err := ParseLocalTimeAuto(time.Now().Format("2006-01-02'T'15:04:05'Z'"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006-01-02T15:04:05Z"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006-01-02'T'15:04:05"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006-01-02T15:04:05"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006-01-02 15:04:05"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006年01月02日15点04分05秒"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006/01/02 15:04:05"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006-01-02"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006/01/02"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("15:04:05"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        _, err = ParseLocalTimeAuto(time.Now().Format("2006-01-02 15:04:05.000000000"))
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
    })
}

// TestFormatDuration_Negative 验证负数 duration 正确处理,保留负号
func TestFormatDuration_Negative(t *testing.T) {
    tests := []struct {
        name     string
        input    time.Duration
        negative bool
    }{
        {"零值", 0, false},
        {"负5秒", -5 * time.Second, true},
        {"负100毫秒", -100 * time.Millisecond, true},
        {"负1分钟", -time.Minute, true},
        {"正5秒", 5 * time.Second, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatDuration(tt.input)
            s := got.String()
            if tt.negative && !strings.HasPrefix(s, "-") {
                t.Errorf("FormatDuration(%v) = %q, 期望负号前缀", tt.input, s)
            }
            if !tt.negative && strings.HasPrefix(s, "-") {
                t.Errorf("FormatDuration(%v) = %q, 不应有负号前缀", tt.input, s)
            }
        })
    }
}

// TestFormatDuration_RoundTrip 验证 FormatDuration 后字符串长度合理
func TestFormatDuration_RoundTrip(t *testing.T) {
    tests := []struct {
        name  string
        input time.Duration
    }{
        {"大于10分钟", 11*time.Minute + 30*time.Second},
        {"1到10分钟之间", 5*time.Minute + 30*time.Second + 100*time.Millisecond},
        {"10秒到1分钟", 30*time.Second + 500*time.Millisecond},
        {"100毫秒到10秒", 5*time.Second + 123*time.Millisecond},
        {"10毫秒到100毫秒", 50*time.Millisecond + 123*time.Microsecond},
        {"1毫秒到10毫秒", 5*time.Millisecond + 123*time.Microsecond},
        {"小于1毫秒", 500 * time.Microsecond},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatDuration(tt.input)
            s := got.String()
            // 格式化后的字符串长度应 <= 7 (负数允许 <= 8)
            maxLen := 7
            if tt.input < 0 {
                maxLen = 8
            }
            if len(s) > maxLen {
                t.Errorf("FormatDuration(%v) = %q, 长度 %d 超过上限 %d", tt.input, s, len(s), maxLen)
            }
        })
    }
}

func FuzzFormatDuration(f *testing.F) {
    f.Add(int64(0))
    f.Add(int64(math.MaxInt64))
    f.Add(int64(time.Hour))
    f.Fuzz(func(t *testing.T, d int64) {
        s := FormatDuration(time.Duration(d))
        if d < int64(time.Hour*10) && len(s.String()) > 7 {
            t.Errorf("Expected length of formatted duration to be <= 7, got %d", len(s.String()))
        }
    })
}
