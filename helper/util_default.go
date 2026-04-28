package helper

// Default 如果test非零值则返回test,否则返回defaultValue
func Default[T any](test, defaultValue T) T {
    if !IsZero(test) {
        return test
    }
    return defaultValue
}

// Defaults 返回第一个非零值,如果全为零值则返回defaultValue
func Defaults[T any](defaultValue T, tests ...T) T {
    for _, t := range tests {
        if !IsZero(t) {
            return t
        }
    }
    return defaultValue
}

// NotZero 判断值是否非零
func NotZero(test any) bool {
    return !IsZero(test)
}
