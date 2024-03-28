package script

import "fmt"

type variable interface {
    value(Scope) any
}

type fixedValue struct {
    v any
}

func (f fixedValue) value(e Scope) any {
    return f.v
}

func (f fixedValue) String() string {
    return fmt.Sprint(f.v)
}

type varValue struct {
    name     string
    negative bool
}

func (v *varValue) value(e Scope) any {
    get, err := e.Get(v.name)
    if err != nil {
        panic(err)
    }
    if v.negative {
        switch get.(type) {
        case int:
            return -get.(int)
        case float64:
            return -get.(float64)
        default:
            panic("无法设置为负数: " + v.String())
        }
    }
    return get
}

func (v *varValue) String() string {
    if v.negative {
        return "-${" + v.name + "}"
    }
    return "${" + v.name + "}"
}
