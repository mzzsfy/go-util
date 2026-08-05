package script

import (
	"bytes"
	"strings"
	"testing"
)

// runScriptWithOutput 执行脚本并捕获输出
func runScriptWithOutput(t *testing.T, input string) (Value, string) {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	var buf bytes.Buffer
	ctx := NewContext()
	ctx.SetOutput(&buf)
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	return result, buf.String()
}

// Test_Print_NoOutput 默认输出到stdout不panic
func Test_Print_NoOutput(t *testing.T) {
	result := runScript(t, `print("hello")`)
	if !result.IsNil() {
		t.Errorf("print应返回nil, 得到%v", result)
	}
}

// Test_Print_Output 输出到buffer
func Test_Print_Output(t *testing.T) {
	_, output := runScriptWithOutput(t, `print("hello", 123)`)
	expected := "hello 123"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Print_NoNewline print不加换行
func Test_Print_NoNewline(t *testing.T) {
	_, output := runScriptWithOutput(t, `print("abc")`)
	if strings.HasSuffix(output, "\n") {
		t.Errorf("print不应有换行, 得到%q", output)
	}
}

// Test_Println_Output println输出带换行
func Test_Println_Output(t *testing.T) {
	_, output := runScriptWithOutput(t, `println("hello", 123)`)
	expected := "hello 123\n"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Print_NoArgs print()无参数不报错
func Test_Print_NoArgs(t *testing.T) {
	_, output := runScriptWithOutput(t, `print()`)
	if output != "" {
		t.Errorf("无参数print应输出空字符串, 得到%q", output)
	}
}

// Test_Println_NoArgs println()无参数输出换行
func Test_Println_NoArgs(t *testing.T) {
	_, output := runScriptWithOutput(t, `println()`)
	if output != "\n" {
		t.Errorf("无参数println应输出换行, 得到%q", output)
	}
}

// Test_Print_MixedTypes 混合类型输出
func Test_Print_MixedTypes(t *testing.T) {
	_, output := runScriptWithOutput(t, `print(1, "a", true, nil)`)
	expected := "1 a true nil"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Print_StringNoQuote 字符串不加引号
func Test_Print_StringNoQuote(t *testing.T) {
	_, output := runScriptWithOutput(t, `print("hello world")`)
	expected := "hello world"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Print_ArrayType 数组输出
func Test_Print_ArrayType(t *testing.T) {
	_, output := runScriptWithOutput(t, `print([1, 2, 3])`)
	expected := "[1, 2, 3]"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Print_MapType map输出
func Test_Print_MapType(t *testing.T) {
	_, output := runScriptWithOutput(t, `print({"a": 1})`)
	if !strings.Contains(output, "a") || !strings.Contains(output, "1") {
		t.Errorf("map输出应包含a和1, 得到%q", output)
	}
}

// Test_Print_ReturnsNil print返回nil
func Test_Print_ReturnsNil(t *testing.T) {
	result, _ := runScriptWithOutput(t, `print(1)
nil`)
	if !result.IsNil() {
		t.Errorf("print应返回nil")
	}
}

// Test_Print_MultipleCalls 多次print连续输出
func Test_Print_MultipleCalls(t *testing.T) {
	_, output := runScriptWithOutput(t, `print("a")
print("b")
print("c")
nil`)
	expected := "abc"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Println_MultipleCalls 多次println各带换行
func Test_Println_MultipleCalls(t *testing.T) {
	_, output := runScriptWithOutput(t, `println("a")
println("b")
nil`)
	expected := "a\nb\n"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Print_FloatType 浮点数输出
func Test_Print_FloatType(t *testing.T) {
	_, output := runScriptWithOutput(t, `print(1.5)`)
	expected := "1.5"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}

// Test_Print_BoolType 布尔输出
func Test_Print_BoolType(t *testing.T) {
	_, output := runScriptWithOutput(t, `print(true, false)`)
	expected := "true false"
	if output != expected {
		t.Errorf("期望%q, 得到%q", expected, output)
	}
}
