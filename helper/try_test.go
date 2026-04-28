package helper

import (
    "runtime"
    "strings"
    "testing"
)

func Test_TryWithStack_NoPanic(t *testing.T) {
    TryWithStack(func() {}, func(e any, stack Stacks) {
        t.Errorf("Unexpected error: %v", e)
    })
}

func Test_TryWithStack_PanicWithoutError(t *testing.T) {
    TryWithStack(func() { panic(nil) }, func(err any, stack Stacks) {
        if err != nil {
            t.Errorf("Expected nil error, got: %v", err)
        }
    })
}

func Test_TryWithStack_PanicWithError(t *testing.T) {
    expectedError := "test error"
    TryWithStack(func() { panic(expectedError) }, func(err any, stack Stacks) {
        if err != expectedError {
            t.Errorf("Expected error: %v, got: %v", expectedError, err)
        }
    })
}

func Test_TryWithStack_PanicWithStack(t *testing.T) {
    s := "Test_TryWithStack_PanicWithStack"
    t.Run("Test_TryWithStack_PanicWithStack", func(t *testing.T) {
        TryWithStack(func() {
            panic("test error")
        }, func(err any, stack Stacks) {
            //t.Log("\n", FormatStack(stack))
            name := SimpleFunctionName(stack[0].PC)
            i := len(s) + 5
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[1].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
        })
    })
    t.Run("Test_TryWithStack_PanicWithStack1", func(t *testing.T) {
        TryWithStack(func() {
            func() {
                panic("test error")
            }()
        }, func(err any, stack Stacks) {
            //t.Log("\n", FormatStack(stack))
            name := SimpleFunctionName(stack[0].PC)
            i := len(s) + 5
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[1].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[2].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
        })
    })
    t.Run("Test_TryWithStack_PanicWithStack2", func(t *testing.T) {
        TryWithStack(func() {
            func() {
                func() {
                    panic("test error")
                }()
            }()
        }, func(err any, stack Stacks) {
            //t.Log("\n", FormatStack(stack))
            name := SimpleFunctionName(stack[0].PC)
            i := len(s) + 5
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[1].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[2].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[3].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
        })
    })
}

// Test_TryWithStack_BoundaryRemoved 验证 TryWithStack 内部帧被正确移除
// 回调中拿到的栈不应包含 TryWithStack 函数自身的帧
func Test_TryWithStack_BoundaryRemoved(t *testing.T) {
    TryWithStack(func() {
        panic("boundary test")
    }, func(err any, stack Stacks) {
        if err != "boundary test" {
            t.Fatalf("got error %v, want %q", err, "boundary test")
        }
        // 栈中不应包含 helper.TryWithStack 函数本身的帧
        // 注意: 测试函数名含 "TryWithStack", 所以需要精确匹配包名前缀
        for _, s := range stack {
            name := runtime.FuncForPC(s.PC).Name()
            // 精确匹配 helper.TryWithStack 函数,而非调用它的测试函数
            if strings.HasSuffix(name, ".TryWithStack") || strings.HasSuffix(name, ".TryWithStack.func") {
                t.Errorf("栈中不应包含 TryWithStack 内部帧, 但找到: %s", name)
            }
        }
    })
}

// Test_TryWithStack_StackDepth 验证不同调用深度下栈帧数量正确
func Test_TryWithStack_StackDepth(t *testing.T) {
    tests := []struct {
        name      string
        depth     int
        minFrames int
    }{
        {"直接panic", 1, 1},
        {"嵌套1层", 2, 2},
        {"嵌套3层", 4, 4},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var captured Stacks
            // 构造指定深度的调用栈
            var recurse func(depth int, f func())
            recurse = func(depth int, f func()) {
                if depth <= 1 {
                    f()
                    return
                }
                recurse(depth-1, f)
            }
            TryWithStack(func() {
                recurse(tt.depth, func() { panic("depth test") })
            }, func(err any, stack Stacks) {
                captured = stack
            })
            if len(captured) < tt.minFrames {
                t.Errorf("栈帧数 = %d, 最少应有 %d", len(captured), tt.minFrames)
            }
        })
    }
}

// TestCleanTryFrame 验证 cleanTryFrame 的 boundary 缓存和移除逻辑
func TestCleanTryFrame(t *testing.T) {
    // cleanTryFrame 依赖 panic 上下文的 runtime.Callers, 无法用伪栈直接测试
    // 通过 Test_TryWithStack_BoundaryRemoved 实际调用验证
    // 此处验证空栈和无关栈不被修改
    var stack Stacks
    cleanTryFrame(&stack)
    if len(stack) != 0 {
        t.Fatal("空栈不应被修改")
    }
    stack2 := Stacks{{PC: uintptr(1)}, {PC: uintptr(2)}}
    cleanTryFrame(&stack2)
    if len(stack2) != 2 {
        t.Fatal("无匹配 PC 时栈不应被修改")
    }
}
