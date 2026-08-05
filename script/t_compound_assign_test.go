package script

import "testing"

func Test_CompoundAssign_Basic(t *testing.T) {
	cases := []struct{ name, src string; expected int }{
		{"plus", "x := 10\nx += 5\nx", 15},
		{"minus", "x := 10\nx -= 3\nx", 7},
		{"mul", "x := 10\nx *= 3\nx", 30},
		{"div", "x := 10\nx /= 2\nx", 5},
	}
	for _, tc := range cases {
		result, err := Eval(tc.src)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if got := result.Int(); got != tc.expected {
			t.Errorf("%s: 期望 %d, 实际 %d", tc.name, tc.expected, got)
		}
	}
}

func Test_CompoundAssign_Chained(t *testing.T) {
	src := "x := 1\nx += 2\nx += 3\nx *= 4\nx -= 5\nx"
	result, err := Eval(src)
	if err != nil {
		t.Fatal(err)
	}
	// 1+2=3, 3+3=6, 6*4=24, 24-5=19
	if got := result.Int(); got != 19 {
		t.Fatalf("期望 19, 实际 %d", got)
	}
}

func Test_CompoundAssign_Index(t *testing.T) {
	src := `
a := [1, 2, 3]
a[0] += 10
a[1] *= 5
a[2] -= 1
a[0] + a[1] + a[2]
`
	result, err := Eval(src)
	if err != nil {
		t.Fatal(err)
	}
	// a[0]=11, a[1]=10, a[2]=2 => 23
	if got := result.Int(); got != 23 {
		t.Fatalf("期望 23, 实际 %d", got)
	}
}

func Test_CompoundAssign_Map(t *testing.T) {
	src := `
m := {"c": 0}
m["c"] += 5
m["c"] *= 2
m["c"]
`
	result, err := Eval(src)
	if err != nil {
		t.Fatal(err)
	}
	// 0+5=5, 5*2=10
	if got := result.Int(); got != 10 {
		t.Fatalf("期望 10, 实际 %d", got)
	}
}

func Test_CompoundAssign_Expr(t *testing.T) {
	src := `
x := 10
x += 2 * 3
x
`
	result, err := Eval(src)
	if err != nil {
		t.Fatal(err)
	}
	// 10 + 6 = 16
	if got := result.Int(); got != 16 {
		t.Fatalf("期望 16, 实际 %d", got)
	}
}

func Test_CompoundAssign_InLoop(t *testing.T) {
	src := `
sum := 0
for i := 5 {
  sum += i
}
sum
`
	result, err := Eval(src)
	if err != nil {
		t.Fatal(err)
	}
	// 1+2+3+4+5 = 15
	if got := result.Int(); got != 15 {
		t.Fatalf("期望 15, 实际 %d", got)
	}
}
