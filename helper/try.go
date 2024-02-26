package helper

import "runtime"

var (
    removeStack = uintptr(0)
)

type Err struct {
    Error any
    Stack []Stack
}

func TryWithStack(f func(), callback func(recoverValue any, stack []Stack)) {
    defer func() {
        if err := recover(); err != nil {
            stack := CallerStack(2)
            if cleanStack(&stack) {
                callback(err, stack)
                return
            }
            TryWithStack(func() {
                removeStack, _, _, _ = runtime.Caller(1)
            }, func(err any, stack []Stack) {})
            cleanStack(&stack)
            callback(err, stack)
        }
    }()
    f()
}

func cleanStack(stack *[]Stack) bool {
    for i, s := range *stack {
        if s.PC == removeStack {
            *stack = append((*stack)[:i], (*stack)[i+1:]...)
            return true
        }
    }
    return false
}
