package helper

import (
    "strings"
    "unicode/utf8"
    "unsafe"
)

// PaddingOrTruncate 填充空格或截取到指定长度
// leftOrRight 填充或者截取方向,与对齐方向逻辑相反,默认false,左对齐
func PaddingOrTruncate(str string, toLen int, leftOrRight ...bool) string {
    // 一次遍历: 计算 rune 数量,超过 toLen 时立即截断
    runeCount := 0
    for _ = range str {
        runeCount++
        if runeCount > toLen {
            return Truncate(str, toLen, leftOrRight...)
        }
    }
    if runeCount < toLen {
        return Padding(str, toLen, leftOrRight...)
    }
    return str
}

// TruncateAndAppendSuffix 截断字符串到指定长度,如果截断了,则添加后缀
func TruncateAndAppendSuffix(str string, toLen int, suffix string, leftOrRight ...bool) string {
    truncate := Truncate(str, toLen, leftOrRight...)
    if str != truncate {
        return truncate + suffix
    }
    return str
}

// Truncate 截断字符串,保证长度不超过指定长度
// 使用 utf8 遍历避免 []rune 分配
func Truncate(str string, toLen int, leftOrRight ...bool) string {
    lor := false
    if len(leftOrRight) > 0 {
        lor = leftOrRight[0]
    }
    // 快速路径: 纯 ASCII 且字节长度小于等于目标, 直接返回
    if len(str) <= toLen {
        return str
    }

    if lor {
        // 从右侧截取: 找到后 toLen 个 rune 的起始位置
        totalRunes := 0
        for range str {
            totalRunes++
        }
        if totalRunes <= toLen {
            return str
        }
        skip := totalRunes - toLen
        runeCount := 0
        for i := range str {
            if runeCount == skip {
                return str[i:]
            }
            runeCount++
        }
        return str
    }

    // 从左侧截取: 找到前 toLen 个 rune 的结束位置
    runeCount := 0
    for i := range str {
        if runeCount == toLen {
            return str[:i]
        }
        runeCount++
    }
    // rune 总数不足 toLen, 直接返回
    return str
}

// Sub 截取字符串
// before 保留flag之前的字符串还是之后
// last 从前开始查找还是从后开始查找
func Sub(src, flag string, before, last bool) string {
    var index int
    if last {
        index = strings.LastIndex(src, flag)
    } else {
        index = strings.Index(src, flag)
    }
    if index == -1 {
        return src
    }
    if before {
        return src[:index]
    }
    return src[index+len(flag):]
}

// SubBefore 截取flag之前的字符串
func SubBefore(src, flag string) string {
    return Sub(src, flag, true, false)
}

//SubAfter 截取flag之后的字符串
func SubAfter(src, flag string) string {
    return Sub(src, flag, false, true)
}

// SubByte 通过byte截取字符串
// before 保留flag之前的字符串还是之后
// last 从前开始查找还是从后开始查找
func SubByte(src string, flag byte, before, last bool) string {
    var index int
    if last {
        index = strings.LastIndexByte(src, flag)
    } else {
        index = strings.IndexByte(src, flag)
    }
    if index == -1 {
        return src
    }
    if before {
        return src[:index]
    }
    return src[index+1:]
}

// SubByteBefore 截取flag之前的字符串
func SubByteBefore(src string, flag byte) string {
    return SubByte(src, flag, true, false)
}

//SubByteAfter 截取flag之后的字符串
func SubByteAfter(src string, flag byte) string {
    return SubByte(src, flag, false, true)
}

// Padding 填充字符串到指定长度
func Padding(str string, toLen int, leftOrRight ...bool) string {
    lor := false
    if len(leftOrRight) > 0 {
        lor = leftOrRight[0]
    }
    l := utf8.RuneCountInString(str)
    if l < toLen {
        if lor {
            return strings.Repeat(" ", toLen-l) + string(str)
        }

        return str + strings.Repeat(" ", toLen-l)
    }
    return str
}

// stringHashSeed 字符串哈希种子
var stringHashSeed = newStringHashSeed()

// strhashFunc 运行时字符串哈希函数, 在 init 阶段初始化
var strhashFunc func(p unsafe.Pointer, h uintptr) uintptr

var StringHasher = stringHasher{}

type stringHasher struct{}

func (stringHasher) Hash(text string) uint64 {
    return uint64(strhashFunc(noescape(unsafe.Pointer(&text)), stringHashSeed))
}

func Hash(text string) uint64 {
    return StringHasher.Hash(text)
}

// noescape 阻止逃逸分析, 与 runtime 内部实现相同
//
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
    x := uintptr(p)
    return unsafe.Pointer(x ^ 0)
}

type StringBuilder struct {
    strings.Builder
}

// Append 便于链式调用
func (s *StringBuilder) Append(str string) *StringBuilder {
    _, _ = s.WriteString(str)
    return s
}

// AppendByte 便于链式调用
func (s *StringBuilder) AppendByte(b byte) *StringBuilder {
    _ = s.WriteByte(b)
    return s
}

// AppendBytes 便于链式调用
func (s *StringBuilder) AppendBytes(b []byte) *StringBuilder {
    _, _ = s.Write(b)
    return s
}

// StringIsInteger 判断字符串是否是整数, 支持正负号前缀
func StringIsInteger(str string) bool {
    if len(str) == 0 {
        return false
    }
    // 允许前导符号
    i := 0
    if str[0] == '-' || str[0] == '+' {
        i = 1
    }
    // 符号后必须有至少一个数字
    if i >= len(str) {
        return false
    }
    // 检查剩余部分是否全是数字

    for ; i < len(str); i++ {
        if str[i] < '0' || str[i] > '9' {
            return false
        }
    }
    return true
}

// StringAllIsNumber 字符串是否全是数字, 使用字节索引而非 range 遍历
func StringAllIsNumber(str string) bool {
    if str == "" {
        return false
    }
    for i := 0; i < len(str); i++ {
        if str[i] < '0' || str[i] > '9' {
            return false
        }
    }
    return true
}
