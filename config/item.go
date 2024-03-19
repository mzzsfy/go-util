package config

import (
    "fmt"
    "strconv"
)

type Item string

func (i Item) String() string {
    return "configItem(" + string(i) + ")"
}

func (i Item) AnyValue(m any) any {
    return GetByPathAny(m, string(i))
}

func (i Item) StringValue(m any) string {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return fmt.Sprint(v)
}

func (i Item) IntValue(m any) int {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return 0
    }
    switch v1 := v.(type) {
    case int:
        return v1
    case int8:
        return int(v1)
    case int16:
        return int(v1)
    case int32:
        return int(v1)
    case int64:
        return int(v1)
    case uint:
        return int(v1)
    case uint8:
        return int(v1)
    case uint16:
        return int(v1)
    case uint32:
        return int(v1)
    case uint64:
        return int(v1)
    case float32:
        return int(v1)
    case float64:
        return int(v1)
    case string:
        p, _ := strconv.ParseInt(v1, 10, 64)
        return int(p)
    default:
        panic("无法转换为int")
    }
}
func (i Item) FloatValue(m any) float64 {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return 0
    }
    switch v1 := v.(type) {
    case float32:
        return float64(v1)
    case float64:
        return v1
    case string:
        p, _ := strconv.ParseFloat(v1, 64)
        return p
    case int:
        return float64(v1)
    case int8:
        return float64(v1)
    case int16:
        return float64(v1)
    case int32:
        return float64(v1)
    case int64:
        return float64(v1)
    case uint:
        return float64(v1)
    case uint8:
        return float64(v1)
    case uint16:
        return float64(v1)
    case uint32:
        return float64(v1)
    case uint64:
        return float64(v1)
    default:
        panic("无法转换为float64")
    }
}

func (i Item) BoolValue(m any) bool {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return false
    }
    if b, ok := v.(bool); ok {
        return b
    }
    if s, ok := v.(string); ok {
        return s == "true" || s == "yes"
    }
    if i1, ok := v.(int); ok {
        return i1 > 0
    }
    if i1, ok := v.(uint); ok {
        return i1 > 0
    }
    return false
}
