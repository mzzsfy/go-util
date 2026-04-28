package helper

// OneOfL 返回左值
func OneOfL[L, R any](data L, _ R) L {
    return data
}

// OneOfR 返回右值
func OneOfR[L, R any](_ L, data R) R {
    return data
}

// OneOf3L 返回三值中的左值
func OneOf3L[L, M, R any](data L, _ M, _ R) L {
    return data
}

// OneOf3M 返回三值中的中值
func OneOf3M[L, M, R any](_ L, data M, _ R) M {
    return data
}

// OneOf3R 返回三值中的右值
func OneOf3R[L, M, R any](_ L, _ M, data R) R {
    return data
}
