package script

import "testing"

// Test_NestedIfValue 嵌套if表达式的值保留
func Test_NestedIfValue(t *testing.T) {
	cases := []struct{ name, src string; expected string }{
		{"nested_if_else", "if true { if true { 1 } else { 2 } } else { 3 }", "1"},
		{"nested_if_false", "if false { 1 } else { if true { 2 } else { 3 } }", "2"},
		{"triple_nested", "if true { if true { if true { 42 } else { 0 } } else { 0 } } else { 0 }", "42"},
		{"fn_nested_if", "fn f() { if true { if true { 42 } else { 0 } } else { 0 } }\nf()", "42"},
	}
	for _, tc := range cases {
		result, err := Eval(tc.src)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if got := result.GoString(); got != tc.expected {
			t.Errorf("%s: 期望 %s, 实际 %s", tc.name, tc.expected, got)
		}
	}
}
