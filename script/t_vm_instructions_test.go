package script

import (
	"testing"
)

// ========== VM 指令测试 ==========

func Test_vm_UnaryOps(t *testing.T) {
	t.Run("Neg", func(t *testing.T) { runIntTest(t, "-10", -10) })
	t.Run("Not", func(t *testing.T) { runBoolTest(t, "!true", false) })
	t.Run("BitNot", func(t *testing.T) { runIntTest(t, "^15", ^15) })
}

func Test_vm_StringConcat(t *testing.T) {
	runStringTest(t, `"hello" + " " + "world"`, "hello world")
}

func Test_vm_FloatArithmetic(t *testing.T) {
	tests := []struct {
		Input    string
		Expected float64
	}{
		{"10.5 + 20.3", 30.8},
		{"30.7 - 10.2", 20.5},
		{"5.0 * 4.0", 20.0},
		{"20.0 / 4.0", 5.0},
	}
	RunFloatTestsSimple(t, tests)
}

func Test_vm_NestedExpressions(t *testing.T) {
	runIntTest(t, "(10 + 20) * (5 - 2)", 90)
}

func Test_vm_ArrayIndex(t *testing.T) {
	input := `arr := [10, 20, 30]
arr[1]`
	runIntTest(t, input, 20)
}

func Test_vm_MapAccess(t *testing.T) {
	input := `m := {"a": 10, "b": 20}
m["a"]`
	runIntTest(t, input, 10)
}

func Test_vm_LenVariants(t *testing.T) {
	tests := []struct {
		Input    string
		Expected int
	}{
		{"len([1, 2, 3])", 3},
		{`len("hello")`, 5},
		{`len([])`, 0},
		{`len({})`, 0},
		{`len("")`, 0},
	}
	RunIntTestsSimple(t, tests)
}

func Test_vm_TypeOfVariants(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{"typeof(10)", "int"},
		{"typeof(3.14)", "float"},
		{`typeof("hello")`, "string"},
		{"typeof(true)", "bool"},
		{"typeof(nil)", "nil"},
	}
	RunStringTestsSimple(t, tests)
}
