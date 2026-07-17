package helper

import (
    "math"
    "testing"
)

func Test_Abs(t *testing.T) {
    t.Parallel()
    // 正数不变
    if got := Abs(5); got != 5 {
        t.Errorf("Abs(5) = %d, want 5", got)
    }
    // 零不变
    if got := Abs(0); got != 0 {
        t.Errorf("Abs(0) = %d, want 0", got)
    }
    // 普通负数
    if got := Abs(-42); got != 42 {
        t.Errorf("Abs(-42) = %d, want 42", got)
    }
    // float64 负数
    if got := Abs(-3.14); got != 3.14 {
        t.Errorf("Abs(-3.14) = %f, want 3.14", got)
    }
    // int64 MinInt: 饱和到 MaxInt
    if got := Abs(int64(math.MinInt64)); got != math.MaxInt64 {
        t.Errorf("Abs(MinInt64) = %d, want MaxInt64 %d", got, math.MaxInt64)
    }
    // int32 MinInt32: 饱和到 MaxInt32
    if got := Abs(int32(math.MinInt32)); got != math.MaxInt32 {
        t.Errorf("Abs(MinInt32) = %d, want MaxInt32 %d", got, math.MaxInt32)
    }
    // int8 MinInt8: 饱和到 MaxInt8
    if got := Abs(int8(-128)); got != int8(127) {
        t.Errorf("Abs(MinInt8) = %d, want MaxInt8 127", got)
    }
    // int64 边界: MaxInt64 的绝对值应不变
    if got := Abs(int64(math.MaxInt64)); got != math.MaxInt64 {
        t.Errorf("Abs(MaxInt64) = %d, want MaxInt64 %d", got, math.MaxInt64)
    }
}

func Test_MinN(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name string
        args []int
        want int
    }{
        {"single", []int{5}, 5},
        {"two", []int{3, 7}, 3},
        {"negative", []int{-3, -7, -1}, -7},
        {"mixed", []int{-1, 0, 1}, -1},
        {"duplicates", []int{2, 2, 2}, 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := MinN(tt.args...); got != tt.want {
                t.Errorf("MinN(%v) = %v, want %v", tt.args, got, tt.want)
            }
        })
    }

    // int64 边界
    if got := MinN(int64(0), math.MinInt64, math.MaxInt64); got != math.MinInt64 {
        t.Errorf("MinN int64 boundary = %v, want %v", got, math.MinInt64)
    }

    // float64
    if got := MinN(1.5, -2.5, 0.0); got != -2.5 {
        t.Errorf("MinN float64 = %v, want -2.5", got)
    }
}

func Test_MaxN(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name string
        args []int
        want int
    }{
        {"single", []int{5}, 5},
        {"two", []int{3, 7}, 7},
        {"negative", []int{-3, -7, -1}, -1},
        {"mixed", []int{-1, 0, 1}, 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := MaxN(tt.args...); got != tt.want {
                t.Errorf("MaxN(%v) = %v, want %v", tt.args, got, tt.want)
            }
        })
    }
}

func Test_StringIsInteger(t *testing.T) {
    t.Parallel()
    tests := []struct {
        input string
        want  bool
    }{
        {"", false},
        {"0", true},
        {"123", true},
        {"-456", true},
        {"+123", true},
        {"1.5", false},
        {"42", true},
        {"-", false},
        {"+", false},
        {"99999999999999999999", true},
        {"12abc", false},
        {"3.14", false},
        {"0xff", false},
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            if got := StringIsInteger(tt.input); got != tt.want {
                t.Errorf("StringIsInteger(%q) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}

func Test_Max(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"正常比较小大", 3, 5, 5},
        {"正常比较大小", 5, 3, 5},
        {"相等", 3, 3, 3},
        {"负数", -3, -5, -3},
        {"零值", 0, -1, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Max(tt.a, tt.b); got != tt.want {
                t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
    // float64 类型测试
    if got := Max(1.5, 2.5); got != 2.5 {
        t.Errorf("Max(1.5, 2.5) = %v, want 2.5", got)
    }
}

func Test_Min(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"正常比较小大", 3, 5, 3},
        {"正常比较大小", 5, 3, 3},
        {"相等", 3, 3, 3},
        {"负数", -3, -5, -5},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Min(tt.a, tt.b); got != tt.want {
                t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
    // float64 类型测试
    if got := Min(1.5, 2.5); got != 1.5 {
        t.Errorf("Min(1.5, 2.5) = %v, want 1.5", got)
    }
}

func Test_ParseStringToInt(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name         string
        input        string
        defaultValue int
        want         int
    }{
        {"空字符串返回默认值0", "", 0, 0},
        {"空字符串返回非零默认值", "", 99, 99},
        {"正常数字", "123", 0, 123},
        {"非法字符串", "abc", 0, 0},
        {"非法字符串非零默认值", "abc", -1, -1},
        {"负数", "-42", 0, -42},
        {"超大数溢出返回默认值", "99999999999999999999", 0, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := ParseStringToInt(tt.input, tt.defaultValue); got != tt.want {
                t.Errorf("ParseStringToInt(%q, %d) = %d, want %d",
                    tt.input, tt.defaultValue, got, tt.want)
            }
        })
    }
}

func Test_ParseStringToFloat(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name         string
        input        string
        defaultValue float64
        want         float64
    }{
        {"空字符串返回默认值", "", 0, 0},
        {"空字符串返回非零默认值", "", -1.0, -1.0},
        {"正常浮点", "3.14", 0, 3.14},
        {"整数形式", "42", 0, 42.0},
        {"非法字符串", "abc", 0, 0},
        {"非法字符串非零默认值", "abc", 99.5, 99.5},
        {"负数", "-1.5", 0, -1.5},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := ParseStringToFloat(tt.input, tt.defaultValue); got != tt.want {
                t.Errorf("ParseStringToFloat(%q, %v) = %v, want %v",
                    tt.input, tt.defaultValue, got, tt.want)
            }
        })
    }
}

func Test_NumberToString(t *testing.T) {
    t.Parallel()
    // int 类型测试
    if got := NumberToString(0); got != "0" {
        t.Errorf("NumberToString(0) = %q, want %q", got, "0")
    }
    if got := NumberToString(123); got != "123" {
        t.Errorf("NumberToString(123) = %q, want %q", got, "123")
    }
    if got := NumberToString(-456); got != "-456" {
        t.Errorf("NumberToString(-456) = %q, want %q", got, "-456")
    }
    // 大数: int64
    if got := NumberToString(int64(9999999999)); got != "9999999999" {
        t.Errorf("NumberToString(int64(9999999999)) = %q, want %q", got, "9999999999")
    }
    // uint64 最大值
    if got := NumberToString(uint64(18446744073709551615)); got != "18446744073709551615" {
        t.Errorf("NumberToString(uint64 max) = %q, want %q", got, "18446744073709551615")
    }
    // int64 最小值
    if got := NumberToString(int64(-9223372036854775808)); got != "-9223372036854775808" {
        t.Errorf("NumberToString(int64 min) = %q, want %q", got, "-9223372036854775808")
    }
}

func Test_StringAllIsNumber(t *testing.T) {
    t.Parallel()
    tests := []struct {
        input string
        want  bool
    }{
        {"", false},
        {"0", true},
        {"12345", true},
        {"-123", false},
        {"1.5", false},
        {"12abc", false},
        {"3.14", false},
        {"-42", false},
        {"+42", false},
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            if got := StringAllIsNumber(tt.input); got != tt.want {
                t.Errorf("StringAllIsNumber(%q) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}
