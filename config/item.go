package config

import (
    "fmt"
    "strconv"
)

func ValueFromPath(m any, path string) Value {
    a := GetByPathAny(m, path)
    if a == nil {
        return valueNil{}
    }
    switch v := a.(type) {
    case string:
        return valueString(v)
    default:
        return valueAny{a}
    }
}

// ValueFrom 从any获取值
func ValueFrom(a any) Value {
    if a == nil {
        return valueNil{}
    }
    switch v := a.(type) {
    case string:
        return valueString(v)
    default:
        return valueAny{a}
    }
}

type Value interface {
    Any() any
    AnyD(any) any
    String() string
    StringD(defaultValue string) string
    Int() int
    IntD(defaultValue int) int
    Float() float64
    FloatD(defaultValue float64) float64
    Bool() bool
    BoolD(defaultValue bool) bool
    Child(name string) Value
}

type valueNil struct{}

func (v valueNil) Any() any {
    return nil
}
func (v valueNil) AnyD(a any) any {
    return a
}
func (v valueNil) String() string {
    return "<nil>"
}
func (v valueNil) StringD(defaultValue string) string {
    return defaultValue
}
func (v valueNil) Int() int {
    panic("nil can not be converted to int")
}
func (v valueNil) IntD(defaultValue int) int {
    return defaultValue
}
func (v valueNil) Float() float64 {
    panic("nil can not be converted to float")
}
func (v valueNil) FloatD(defaultValue float64) float64 {
    return defaultValue
}
func (v valueNil) Bool() bool {
    panic("nil can not be converted to bool")
}
func (v valueNil) BoolD(defaultValue bool) bool {
    return defaultValue
}
func (v valueNil) Child(name string) Value {
    return valueNil{}
}

type valueString string

func (v valueString) Any() any {
    return string(v)
}
func (v valueString) AnyD(a any) any {
    if v != "" {
        return string(v)
    }
    return a
}
func (v valueString) String() string {
    return string(v)
}
func (v valueString) StringD(defaultValue string) string {
    return string(v)
}
func (v valueString) Int() int {
    i, err := strconv.Atoi(string(v))
    if err != nil {
        return 0
    }
    return i
}
func (v valueString) IntD(defaultValue int) int {
    i, err := strconv.Atoi(string(v))
    if err != nil {
        return defaultValue
    }
    return i
}
func (v valueString) Float() float64 {
    f, err := strconv.ParseFloat(string(v), 64)
    if err != nil {
        return 0
    }
    return f
}
func (v valueString) FloatD(defaultValue float64) float64 {
    f, err := strconv.ParseFloat(string(v), 64)
    if err != nil {
        return defaultValue
    }
    return f
}
func (v valueString) Bool() bool {
    return string(v) == "true" || string(v) == "yes"
}
func (v valueString) BoolD(defaultValue bool) bool {
    return string(v) == "true" || string(v) == "yes"
}
func (v valueString) Child(name string) Value {
    return valueNil{}
}

type valueAny struct{ value any }

func (v valueAny) Any() any {
    return v.value
}
func (v valueAny) AnyD(a any) any {
    if v.value == nil {
        return a
    }
    return v.value
}
func (v valueAny) String() string {
    return fmt.Sprintf("%v", v.value)
}
func (v valueAny) StringD(defaultValue string) string {
    return fmt.Sprintf("%v", v.value)
}
func (v valueAny) Int() int {
    return v.IntD(0)
}
func (v valueAny) IntD(defaultValue int) int {
    switch v1 := v.value.(type) {
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
    }
    return defaultValue
}
func (v valueAny) Float() float64 {
    return v.FloatD(0)
}
func (v valueAny) FloatD(defaultValue float64) float64 {
    switch v1 := v.value.(type) {
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
    case float32:
        return float64(v1)
    case float64:
        return v1
    }
    return defaultValue
}
func (v valueAny) Bool() bool {
    return v.BoolD(false)
}
func (v valueAny) BoolD(defaultValue bool) bool {
    switch v1 := v.value.(type) {
    case bool:
        return v1
    case string:
        return v1 == "true" || v1 == "yes"
    }
    return defaultValue
}

func (v valueAny) Child(name string) Value {
    return ValueFromPath(v.value, name)
}
