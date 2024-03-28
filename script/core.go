package script

import "fmt"

type Scope interface {
    Get(string) (any, error)
    New(string, any) error
    Update(string, any) error
}
type RuntimeError struct {
    err    any
    offset int
}

func (r RuntimeError) String() string {
    return fmt.Sprintf("runtime error: %v, offset: %d", r.err, r.offset)
}

func recoverRunTimeError(r any, offset int) {
    if r != nil {
        if _, ok := r.(RuntimeError); ok {
            panic(r)
        }
        panic(RuntimeError{
            err:    r,
            offset: offset,
        })
    }
}

type Engine interface {
    // Execute 执行脚本
    Execute(map[string]any) Result
    // Bind 支持绑定变量和函数
    Bind(map[string]any) (Engine, error)
}

type expression interface {
    compute(Scope) any
}

type position interface {
    offset() int
}

type Result interface {
    // IsErr 是否错误, 如果为true, 则使用Any获取错误信息
    IsErr() bool
    Any() any
    Int() int
    Bool() bool
    String() string
    Float() float64
}

type res struct {
    err   error
    value any
}

func (r res) IsErr() bool {
    return r.err != nil
}

func (r res) Any() any {
    if r.IsErr() {
        return r.err
    }
    return r.value
}

func (r res) Int() int {
    if i, ok := r.value.(int); ok {
        return i
    }
    panic("不是int")
}

func (r res) Bool() bool {
    if i, ok := r.value.(bool); ok {
        return i
    }
    panic("不是bool")
}

func (r res) String() string {
    if i, ok := r.value.(string); ok {
        return i
    }
    return fmt.Sprint(r.value)
}

func (r res) Float() float64 {
    if i, ok := r.value.(float64); ok {
        return i
    }
    panic("不是float")
}
