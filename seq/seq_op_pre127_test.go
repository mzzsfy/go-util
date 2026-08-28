//go:build !go1.27

package seq

import "testing"

// 场景: GroupBy 家族 any键形态
func Test_OpPre_GroupBy(t *testing.T) {
	t.Parallel()
	m := FromT(1, 2, 3, 4).GroupBy(func(i int) any { return i % 2 })
	if len(m) != 2 || len(m[0]) != 2 || len(m[1]) != 2 {
		t.Fatalf("GroupBy分组错误:%v", m)
	}
	first := FromT(2, 4, 3).GroupByFirst(func(i int) any { return i % 2 })
	if first[0] != 2 || first[1] != 3 {
		t.Fatalf("GroupByFirst错误:%v", first)
	}
	last := FromT(2, 4, 3).GroupByLast(func(i int) any { return i % 2 })
	if last[0] != 4 || last[1] != 3 {
		t.Fatalf("GroupByLast错误:%v", last)
	}
}

// 场景: MapString/MapInt 定向转换
func Test_OpPre_MapStringInt(t *testing.T) {
	t.Parallel()
	s := FromT(1, 2).MapString(func(i int) string { return string(rune('a' + i - 1)) }).JoinString("")
	if s != "ab" {
		t.Fatalf("MapString错误:%s", s)
	}
	assertSlice(t, FromT("aa", "bbb").MapInt(func(s string) int { return len(s) }).ToSlice(), []int{2, 3})
}

// 场景: MapFlatString 扁平化
func Test_OpPre_MapFlatString(t *testing.T) {
	t.Parallel()
	r := MapFlatString(FromT("ab", "c"), func(s string) Seq[string] {
		var out []string
		for _, r := range s {
			out = append(out, string(r))
		}
		return FromSlice(out)
	}).JoinString("")
	if r != "abc" {
		t.Fatalf("MapFlatString错误:%s", r)
	}
}

// 场景: 包级 JoinL/JoinBy 合并转换
func Test_OpPre_JoinLJoinBy(t *testing.T) {
	t.Parallel()
	r := JoinL(FromT(1, 2), FromT("aa"), func(s string) int { return len(s) }).ToSlice()
	assertSlice(t, r, []int{1, 2, 2})
	r2 := JoinBy(FromT(1), func(i int) string { return "i" }, FromT("x"), func(s string) string { return "s" }).ToSlice()
	assertSlice(t, r2, []string{"i", "s"})
}

// sliceIter 切片转迭代器
func sliceIter[T any](arr []T) Iterator[T] {
	i := 0
	return func() (T, bool) {
		if i >= len(arr) {
			var zero T
			return zero, false
		}
		v := arr[i]
		i++
		return v, true
	}
}

// biCollect 收集BiSeq为ks,vs
func biCollect[K, V any](t BiSeq[K, V]) (ks []K, vs []V) {
	t(func(k K, v V) {
		ks = append(ks, k)
		vs = append(vs, v)
	})
	return
}

// 场景: MergeBi 家族配对方向
func Test_OpPre_MergeBi(t *testing.T) {
	t.Parallel()
	src := FromT("a", "b")
	mk, mv := MergeBiString(src, sliceIter([]string{"x", "y"})).First()
	if *mk != "x" || *mv != "a" {
		t.Fatalf("MergeBiString配对错误:%v %v", *mk, *mv)
	}
	rk, rv := MergeBiIntRight(src, sliceIter([]int{7, 8})).First()
	if *rk != "a" || *rv != 7 {
		t.Fatalf("MergeBiIntRight配对错误:%v %v", *rk, *rv)
	}
	sk, sv := MergeBiStringRight(src, sliceIter([]string{"z"})).First()
	if *sk != "a" || *sv != "z" {
		t.Fatalf("MergeBiStringRight配对错误:%v %v", *sk, *sv)
	}
	ak, av := MergeBiAny(src, sliceIter([]any{1})).First()
	if *ak != any(1) || *av != any("a") {
		t.Fatalf("MergeBiAny配对错误:%v %v", *ak, *av)
	}
	//迭代器耗尽时提前停止
	ks, _ := biCollect(MergeBiString(FromT("a", "b", "c"), sliceIter([]string{"x"})))
	assertSlice(t, ks, []string{"x"})
}

// 场景: MapBi 家族生成键方向
func Test_OpPre_MapBi(t *testing.T) {
	t.Parallel()
	src := FromT("aaa")
	mk, mv := MapBiString(src, func(s string) string { return s + "k" }).First()
	if *mk != "aaak" || *mv != "aaa" {
		t.Fatalf("MapBiString错误:%v %v", *mk, *mv)
	}
	ak, av := MapBiAny(src, func(s string) any { return len(s) }).First()
	if *ak != any(3) || *av != any("aaa") {
		t.Fatalf("MapBiAny错误:%v %v", *ak, *av)
	}
	rk, rv := MapBiAnyRight(src, func(s string) any { return s + "r" }).First()
	if *rk != any("aaa") || *rv != any("aaar") {
		t.Fatalf("MapBiAnyRight错误:%v %v", *rk, *rv)
	}
}

