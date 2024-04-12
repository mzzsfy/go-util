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
}

func TestLocalTime_MarshalJSON(t *testing.T) {
    t.Run("MarshalTime", func(t *testing.T) {
        lt := LocalTime(time.Date(2022, 12, 31, 23, 59, 59, 0, time.Local))
        b, err := lt.MarshalJSON()
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        if string(b) != "\"2022-12-31 23:59:59\"" {
            t.Errorf("Expected \"2022-12-31 23:59:59\", got %s", string(b))
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
