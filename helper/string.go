package helper

import (
    "github.com/mzzsfy/go-util/unsafe"
    "strconv"
    "strings"
)

// PaddingOrTruncate 填充空格或截取到指定长度
// leftOrRight 填充或者截取方向,与对齐方向逻辑相反,默认false,左对齐
func PaddingOrTruncate(str string, toLen int, leftOrRight ...bool) string {
    if len([]rune(str)) > toLen {
        return Truncate(str, toLen, leftOrRight...)
    } else {
        return Padding(str, toLen, leftOrRight...)
    }
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
func Truncate(str string, toLen int, leftOrRight ...bool) string {
    lor := false
    if len(leftOrRight) > 0 {
        lor = leftOrRight[0]
    }
    r := []rune(str)
    l := len(r)
    if l > toLen {
        if lor {
            return string(r[l-toLen:])
        } else {
            return string(r[:toLen])
        }
    }
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
    l := len([]rune(str))
    if l < toLen {
        if lor {
            return strings.Repeat(" ", toLen-l) + string(str)
        } else {
            return str + strings.Repeat(" ", toLen-l)
        }
    }
    return str
}

var StringHasher = unsafe.NewHasher[string]().WithSeed(0)

func Hash(text string) uint64 {
    return StringHasher.Hash(text)
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

// StringIsInteger 判断字符串是否是整数,可优化
func StringIsInteger(str string) bool {
    if str == "" {
        return false
    }
    _, err := strconv.Atoi(str)
    if err != nil {
        return false
    }
    return true
}

// StringAllIsNumber 字符串是否全是数字
func StringAllIsNumber(str string) bool {
    if str == "" {
        return false
    }
    for _, r := range str {
        if r < '0' || r > '9' {
            return false
        }
    }
    return true
}
