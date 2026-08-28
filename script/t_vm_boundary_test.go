package script

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// ========== VM边界条件测试 ==========

func Test_VM_MaxIntArithmetic(t *testing.T) {
	maxInt := strconv.FormatInt(math.MaxInt, 10)
	runIntTest(t, maxInt, math.MaxInt)
	runIntTest(t, "0 - "+maxInt, math.MinInt+1)
}

func Test_VM_ZeroOperations(t *testing.T) {
	runIntTest(t, "0 + 0", 0)
	runIntTest(t, "0 * 100", 0)
	runIntTest(t, "0 - 0", 0)
}

func Test_VM_OneOperations(t *testing.T) {
	runIntTest(t, "1 * 1", 1)
	runIntTest(t, "1 + 0", 1)
	runIntTest(t, "1 - 1", 0)
	runIntTest(t, "1 * 0", 0)
}

func Test_VM_LargeNumbers(t *testing.T) {
	// 期望值与引擎同为int回绕语义
	a := 1000000
	runIntTest(t, "1000000 * 1000000", a*a)
	runIntTest(t, "999999 + 1", 1000000)
}

func Test_VM_EmptyArray(t *testing.T) {
	result := runScript(t, "[]")
	arr := result.Array()
	if arr == nil || len(arr.Elements) != 0 {
		t.Error("空数组应长度为0")
	}
}

func Test_VM_EmptyMap(t *testing.T) {
	result := runScript(t, "{}")
	m := result.Map()
	if m == nil || len(m.Pairs) != 0 {
		t.Error("空map应大小为0")
	}
}

func Test_VM_SingleElementArray(t *testing.T) {
	runIntTest(t, "[42][0]", 42)
}

func Test_VM_SingleEntryMap(t *testing.T) {
	runIntTest(t, `{"key": 99}["key"]`, 99)
}

func Test_VM_DeeplyNestedArray(t *testing.T) {
	result := runScript(t, "[[[[[1]]]]]")
	arr := result.Array()
	for i := 0; i < 4; i++ {
		arr = arr.Elements[0].Array()
		if arr == nil {
			t.Fatalf("嵌套层级%d处不是数组", i)
		}
	}
	if arr.Elements[0].Int() != 1 {
		t.Errorf("最内层值应为1, got %d", arr.Elements[0].Int())
	}
}

// ========== 字符串操作测试 ==========

func Test_VM_StringConcatMultiple(t *testing.T) {
	result := runScript(t, `"a" + "b" + "c" + "d" + "e"`)
	if result.String() != "abcde" {
		t.Errorf("got %q, want %q", result.String(), "abcde")
	}
}

func Test_VM_EmptyStringConcat(t *testing.T) {
	runStringTest(t, `"" + ""`, "")
	runStringTest(t, `"" + "x"`, "x")
	runStringTest(t, `"x" + ""`, "x")
}

func Test_VM_StringInArray(t *testing.T) {
	result := runScript(t, `["a", "b", "c"]`)
	arr := result.Array()
	if len(arr.Elements) != 3 {
		t.Fatalf("got %d elements, want 3", len(arr.Elements))
	}
	if arr.Elements[0].String() != "a" {
		t.Errorf("element[0] = %q, want %q", arr.Elements[0].String(), "a")
	}
}

func Test_VM_StringInMap(t *testing.T) {
	result := runScript(t, `{"name": "Alice", "city": "NYC"}`)
	m := result.Map()
	if m.Pairs["name"].String() != "Alice" {
		t.Errorf("name = %q", m.Pairs["name"].String())
	}
}

// ========== 循环边界测试 ==========

func Test_VM_LoopZeroIterations(t *testing.T) {
	result := runScript(t, `
		sum := 0
		for i := 0; i < 0; i = i + 1 {
			sum = sum + i
		}
		sum
	`)
	if result.Int() != 0 {
		t.Errorf("零次循环sum应为0, got %d", result.Int())
	}
}

func Test_VM_LoopSingleIteration(t *testing.T) {
	result := runScript(t, `
		sum := 0
		for i := 0; i < 1; i = i + 1 {
			sum = sum + i
		}
		sum
	`)
	if result.Int() != 0 {
		t.Errorf("一次循环sum应为0(i=0), got %d", result.Int())
	}
}

