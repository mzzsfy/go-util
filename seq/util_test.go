package seq

import "testing"

func Test_LessT_GreatT(t *testing.T) {
	t.Parallel()
	if !LessT(1, 2) {
		t.Fatal("LessT(1,2) 应为 true")
	}
	if LessT(2, 1) {
		t.Fatal("LessT(2,1) 应为 false")
	}
	if GreatT(1, 2) {
		t.Fatal("GreatT(1,2) 应为 false")
	}
	if !GreatT(2, 1) {
		t.Fatal("GreatT(2,1) 应为 true")
	}
	if GreatT(1, 1) || LessT(1, 1) {
		t.Fatal("相等元素应返回 false")
	}
	if !LessT("a", "b") {
		t.Fatal("LessT(\"a\",\"b\") 应为 true")
	}
	if GreatT("a", "b") {
		t.Fatal("GreatT(\"a\",\"b\") 应为 false")
	}
	if !GreatT("b", "a") {
		t.Fatal("GreatT(\"b\",\"a\") 应为 true")
	}
}

func assertOrderSymmetry[T Comparable](t *testing.T, a, b T) {
	t.Helper()
	if LessT(a, b) != GreatT(b, a) {
		t.Fatalf("LessT/GreatT 不对称: %v %v", a, b)
	}
	if LessT(b, a) != GreatT(a, b) {
		t.Fatalf("LessT/GreatT 不对称: %v %v", b, a)
	}
}

func Test_LessT_GreatT_Symmetric(t *testing.T) {
	t.Parallel()
	assertOrderSymmetry(t, 1, 2)
	assertOrderSymmetry(t, int8(-1), int8(1))
	assertOrderSymmetry(t, int64(-100), int64(100))
	assertOrderSymmetry(t, uint8(0), uint8(255))
	assertOrderSymmetry(t, float32(0.5), float32(1.5))
	assertOrderSymmetry(t, 3.14, 2.71)
	assertOrderSymmetry(t, "a", "b")
	assertOrderSymmetry(t, "", "x")
}
