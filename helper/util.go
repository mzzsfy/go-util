package helper

// AnyArray 将多个任意类型参数转换为[]any
func AnyArray(vars ...any) []any {
    return vars
}

// AnyArrayT 将多个同类型参数转换为切片
func AnyArrayT[T any](vars ...T) []T {
    return vars
}

// Must 如果err不为nil则panic,否则返回data
func Must[T any](data T, err error) T {
    if err != nil {
        panic(err)
    }
    return data
}

// MustR 如果err不为nil则panic,否则返回data(参数顺序与Must相反)
func MustR[T any](err error, data T) T {
    if err != nil {
        panic(err)
    }
    return data
}