func Test_VM_LoopBreakInFirstIteration(t *testing.T) {
	result := runScript(t, `
		count := 0
		for i := 0; i < 100; i = i + 1 {
			count = count + 1
			break
		}
		count
	`)
	if result.Int() != 1 {
		t.Errorf("break在第一次迭代后count应为1, got %d", result.Int())
	}
}

func Test_VM_NestedLoopBreak(t *testing.T) {
	result := runScript(t, `
		count := 0
		for i := 0; i < 3; i = i + 1 {
			for j := 0; j < 3; j = j + 1 {
				if j == 1 {
					break
				}
				count = count + 1
			}
		}
		count
	`)
	// 内层循环每次迭代j=0时count++, j=1时break
	// 外层循环3次, 每次内层count加1, 总计3
	if result.Int() != 3 {
		t.Errorf("嵌套break count应为3, got %d", result.Int())
	}
}

func Test_VM_LoopContinueAll(t *testing.T) {
	result := runScript(t, `
		sum := 0
		for i := 0; i < 5; i = i + 1 {
			continue
			sum = sum + i
		}
		sum
	`)
	if result.Int() != 0 {
		t.Errorf("全部continue后sum应为0, got %d", result.Int())
	}
}

// ========== 并发执行测试 ==========

func Test_VM_ConcurrentExecution(t *testing.T) {
	script := compileScript(t, `
		x := 10 + 20
		y := x * 2
		y
	`)

	ctx := NewContext()
	engine := NewEngine()

	var wg sync.WaitGroup
	results := make([]int, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			result, err := engine.Run(ctx, script)
			errors[idx] = err
			results[idx] = result.Int()
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		if errors[i] != nil {
			t.Errorf("goroutine %d error: %v", i, errors[i])
		}
		if results[i] != 60 {
			t.Errorf("goroutine %d result = %d, want 60", i, results[i])
		}
	}
}

func Test_Context_ConcurrentBindAndRun(t *testing.T) {
	ctx := NewContext()
	ctx.BindValue("base", 100)

	script := compileScript(t, `
		v :=>int getBindValue("base")
		v
	`)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine := NewEngine()
			result, err := engine.Run(ctx, script)
			if err != nil {
				t.Errorf("concurrent run error: %v", err)
			}
			if result.Int() != 100 {
				t.Errorf("got %d, want 100", result.Int())
			}
		}()
	}
	wg.Wait()
}

// ========== Value.GoString测试 ==========

func Test_Value_GoString(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{"int", NewValue(42)},
		{"float", NewValue(3.14)},
		{"string", NewValue("hello")},
		{"bool", NewValue(true)},
		{"nil", NewValue(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.value.GoString()
			if s == "" {
				t.Error("GoString不应为空")
			}
		})
	}
}

// ========== 错误消息质量测试 ==========

func Test_Error_ContainsLineNumber(t *testing.T) {
	_, err := NewParser().Compile("x :=\n  +")
	if err == nil {
		t.Fatal("应返回错误")
	}
	msg := err.Error()
	if !strings.Contains(msg, "行") {
		t.Errorf("错误消息应包含行号信息, got: %s", msg)
	}
}

func Test_Error_ContainsColumnNumber(t *testing.T) {
	_, err := NewParser().Compile("(1 + 2")
	if err == nil {
		t.Fatal("应返回错误")
	}
	msg := err.Error()
	if !strings.Contains(msg, "列") {
		t.Errorf("错误消息应包含列号信息, got: %s", msg)
	}
}

func Test_Error_ThrowContainsMessage(t *testing.T) {
	parser := NewParser()
	script, _ := parser.Compile(`throw "custom error"`)
	ctx := NewContext()
	engine := NewEngine()
	_, err := engine.Run(ctx, script)
	if err == nil {
		t.Fatal("throw应产生错误")
	}
	if !strings.Contains(err.Error(), "custom error") {
		t.Errorf("throw错误消息应包含原始消息, got: %s", err.Error())
	}
}

// ========== CompileError结构测试 ==========

func Test_CompileError_HasPosition(t *testing.T) {
	_, err := NewParser().Compile("if { }")
	if err == nil {
		t.Fatal("应返回编译错误")
	}

	// 检查是否为*CompileError
	if ce, ok := err.(*CompileError); ok {
		if ce.Line < 1 {
			t.Errorf("CompileError.Line应>=1, got %d", ce.Line)
		}
		if ce.Column < 1 {
			t.Errorf("CompileError.Column应>=1, got %d", ce.Column)
		}
		if ce.Message == "" {
			t.Error("CompileError.Message不应为空")
		}
	}
}
