package script

type Scope interface {
    Get(string) (any, error)
    Set(string, any) error
}
type Engine interface {
    // Execute 执行脚本
    Execute(map[string]any) Result
    // Bind 支持绑定变量和函数
    Bind(map[string]any) (Engine, error)
}

type expression interface {
    compute() any
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
    return r.value.(int)
}

func (r res) Bool() bool {
    return r.value.(bool)
}

func (r res) String() string {
    return r.value.(string)
}

func (r res) Float() float64 {
    return r.value.(float64)
}
