//go:build go1.27

package seq

import (
	"testing"
	"time"
)

// 场景: 强类型 Map 链 + 推断消除显式类型参数
// Given 一个Seq[int], When Map[string]转换后直接拼接,
// Then 返回Seq[string]且无需CastAny恢复类型
func Test_Method_Map_TypedChain(t *testing.T) {
	t.Parallel()
	out := FromSlice([]int{1, 2, 3}).Map(func(i int) string {
		if i == 1 {
			return "1"
		}
		return "n"
	})
	s := out.JoinString("-")
	if s != "1-n-n" {
		t.Fatalf("类型化Map链结果错误: %s", s)
	}
	// 显式实例化形态
	out2 := FromSlice([]int{2}).Map[string](func(i int) string { return "x" })
	if v := out2.First(); *v != "x" {
		t.Fatal("显式类型参数实例化失败")
	}
}

// 场景: Cast 消灭哑参
// Given Seq[any], When Cast[int](),
// Then 得到Seq[int]
func Test_Method_Cast(t *testing.T) {
	t.Parallel()
	anySeq := FromSlice([]any{"a", "b", "c"}).Map(func(s any) any { return len(s.(string)) })
	out := anySeq.Cast[int]()
	sum := out.SumBy(func(i int) int { return i })
	if sum != 3 {
		t.Fatalf("Cast后求和错误: %d", sum)
	}
}

// 场景: GroupBy 强键类型
func Test_Method_GroupBy(t *testing.T) {
	t.Parallel()
	m := FromSlice([]string{"ab", "ac", "ba"}).GroupBy(func(s string) byte { return s[0] })
	if len(m) != 2 || len(m['a']) != 2 || len(m['b']) != 1 {
		t.Fatalf("强类型GroupBy分组错误: %v", m)
	}
	first := FromSlice([]string{"x1", "x2"}).GroupByFirst(func(s string) string { return s[:1] })
	if len(first) != 1 {
		t.Fatal("GroupByFirst错误")
	}
	last := FromSlice([]string{"x1", "x2"}).GroupByLast(func(s string) string { return s[:1] })
	if last["x"] != "x2" {
		t.Fatal("GroupByLast错误")
	}
}

// 场景: Reduce 泛型累加器
func Test_Method_Reduce(t *testing.T) {
	t.Parallel()
	totalLen := FromSlice([]string{"ab", "cd"}).Reduce(func(s string, acc int) int {
		return acc + len(s)
	}, 0)
	if totalLen != 4 {
		t.Fatalf("Reduce聚合错误: %d", totalLen)
	}
}

// 场景: BiCast 双向收敛
func Test_Method_BiCast(t *testing.T) {
	t.Parallel()
	src := BiFromMap(map[int]int{1: 10, 2: 20}).MapVParallel(func(k, v int) any { return v }, 3, 2)
	out := src.Cast[int, int]()
	m := BiToMap(out)
	if m[1] != 10 || m[2] != 20 {
		t.Fatalf("BiSeq强类型链路损坏: %v", m)
	}
	mk := BiFromTuple(BiTuple[int, string]{K: 1, V: "v"}).MapK(func(k int, v string) string { return "k" })
	if k, _ := mk.First(); *k != "k" {
		t.Fatal("MapK转换错误")
	}
	flat := BiFromT(0, []string{"a", "b"}).MapFlatV(func(k int, vs []string) Seq[string] {
		return FromSlice(vs)
	})
	if c := flat.Count(); c != 2 {
		t.Fatalf("MapFlatV扁平化数量错误: %d", c)
	}
	single := BiFromT(0, []int{1, 2, 3}).MapFlatSingle(func(k int, vs []int) Seq[int] { return FromSlice(vs) })
	if single.Count() != 3 {
		t.Fatal("MapFlatSingle错误")
	}
	joined := BiFrom(BiFromTuple(BiTuple[int, string]{K: 1, V: "one"}).JoinL(BiFromT("L", "two"), func(k2, v2 string) (int, string) {
		return len(v2), v2
	}))
	if _, v := joined.FirstOrF(nil); v != "one" {
		t.Fatal("JoinL合并错误")
	}
	if n := BiFromT(1, 1.5).Map(func(k int, v float64) (string, bool) { return "ok", v > 1 }).Count(); n != 1 {
		t.Fatal("Map双参数转换错误")
	}
}

