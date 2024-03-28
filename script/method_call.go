package script

type method interface {
    call(...any) any
}

type oneArgMethod func(any) any

func (m oneArgMethod) call(args ...any) any {
    if len(args) < 1 {
        panic("方法调用最少传入一个参数")
    }
    return m(args[0])
}

type twoArgMethod func(any, any) any

func (m twoArgMethod) call(args ...any) any {
    if len(args) < 2 {
        panic("方法调用最少传入两个参数")
    }
    return m(args[0], args[1])
}

type threeArgMethod func(any, any, any) any

func (m threeArgMethod) call(args ...any) any {
    if len(args) < 3 {
        panic("方法调用最少传入三个参数")
    }
    return m(args[0], args[1], args[2])
}

type multiArgMethod func(...any) any

func (m multiArgMethod) call(args ...any) any {
    return m(args...)
}
