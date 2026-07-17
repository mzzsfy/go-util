package helper

import (
    "testing"
)

func Test_IsZero(t *testing.T) {
    tests := []struct {
        name  string
        input any
        want  bool
    }{
        {"nil", nil, true},
        {"零值int", 0, true},
        {"非零int", 1, false},
        {"零值string", "", true},
        {"非零string", "a", false},
        {"零值slice", []int(nil), true},
        {"非零slice", []int{}, false},
        {"零值map", map[string]int(nil), true},
        {"非零map", map[string]int{}, false},
        {"零值指针", (*int)(nil), true},
        {"非零指针", Ptr(1), false},
        {"零值bool", false, true},
        {"非零bool", true, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsZero(tt.input)
            if got != tt.want {
                t.Errorf("IsZero(%v) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}

func Test_Ptr(t *testing.T) {
    val := 42
    p := Ptr(val)
    if p == nil {
        t.Fatal("Ptr 返回 nil")
    }
    if *p != val {
        t.Errorf("Ptr(%d) = %d, want %d", val, *p, val)
    }
    // 验证返回的是副本地址
    val = 100
    if *p != 42 {
        t.Errorf("Ptr 应返回值的副本, 修改原值不应影响指针指向的值")
    }
}

func Test_New(t *testing.T) {
    t.Run("非指针类型", func(t *testing.T) {
        got := New(0)
        if got != 0 {
            t.Errorf("New(int) = %v, want 0", got)
        }
    })
    t.Run("指针类型", func(t *testing.T) {
        got := New((*int)(nil))
        if got == nil {
            t.Fatal("New(*int) 返回 nil")
        }
        if *got != 0 {
            t.Errorf("New(*int) 解引用 = %v, want 0", *got)
        }
    })
    t.Run("结构体类型", func(t *testing.T) {
        type testStruct struct {
            A int
            B string
        }
        got := New(testStruct{})
        if got.A != 0 || got.B != "" {
            t.Errorf("New(struct) = %+v, want 零值", got)
        }
    })
}
