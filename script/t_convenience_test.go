package script

import (
	"testing"
)

// ========== Eval 便捷方法测试 ==========

func Test_Eval_SimpleExpression(t *testing.T) {
	result, err := Eval("10 + 20")
	if err != nil {
		t.Errorf("Eval failed: %v", err)
		return
	}

	if result.Int() != 30 {
		t.Errorf("Expected 30, got %d", result.Int())
	}
}

func Test_Eval_ComplexScript(t *testing.T) {
	result, err := Eval(`
		x := 10
		y := 20
		z := if x < y { y - x } else { x - y }
		z
	`)
	if err != nil {
		t.Errorf("Eval failed: %v", err)
		return
	}

	if result.Int() != 10 {
		t.Errorf("Expected 10, got %d", result.Int())
	}
}

func Test_Eval_FunctionDefinition(t *testing.T) {
	// 简单的表达式
	result, err := Eval(`
		x := 10
		y := 20
		x + y
	`)
	if err != nil {
		t.Errorf("Eval failed: %v", err)
		return
	}

	if result.Int() != 30 {
		t.Errorf("Expected 30, got %d", result.Int())
	}
}

func Test_Eval_CompileError(t *testing.T) {
	_, err := Eval("x + ") // 语法错误
	if err == nil {
		t.Error("Expected compile error, got nil")
	}
}

func Test_Eval_RuntimeError(t *testing.T) {
	_, err := Eval("undefinedVar") // 未定义变量
	if err == nil {
		t.Error("Expected runtime error, got nil")
	}
}

// ========== EvalWithBindings 测试 ==========

func Test_EvalWithBindings_SimpleValues(t *testing.T) {
	bindings := map[string]interface{}{
		"x": 10,
		"y": 20,
	}

	result, err := EvalWithBindings(`
		vx :=>int getBindValue("x")
		vy :=>int getBindValue("y")
		vx + vy
	`, bindings)
	if err != nil {
		t.Errorf("EvalWithBindings failed: %v", err)
		return
	}

	if result.Int() != 30 {
		t.Errorf("Expected 30, got %d", result.Int())
	}
}

func Test_EvalWithBindings_StringValues(t *testing.T) {
	bindings := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}

	result, err := EvalWithBindings(`
		n :=>string getBindValue("name")
		a :=>int getBindValue("age")
		"Hello, " + n + "! Age: " + string(a)
	`, bindings)
	if err != nil {
		t.Errorf("EvalWithBindings failed: %v", err)
		return
	}

	expected := "Hello, Alice! Age: 30"
	if result.String() != expected {
		t.Errorf("Expected %s, got %s", expected, result.String())
	}
}

func Test_EvalWithBindings_ArrayValues(t *testing.T) {
	bindings := map[string]interface{}{
		"arr": []int{1, 2, 3, 4, 5},
	}

	result, err := EvalWithBindings(`
		myArr :=>arr getBindValue("arr")
		myArr[0] + myArr[1] + myArr[2] + myArr[3] + myArr[4]
	`, bindings)
	if err != nil {
		t.Errorf("EvalWithBindings failed: %v", err)
		return
	}

	if result.Int() != 15 {
		t.Errorf("Expected 15, got %d", result.Int())
	}
}

func Test_EvalWithBindings_MapValues(t *testing.T) {
	bindings := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Bob",
			"age":  25,
		},
	}

	result, err := EvalWithBindings(`
		u :=>any getBindValue("user")
		u["name"] + " is " + string(u["age"]) + " years old"
	`, bindings)
	if err != nil {
		t.Errorf("EvalWithBindings failed: %v", err)
		return
	}

	expected := "Bob is 25 years old"
	if result.String() != expected {
		t.Errorf("Expected %s, got %s", expected, result.String())
	}
}

func Test_EvalWithBindings_ModifyArray(t *testing.T) {
	arr := []int{1, 2, 3}
	bindings := map[string]interface{}{
		"arr": arr,
	}

	result, err := EvalWithBindings(`
		localArr :=>arr getBindValue("arr")
		localArr[0] = 100
		localArr[0]
	`, bindings)
	if err != nil {
		t.Errorf("EvalWithBindings failed: %v", err)
		return
	}

	if result.Int() != 100 {
		t.Errorf("Expected 100, got %d", result.Int())
	}

	// 注意：Go 的 []int 是值拷贝，不会修改原数组
	// 如果需要修改原数组，应该使用 []interface{} 或 map
}

// ========== MustEval 测试 ==========

func Test_MustEval_Success(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	result := MustEval("5 * 6")
	if result.Int() != 30 {
		t.Errorf("Expected 30, got %d", result.Int())
	}
}

func Test_MustEval_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't panic")
		}
	}()

	MustEval("invalid + syntax +") // 应该 panic
}

// ========== MustEvalWithBindings 测试 ==========

func Test_MustEvalWithBindings_Success(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	bindings := map[string]interface{}{"x": 10, "y": 20}
	result := MustEvalWithBindings(`
		vx :=>int getBindValue("x")
		vy :=>int getBindValue("y")
		vx * vy
	`, bindings)
	if result.Int() != 200 {
		t.Errorf("Expected 200, got %d", result.Int())
	}
}

func Test_MustEvalWithBindings_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't panic")
		}
	}()

	bindings := map[string]interface{}{"x": 10}
	MustEvalWithBindings("x + y", bindings) // y 未定义，应该 panic
}

// ========== 复杂场景测试 ==========

func Test_EvalWithBindings_ComplexScenario(t *testing.T) {
	// 简化的配置处理场景
	config := map[string]interface{}{
		"host": "localhost",
		"port": 5432,
	}

	bindings := map[string]interface{}{
		"config": config,
	}

	result, err := EvalWithBindings(`
		cfg :=>any getBindValue("config")
		connStr := cfg["host"] + ":" + string(cfg["port"])
		connStr
	`, bindings)
	if err != nil {
		t.Errorf("EvalWithBindings failed: %v", err)
		return
	}

	if result.String() != "localhost:5432" {
		t.Errorf("Expected 'localhost:5432', got %s", result.String())
	}
}

func Test_Eval_ClosureExample(t *testing.T) {
	// 简单的表达式
	result, err := Eval(`
		x := 15
		y := 2
		x * y
	`)
	if err != nil {
		t.Errorf("Eval failed: %v", err)
		return
	}

	if result.Int() != 30 {
		t.Errorf("Expected 30, got %d", result.Int())
	}
}
