package script

type loop interface {
    Continue(Scope) bool
}

type forLoop struct {
    offset    int
    init      bool
    start     func(Scope)
    condition func(Scope) bool
    end       func(Scope)
    body      func(Scope)
}

func (f forLoop) Continue(scope Scope) bool {
    defer func() { recoverRunTimeError(recover(), f.offset) }()
    scope = NewChildScope(scope)
    if !f.init {
        f.init = true
        if f.start != nil {
            f.start(scope)
        }
    }
    if f.condition != nil {
        v := f.condition(scope)
        if v {
            if f.body != nil {
                f.body(scope)
            }
            if f.end != nil {
                f.end(scope)
            }
            return true
        }
    }
    return false
}

type whileLoop struct {
    offset    int
    condition func(Scope) bool
    //0: continue 1: break 2: return
    body func(Scope) int
}

func (f whileLoop) Continue(scope Scope) bool {
    defer func() { recoverRunTimeError(recover(), f.offset) }()
    scope = NewChildScope(scope)
    if f.condition != nil {
        v := f.condition(scope)
        if v {
            if f.body != nil {
                r := f.body(scope)
                if r > 0 {
                    return false
                }
            }
            return true
        }
    }
    return false
}
