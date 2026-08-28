//go:build go1.27

package seq

import "testing"

// 场景: Seq.JoinL 方法形态,右侧转换并入
func Test_Op127_SeqJoinL(t *testing.T) {
	t.Parallel()
	r := FromT(1, 2).JoinL(FromT("aa", "b"), func(s string) int { return len(s) }).ToSlice()
	assertSlice(t, r, []int{1, 2, 2, 1})
}

// 场景: MergeBiR 流元素占K位,迭代器值占V位
func Test_Op127_MergeBiR(t *testing.T) {
	t.Parallel()
	merged := FromT("a", "b").MergeBiR(IteratorInt(100))
	k, v := merged.First()
	if *k != "a" || *v != 100 {
		t.Fatalf("MergeBiR配对错误:%v %v", *k, *v)
	}
	if c := merged.Count(); c != 2 {
		t.Fatalf("元素数应2,实际%d", c)
	}
}

// 场景: BiSeq.MapV 方法形态,V转换K不变
func Test_Op127_BiMapV(t *testing.T) {
	t.Parallel()
	k, v := BiFromT("key", 1).MapV(func(k string, v int) bool { return v > 0 }).First()
	if *k != "key" || !*v {
		t.Fatalf("MapV转换错误:%v %v", *k, *v)
	}
}

// 场景: BiSeq.Reduce 强类型聚合
func Test_Op127_BiReduce(t *testing.T) {
	t.Parallel()
	r := BiFromSeq(FromT("a", "bb"), func(v string) (int, string) { return len(v), v }).
		Reduce(func(k int, v string, acc int) int { return acc + k }, 0)
	if r != 3 {
		t.Fatalf("BiReduce聚合错误:%d", r)
	}
}

// 场景: Enumerate 缺省Range从0开始
func Test_Op127_Enumerate_Default(t *testing.T) {
	t.Parallel()
	idx, v := FromT("a", "b").Enumerate().Last()
	if *idx != 1 || *v != "b" {
		t.Fatalf("缺省Enumerate错误:%d %s", *idx, *v)
	}
}
