package helper

type Err struct {
    Error any
    Stack []Stack
}

func TryWithStack(f func(), catch func(err Err)) {
    defer func() {
        if err := recover(); err != nil {
            catch(Err{Error: err, Stack: CallerStack(2)})
        }
    }()
    f()
}
