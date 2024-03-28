package script

type loop interface {
    Continue() bool
}

type forLoop struct {
    offset_   int
    init      bool
    start     func()
    condition func() bool
    end       func()
    body      func()
}

func (f forLoop) offset() int {
    return f.offset_
}

func (f forLoop) Continue() bool {
    if !f.init {
        f.init = true
        if f.start != nil {
            f.start()
        }
    }
    if f.condition != nil {
        v := f.condition()
        if v {
            if f.body != nil {
                f.body()
            }
            if f.end != nil {
                f.end()
            }
            return true
        }
    }
    return false
}

type whileLoop struct {
    offset_   int
    condition func() bool
    //0: continue 1: break 2: return
    body func() int
}

func (f whileLoop) Continue() bool {
    if f.condition != nil {
        v := f.condition()
        if v {
            if f.body != nil {
                r := f.body()
                if r > 0 {
                    return false
                }
            }
            return true
        }
    }
    return false
}

func (f whileLoop) offset() int {
    return f.offset_
}