// 场景: to_bi 单方法覆盖原特化家族
func Test_Method_ToBi(t *testing.T) {
	t.Parallel()
	merged := FromSlice([]int{7, 8}).MergeBi(IteratorInt())
	firstK, firstV := merged.First()
	if *firstK != 0 || *firstV != 7 {
		t.Fatalf("MergeBi配对错误: %v %v", *firstK, *firstV)
	}
	byKey := FromSlice([]string{"go", "go-util"}).MapBi(func(s string) int { return len(s) })
	k, _ := byKey.Last()
	if *k != 7 {
		t.Fatalf("MapBi键生成错误: %d", *k)
	}
	rightVal := FromSlice([]byte{'A'}).MapBiR(func(b byte) string { return string(b) })
	kt, vt := rightVal.First()
	if *kt != 'A' || *vt != "A" {
		t.Fatal("MapBiR右侧值错误")
	}
	enum := FromSlice([]string{"a", "b", "c"}).Enumerate(100)
	idx, ev := enum.Last()
	if *idx != 102 || *ev != "c" {
		t.Fatalf("Enumerate序列号错误: %d", *idx)
	}
}

// 场景: 并行版本保持序约束语义
// Given 带 order=2 的 MapParallel[R],When 后发射的元素先完成,
// Then 消费仍严格按生产顺序,E 保持强类型
func Test_Method_MapParallel_Ordered(t *testing.T) {
	t.Parallel()
	consumed := make([]int, 0)
	done := make(chan struct{})
	go func() {
		defer close(done)
		FromSlice([]int{40, 20, 10}).
			MapParallel(func(v int) int {
				// 越先发射耗时越长,制造真实的乱序完成窗口
				sleeps := map[int]int{40: 40, 20: 25, 10: 10}
				time.Sleep(time.Millisecond * time.Duration(sleeps[v]))
				return v
			}, 2, 3).
			ForEach(func(v int) {
				consumed = append(consumed, v)
			})
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("order=2并行消费超时")
	}
	expect := []int{40, 20, 10}
	if len(consumed) != len(expect) {
		t.Fatalf("order=2消费数量不符: got=%v want=%v", consumed, expect)
	}
	for i, v := range expect {
		if consumed[i] != v {
			t.Fatalf("order=2顺序被破坏: got=%v want=%v", consumed, expect)
		}
	}
}

// 场景: MapFlatBi/MapFlatK/MapFlatV 承接旧轨扁平化家族
func Test_Method_MapFlatBi_K_V(t *testing.T) {
	t.Parallel()
	bi := FromSlice([]string{"ab"}).MapFlatBi(func(s string) BiSeq[byte, int] {
		return FromSlice([]byte(s)).MapFlatV(func(b byte) Seq[int] { return FromT(int(b)) })
	})
	if k, v := bi.First(); *k != 'a' || *v != 'a' {
		t.Fatal("MapFlatBi转换错误")
	}
	flatK := FromSlice([]string{"go", "r"}).MapFlatK(func(s string) Seq[rune] {
		return FromSlice([]rune(s))
	})
	if c := flatK.Count(); c != 3 {
		t.Fatalf("MapFlatK数量错误: %d", c)
	}
	if k, _ := flatK.Last(); *k != 'r' {
		t.Fatal("MapFlatK展开值占K位错误")
	}
	flatV := FromSlice([]string{"go", "r"}).MapFlatV(func(s string) Seq[rune] {
		return FromSlice([]rune(s))
	})
	var lastPair rune
	var lastSrc string
	flatV.ForEach(func(src string, e rune) {
		lastSrc = src
		lastPair = e
	})
	if lastSrc != "r" || lastPair != 'r' {
		t.Fatalf("MapFlatV配对错误: %s %c", lastSrc, lastPair)
	}
}

// 场景: JoinBy 双边收敛为新类型
func Test_Method_JoinBy(t *testing.T) {
	t.Parallel()
	n := FromSlice([]int{1}).JoinBy(
		FromSlice([]float64{2.5}),
		func(i int) string { return "i" },
		func(f float64) string { return "f" },
	).Count()
	if n != 2 {
		t.Fatalf("JoinBy数量错误: %d", n)
	}
	bn := BiFromT(1, 1).JoinBy(
		BiFromT("k", "v"),
		func(k, v int) (bool, bool) { return true, true },
		func(k, v string) (bool, bool) { return false, false },
	).Count()
	if bn != 2 {
		t.Fatalf("BiSeq JoinBy数量错误: %d", bn)
	}
}

// 场景: MapFlat 与 MapParallelCustomize 强类型扁平化和异步转换
func Test_Method_MapFlat_Customize(t *testing.T) {
	t.Parallel()
	flat := FromSlice([]int{1, 4}).MapFlat(func(i int) Seq[rune] {
		rs := make([]rune, 0, i)
		for j := 0; j < i; j++ {
			rs = append(rs, 'x')
		}
		return FromSlice(rs)
	})
	if c := flat.Count(); c != 5 {
		t.Fatalf("MapFlat计数错误: %d", c)
	}
	customized := FromSlice([]int{3, 4}).MapParallelCustomize(func(v int, push func(int)) {
		push(v * 2)
	})
	if s := customized.SumBy(func(v int) int { return v }); s != 14 {
		t.Fatalf("MapParallelCustomize求和错误: %d", s)
	}
}
