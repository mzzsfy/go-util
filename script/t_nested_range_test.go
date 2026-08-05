package script

import "testing"

// Test_NestedRange_TmpVarConflict 嵌套for-range临时变量不应冲突
func Test_NestedRange_TmpVarConflict(t *testing.T) {
	src := `
sum := 0
outer := [1, 2, 3]
for v := outer {
  inner := [10, 20]
  for w := inner {
    sum := sum + v * w
  }
}
sum
`
	result, err := Eval(src)
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}
	// (1*10+1*20) + (2*10+2*20) + (3*10+3*20) = 30+60+90 = 180
	if got, exp := result.Int(), 180; got != exp {
		t.Fatalf("期望 %d, 实际 %d", exp, got)
	}
}

// Test_NestedCountFor_TmpVarConflict 嵌套计数循环
func Test_NestedCountFor_TmpVarConflict(t *testing.T) {
	src := `
sum := 0
for i := 3 {
  for j := 2 {
    sum := sum + i * j
  }
}
sum
`
	result, err := Eval(src)
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}
	// i=1,2,3; j=1,2; 1+2+2+4+3+6 = 18
	if got, exp := result.Int(), 18; got != exp {
		t.Fatalf("期望 %d, 实际 %d", exp, got)
	}
}

// Test_NestedMixedFor_TmpVarConflict 外层计数 + 内层range
func Test_NestedMixedFor_TmpVarConflict(t *testing.T) {
	src := `
sum := 0
arr := [100, 200]
for i := 3 {
  for v := arr {
    sum := sum + i * v
  }
}
sum
`
	result, err := Eval(src)
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}
	// i=1:300, i=2:600, i=3:900 => 1800
	if got, exp := result.Int(), 1800; got != exp {
		t.Fatalf("期望 %d, 实际 %d", exp, got)
	}
}

// Test_TripleNestedFor_TmpVarConflict 三层嵌套循环
func Test_TripleNestedFor_TmpVarConflict(t *testing.T) {
	src := `
sum := 0
a := [1, 2]
b := [10]
c := [100]
for x := a {
  for y := b {
    for z := c {
      sum := sum + x + y + z
    }
  }
}
sum
`
	result, err := Eval(src)
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}
	// x=1: y=10: z=100 => 111
	// x=2: y=10: z=100 => 112
	// total = 223
	if got, exp := result.Int(), 223; got != exp {
		t.Fatalf("期望 %d, 实际 %d", exp, got)
	}
}
