package helper

var (
    removeStack = uintptr(0)
)

type Err struct {
    Error any
    Stack []Stack
}

func TryWithStack(f func(), catch func(err Err)) {
    defer func() {
        if err := recover(); err != nil {
            stack := CallerStack(2)
            for i, s := range stack {
                if s.PC == removeStack {
                    stack = append(stack[:i], stack[i+1:]...)
                    catch(Err{Error: err, Stack: stack})
                    return
                }
            }
            removeStack = CallerStack(3, 1)[0].PC
            for i, s := range stack {
                if s.PC == removeStack {
                    stack = append(stack[:i], stack[i+1:]...)
                    catch(Err{Error: err, Stack: stack})
                    return
                }
            }
        }
    }()
    f()
}