// 场景: BiMapFlatK/V/Single 扁平化方向,pre127形态为Seq[any]
func Test_OpPre_BiMapFlat(t *testing.T) {
	t.Parallel()
	bi := BiFromT(1, "v")
	ks, vs := biCollect(BiMapFlatK(bi, func(k int, _ string) Seq[any] { return FromT[any](k, k*2) }))
	if len(ks) != 2 || ks[0] != any(1) || ks[1] != any(2) {
		t.Fatalf("BiMapFlatK展开值错误:%v", ks)
	}
	if len(vs) != 2 || vs[0] != "v" || vs[1] != "v" {
		t.Fatalf("BiMapFlatK保留V错误:%v", vs)
	}
	ks2, vs2 := biCollect(BiMapFlatV(bi, func(k int, _ string) Seq[any] { return FromT[any](k, k*2) }))
	if len(ks2) != 2 || ks2[0] != 1 || ks2[1] != 1 {
		t.Fatalf("BiMapFlatV保留K错误:%v", ks2)
	}
	if len(vs2) != 2 || vs2[0] != any(1) || vs2[1] != any(2) {
		t.Fatalf("BiMapFlatV展开值错误:%v", vs2)
	}
	//Single 丢弃KV仅保留展开值
	single := BiMapFlatSingle(bi, func(k int, _ string) Seq[any] { return FromT[any](k, k*2) }).ToSlice()
	if len(single) != 2 || single[0] != any(1) || single[1] != any(2) {
		t.Fatalf("BiMapFlatSingle错误:%v", single)
	}
}

// 场景: BiJoinL/BiJoinBy 合并转换,BiSeq.JoinBy 方法any形态
func Test_OpPre_BiJoinLJoinBy(t *testing.T) {
	t.Parallel()
	ks, vs := biCollect(BiJoinL(BiFromT(1, "a"), BiFromT("xx", true), func(k string, v bool) (int, string) {
		return len(k), "c"
	}))
	assertSlice(t, ks, []int{1, 2})
	assertSlice(t, vs, []string{"a", "c"})
	//方法形态 JoinBy(any源并入)
	ks, vs = biCollect(BiFromT(1, "a").JoinBy(BiFromT[any, any](9, 9), func(k, v any) (int, string) {
		return int(k.(int)) * 2, "j"
	}))
	assertSlice(t, ks, []int{1, 18})
	assertSlice(t, vs, []string{"a", "j"})
	//包级 BiJoinBy 双边收敛
	ks2, vs2 := biCollect(BiJoinBy(BiFromT(1, "a"), func(k int, v string) (string, int) { return v, k },
		BiFromT(true, 2.5), func(k bool, v float64) (string, int) { return "b", int(v) }))
	assertSlice(t, ks2, []string{"a", "b"})
	assertSlice(t, vs2, []int{1, 2})
}

// 场景: MapFlatBi 家族,Seq扁平化为BiSeq
func Test_OpPre_MapFlatBi(t *testing.T) {
	t.Parallel()
	src := FromT("ab")
	ks, vs := biCollect(MapFlatBi(src, func(s string) BiSeq[byte, int] {
		return BiFromTuple(BiTuple[byte, int]{K: s[0], V: len(s)})
	}))
	assertSlice(t, ks, []byte{'a'})
	assertSlice(t, vs, []int{2})
	//MapFlatBiK 展开值占K,原元素占V
	ks2, vs2 := biCollect(MapFlatBiK(src, func(s string) BiSeq[byte, string] {
		return BiFromTuple(BiTuple[byte, string]{K: s[0], V: s})
	}))
	assertSlice(t, ks2, []byte{'a'})
	assertSlice(t, vs2, []string{"ab"})
	//MapFlatBiV 展开值占V,原元素占K
	ks3, vs3 := biCollect(MapFlatBiV(src, func(s string) BiSeq[string, int] {
		return BiFromTuple(BiTuple[string, int]{K: s, V: len(s)})
	}))
	assertSlice(t, ks3, []string{"ab"})
	assertSlice(t, vs3, []int{2})
	//特化家族方向验证
	ks4, vs4 := biCollect(MapFlatBiInt(src, func(s string) Seq[int] { return FromT(len(s)) }))
	assertSlice(t, ks4, []int{2})
	assertSlice(t, vs4, []string{"ab"})
	ks5, vs5 := biCollect(MapFlatBiString(src, func(s string) Seq[string] { return FromT(s + "!") }))
	assertSlice(t, ks5, []string{"ab!"})
	assertSlice(t, vs5, []string{"ab"})
	ks6, vs6 := biCollect(MapFlatBiAny(src, func(s string) Seq[any] { return FromT[any](1) }))
	if ks6[0] != any(1) || vs6[0] != any("ab") {
		t.Fatalf("MapFlatBiAny错误:%v %v", ks6, vs6)
	}
	ks7, vs7 := biCollect(MapFlatBiAnyRight(src, func(s string) Seq[any] { return FromT[any](2) }))
	if ks7[0] != any("ab") || vs7[0] != any(2) {
		t.Fatalf("MapFlatBiAnyRight错误:%v %v", ks7, vs7)
	}
}

// 场景: BiMapK/BiMapV 与 BiCastAnyK/V 直转
func Test_OpPre_BiMapCast(t *testing.T) {
	t.Parallel()
	mk, mv := BiMapK(BiFromT(1, "a"), func(k int, v string) string { return v + string(rune('0'+k)) }).First()
	if *mk != "a1" || *mv != "a" {
		t.Fatalf("BiMapK错误:%v %v", *mk, *mv)
	}
	uk, uv := BiMapV(BiFromT(1, "a"), func(k int, v string) int { return k * 10 }).First()
	if *uk != 1 || *uv != 10 {
		t.Fatalf("BiMapV错误:%v %v", *uk, *uv)
	}
	ck, cv := BiCastAnyK[int](BiMapK(BiFromT(1, "a"), func(k int, _ string) any { return any(k) })).First()
	if *ck != 1 || *cv != "a" {
		t.Fatalf("BiCastAnyK错误:%v %v", *ck, *cv)
	}
	ek, ev := BiCastAnyV[string](BiMapV(BiFromT(1, "a"), func(_ int, v string) any { return any(v) })).First()
	if *ek != 1 || *ev != "a" {
		t.Fatalf("BiCastAnyV错误:%v %v", *ek, *ev)
	}
}
