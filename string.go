package util

import (
    "hash/fnv"
    "strconv"
    "strings"
)

// PaddingOrTruncate 填充空格或截取到指定长度
// leftOrRight 填充或者截取方向,与对齐方向逻辑相反,默认false,左对齐
func PaddingOrTruncate[S ~string](str S, toLen int, leftOrRight ...bool) S {
    if len([]rune(str)) > toLen {
        return Truncate(str, toLen, leftOrRight...)
    } else {
        return Padding(str, toLen, leftOrRight...)
    }
}

// TruncateAndAppendSuffix 截断字符串到指定长度,如果截断了,则添加后缀
func TruncateAndAppendSuffix[S ~string](str S, toLen int, suffix S, leftOrRight ...bool) S {
    truncate := Truncate(str, toLen, leftOrRight...)
    if str != truncate {
        return truncate + suffix
    }
    return str
}

// Truncate 截断字符串
func Truncate[S ~string](str S, toLen int, leftOrRight ...bool) S {
    lor := false
    if len(leftOrRight) > 0 {
        lor = leftOrRight[0]
    }
    r := []rune(str)
    l := len(r)
    if l > toLen {
        if lor {
            return S(r[l-toLen:])
        } else {
            return S(r[:toLen])
        }
    }
    return str
}

// Padding 填充字符串到指定长度
func Padding[S ~string](str S, toLen int, leftOrRight ...bool) S {
    lor := false
    if len(leftOrRight) > 0 {
        lor = leftOrRight[0]
    }
    l := len([]rune(str))
    if l < toLen {
        if lor {
            return S(strings.Repeat(" ", toLen-l) + string(str))
        } else {
            return S(string(str) + strings.Repeat(" ", toLen-l))
        }
    }
    return str
}

func Hash(text string) uint32 {
    algorithm := fnv.New32a()
    algorithm.Write([]byte(text))
    return algorithm.Sum32()
}
func HashInt(text string) int {
    return int(Hash(text))
}

type StringBuilder struct {
    strings.Builder
}

func (s *StringBuilder) Append(str string) *StringBuilder {
    _, _ = s.WriteString(str)
    return s
}

func IsInteger(str string) bool {
    if str == "" {
        return false
    }
    _, err := strconv.Atoi(str)
    if err != nil {
        return false
    }
    return true
}
