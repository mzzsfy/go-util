package script

import "testing"

func Test_TrailingComma_Array(t *testing.T) {
	cases := []struct{ src string; expected int }{
		{"[1, 2, 3,][0]", 1},
		{"[1, 2, 3,][2]", 3},
		{"[1,][0]", 1},
		{"len([1, 2, 3,])", 3},
		{"len([1,])", 1},
	}
	for _, tc := range cases {
		result, err := Eval(tc.src)
		if err != nil {
			t.Errorf("%s: %v", tc.src, err)
			continue
		}
		if got := result.Int(); got != tc.expected {
			t.Errorf("%s: 期望 %d, 实际 %d", tc.src, tc.expected, got)
		}
	}
}

func Test_TrailingComma_Map(t *testing.T) {
	cases := []struct{ src string; expected string }{
		{`{"a": 1, "b": 2,}["a"]`, "1"},
		{`{"a": 1,}["a"]`, "1"},
		{`m := {"a": 1, "b": 2,}
m["b"]`, "2"},
	}
	for _, tc := range cases {
		result, err := Eval(tc.src)
		if err != nil {
			t.Errorf("%s: %v", tc.src, err)
			continue
		}
		if got := result.GoString(); got != tc.expected {
			t.Errorf("%s: 期望 %s, 实际 %s", tc.src, tc.expected, got)
		}
	}
}

func Test_TrailingComma_NoTrailing(t *testing.T) {
	// 无尾随逗号仍然正常
	result, err := Eval("[1, 2, 3][1]")
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Int(); got != 2 {
		t.Fatalf("期望 2, 实际 %d", got)
	}
}

func Test_TrailingComma_Empty(t *testing.T) {
	// 空数组/Map不受影响
	result, err := Eval("len([])")
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Int(); got != 0 {
		t.Fatalf("期望 0, 实际 %d", got)
	}
}
