package config

import (
    "fmt"
    "strconv"
)

type Item string

func (i Item) String() string {
    return "configItem(" + string(i) + ")"
}

func (i Item) ValueAny(m any) any {
    return GetByPathAny(m, string(i))
}

func (i Item) ValueString(m any) string {
    return i.DefaultString(m, "")
}

func (i Item) DefaultString(m any, defaultValue string) string {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return defaultValue
    }
    if s, ok := v.(string); ok {
        if s == "" {
            return defaultValue
        }
        return s
    }
    r := fmt.Sprint(v)
    if r == "" {
        return defaultValue
    }
    return r
}

func (i Item) ValueInt(m any) int {
    return i.DefaultInt(m, 0)
}

func (i Item) DefaultInt(m any, defaultValue int) int {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return defaultValue
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
        p, err := strconv.ParseInt(v1, 10, 64)
        if err != nil {
            return defaultValue
        }
        return int(p)
    default:
        return defaultValue
    }
}

func (i Item) ValueFloat(m any) float64 {
    return i.DefaultFloat(m, 0)
}

func (i Item) DefaultFloat(m any, defaultValue float64) float64 {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return defaultValue
    }
    switch v1 := v.(type) {
    case float32:
        return float64(v1)
    case float64:
        return v1
    case string:
        p, err := strconv.ParseFloat(v1, 64)
        if err != nil {
            return defaultValue
        }
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
        return defaultValue
    }
}

func (i Item) ValueBool(m any) bool {
    return i.DefaultBool(m, false)
}

func (i Item) DefaultBool(m any, defaultValue bool) bool {
    v := GetByPathAny(m, string(i))
    if v == nil {
        return defaultValue
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
    return defaultValue
}

func NewDataItem(data any, key string) DataItem {
    return DataItem{data, key}
}

type DataItem struct {
    m   any
    key string
}

func (i DataItem) String() string {
    return "dataItem(" + i.key + ")"
}

func (i DataItem) ValueAny() any {
    return GetByPathAny(i.m, i.key)
}
func (i DataItem) ValueString() string {
    return Item(i.key).DefaultString(i.m, "")
}
func (i DataItem) DefaultString(defaultValue string) string {
    return Item(i.key).DefaultString(i.m, defaultValue)
}
func (i DataItem) ValueInt() int {
    return Item(i.key).DefaultInt(i.m, 0)
}
func (i DataItem) DefaultInt(defaultValue int) int {
    return Item(i.key).DefaultInt(i.m, defaultValue)
}
func (i DataItem) ValueFloat() float64 {
    return Item(i.key).DefaultFloat(i.m, 0)
}
func (i DataItem) DefaultFloat(defaultValue float64) float64 {
    return Item(i.key).DefaultFloat(i.m, defaultValue)
}
func (i DataItem) ValueBool() bool {
    return Item(i.key).DefaultBool(i.m, false)
}
func (i DataItem) DefaultBool(defaultValue bool) bool {
    return Item(i.key).DefaultBool(i.m, defaultValue)
}
func (i DataItem) ValueChild(name string) DataItem {
    if name == "" {
        return i
    }
    if i.key == "" {
        return DataItem{i.m, name}
    }
    return DataItem{i.m, i.key + "." + name}
}
