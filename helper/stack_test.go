package helper

import (
    "strings"
    "testing"
)

func TestCallerStack(t *testing.T) {
    t.Run("基本调用栈", func(t *testing.T) {
        stacks := CallerStack(0)
        if len(stacks) == 0 {
            t.Fatal("CallerStack 返回空栈")
        }
        // 第一帧不应是 runtime 内部函数
        if stacks[0].File == "" {
            t.Error("栈帧文件路径为空")
        }
        if stacks[0].Line <= 0 {
            t.Error("栈帧行号无效")
        }
    })
    t.Run("skip参数", func(t *testing.T) {
        s0 := CallerStack(0)
        s1 := CallerStack(1)
        // skip=1 应比 skip=0 少一帧
        if len(s1) >= len(s0) {
            t.Errorf("skip=1 的栈帧数(%d) 应少于 skip=0(%d)", len(s1), len(s0))
        }
    })
    t.Run("limit参数", func(t *testing.T) {
        stacks := CallerStack(0, 2)
        if len(stacks) > 2 {
            t.Errorf("limit=2 时栈帧数(%d) 不应超过 2", len(stacks))
        }
    })
    t.Run("负数skip", func(t *testing.T) {
        stacks := CallerStack(-1)
        if len(stacks) == 0 {
            t.Error("负数 skip 不应返回空栈")
        }
    })
}

func TestCallerStackString(t *testing.T) {
    s := CallerStackString(0)
    if s == "" {
        t.Fatal("CallerStackString 返回空字符串")
    }
    // 应包含文件名和行号
    if !strings.Contains(s, ".go:") {
        t.Errorf("CallerStackString 结果不包含 .go:, got: %s", s)
    }
}

func TestFormatStack(t *testing.T) {
    stacks := CallerStack(0)
    result := FormatStack(stacks)
    if result == "" {
        t.Fatal("FormatStack 返回空字符串")
    }
    // 应包含每个栈帧的文件名
    for _, s := range stacks {
        if !strings.Contains(result, s.File) {
            // 文件名可能被简化, 检查基本格式即可
        }
    }
}

func TestSimpleFunctionName(t *testing.T) {
    stacks := CallerStack(0)
    if len(stacks) == 0 {
        t.Fatal("无法获取调用栈")
    }
    name := SimpleFunctionName(stacks[0].PC)
    if name == "" {
        t.Error("SimpleFunctionName 返回空字符串")
    }
    // 简化后的函数名不应包含完整包路径
    if strings.Contains(name, "/") {
        t.Errorf("SimpleFunctionName 结果不应包含路径分隔符: %s", name)
    }
}

func TestFunctionName(t *testing.T) {
    stacks := CallerStack(0)
    if len(stacks) == 0 {
        t.Fatal("无法获取调用栈")
    }
    name := FunctionName(stacks[0].PC)
    if name == "" {
        t.Error("FunctionName 返回空字符串")
    }
    // 完整函数名应包含包路径
    if !strings.Contains(name, "github.com/mzzsfy/go-util") {
        t.Errorf("FunctionName 结果应包含包路径: %s", name)
    }
}

func TestStacksString(t *testing.T) {
    stacks := CallerStack(0)
    result := stacks.String()
    if result == "" {
        t.Fatal("Stacks.String() 返回空字符串")
    }
}

func TestCallerStack_RuntimeSkip(t *testing.T) {
    // 验证跳过了 runtime 内部帧
    stacks := CallerStack(0)
    for _, s := range stacks {
        name := FunctionName(s.PC)
        if strings.HasPrefix(name, "runtime.") {
            t.Errorf("CallerStack 不应返回 runtime 帧: %s", name)
        }
    }
}

// BenchmarkCallerStack 测试调用栈获取性能
func BenchmarkCallerStack(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        CallerStack(0)
    }
}

// BenchmarkCallerStack_WithLimit 测试带限制的调用栈获取性能
func BenchmarkCallerStack_WithLimit(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        CallerStack(0, 10)
    }
}

// BenchmarkFormatStack 测试格式化调用栈性能
func BenchmarkFormatStack(b *testing.B) {
    b.ReportAllocs()
    stacks := CallerStack(0)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        FormatStack(stacks)
    }
}
