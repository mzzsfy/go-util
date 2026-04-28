package helper

// Ternary 三元运算符,根据test返回trueValue或falseValue
func Ternary[T any](test bool, trueValue, falseValue T) T {
    if test {
        return trueValue
    }
    return falseValue
}

// TernaryF 三元运算符,根据test调用trueValue或falseValue并返回结果
func TernaryF[T any, F func() T](test bool, trueValue, falseValue F) T {
    if test {
        return trueValue()
    }
    return falseValue()
}

// TernaryVF 三元运算符,根据test返回trueValue或调用falseValue
func TernaryVF[T any, F func() T](test bool, trueValue T, falseValue F) T {
    if test {
        return trueValue
    }
    return falseValue()
}
