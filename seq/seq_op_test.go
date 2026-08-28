package seq

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// awaitTimeout 统一超时保护基准
const awaitTimeout = 10 * time.Second

// 本文件只使用双轨通用API,go1.18~go1.27 均编译执行
// 可能死循环的断言段统一经 awaitResult 包裹超时保护

// assertSlice 断言切片逐元素相等
func assertSlice[T comparable](t *testing.T, got, expect []T) {
	t.Helper()
	if len(got) != len(expect) {
		t.Fatalf("长度 %d != %d: %v", len(got), len(expect), got)
	}
	for i := range got {
		if got[i] != expect[i] {
			t.Fatalf("索引%d: %v != %v", i, got[i], expect[i])
		}
	}
}

// recordPeak CAS记录并发峰值
func recordPeak(maxC *int32, c int32) {
	for {
		m := atomic.LoadInt32(maxC)
		if c <= m || atomic.CompareAndSwapInt32(maxC, m, c) {
			return
		}
	}
}

// waitPeak 启动监视goroutine,并发数达到target时关闭gate;超时未达标自动退出不关gate,
// 以事件同步替代时序断言证明并行峰值,配合awaitResult在死锁路径上超时失败
func waitPeak(target int32, current *int32, timeout time.Duration) chan struct{} {
	gate := make(chan struct{})
	go func() {
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			if atomic.LoadInt32(current) >= target {
				close(gate)
				return
			}
			runtime.Gosched()
		}
	}()
	return gate
}

// 场景: First 惰性取首个并中止生产
func Test_Op_First(t *testing.T) {
	t.Parallel()
	produced := 0
	f := From(func(f func(int)) {
		for {
			produced++
			f(produced)
		}
	}).First()
	if f == nil || *f != 1 {
		t.Fatalf("First应返回1,实际:%v", f)
	}
	//推送模型下源会多预取1个元素后才收到Stop中止
	if produced > 2 {
		t.Fatalf("First应至多生产2个元素,实际%d", produced)
	}
	if e := FromT[int]().First(); e != nil {
		t.Fatalf("空流First应为nil,实际:%v", *e)
	}
}

// 场景: FirstOrF 命中不调用默认,空流调用
func Test_Op_FirstOrF(t *testing.T) {
	t.Parallel()
	defaultCalled := false
	df := func() int { defaultCalled = true; return -1 }
	if v := FromT(3, 1).FirstOrF(df); v != 3 || defaultCalled {
		t.Fatalf("命中应返回3且不调用默认,实际:%v called:%v", v, defaultCalled)
	}
	if v := FromT[int]().FirstOrF(df); v != -1 || !defaultCalled {
		t.Fatalf("空流应返回默认值,实际:%v called:%v", v, defaultCalled)
	}
}

// 场景: Last/LastOrF 取末元素与空流分支
func Test_Op_Last_LastOrF(t *testing.T) {
	t.Parallel()
	if l := FromT(1, 2, 5).Last(); l == nil || *l != 5 {
		t.Fatalf("Last应返回5,实际:%v", l)
	}
	if l := FromT[int]().Last(); l != nil {
		t.Fatal("空流Last应为nil")
	}
	if v := FromT(1, 2, 5).LastOrF(func() int { return -1 }); v != 5 {
		t.Fatalf("LastOrF应返回5,实际:%d", v)
	}
	if v := FromT[int]().LastOrF(func() int { return -1 }); v != -1 {
		t.Fatalf("空流LastOrF应返回默认,实际:%d", v)
	}
}

// 场景: FindFirstBy 短路返回首个命中
func Test_Op_FindFirstBy(t *testing.T) {
	t.Parallel()
	scanned := 0
	f := FromT(1, 2, 3, 4).FindFirstBy(func(i int) bool {
		scanned++
		return i > 2
	})
	if f == nil || *f != 3 {
		t.Fatalf("应命中3,实际:%v", f)
	}
	//推送模型下命中后源会多评估1个元素才中止
	if scanned > 4 {
		t.Fatalf("应短路至多扫描4个,实际%d", scanned)
	}
	if f := FromT(1, 2).FindFirstBy(func(i int) bool { return false }); f != nil {
		t.Fatal("未命中应返回nil")
	}
}

// 场景: AnyMatch/AllMatch/NonMatch 空流/全命中/全不命中三态
func Test_Op_Match_ThreeStates(t *testing.T) {
	t.Parallel()
	//空流: AnyMatch=false, AllMatch=true, NonMatch=true
	if FromT[int]().AnyMatch(func(int) bool { return true }) {
		t.Fatal("空流AnyMatch应为false")
	}
	if !FromT[int]().AllMatch(func(int) bool { return false }) {
		t.Fatal("空流AllMatch应为true")
	}
	if !FromT[int]().NonMatch(func(int) bool { return true }) {
		t.Fatal("空流NonMatch应为true")
	}
	//AllMatch 全命中,首个不命中短路
	if !FromT(2, 4).AllMatch(func(i int) bool { return i%2 == 0 }) {
		t.Fatal("全命中AllMatch应为true")
	}
	scanned := 0
	if FromT(2, 3, 4).AllMatch(func(i int) bool {
		scanned++
		return i%2 == 0
	}) {
		t.Fatal("存在不命中AllMatch应为false")
	}
	if scanned != 2 {
		t.Fatalf("AllMatch应短路扫描2个,实际%d", scanned)
	}
}

// 场景: Count/SumBy/SumByFloat64 聚合
func Test_Op_Count_Sum(t *testing.T) {
	t.Parallel()
	if c := FromT(1, 2, 3, 4).Filter(func(i int) bool { return i%2 == 0 }).Count(); c != 2 {
		t.Fatalf("Count应2,实际%d", c)
	}
	if s := FromT(1, 2, 3).SumBy(func(i int) int { return i * 2 }); s != 12 {
		t.Fatalf("SumBy应12,实际%d", s)
	}
	if s := FromT(1, 2, 3).SumByFloat64(func(i int) float64 { return float64(i) }); s != 6 {
		t.Fatalf("SumByFloat64应6,实际%v", s)
	}
}

// 场景: JoinStringBy 自定义转换与分隔符
func Test_Op_JoinStringBy(t *testing.T) {
	t.Parallel()
	s := FromT(1, 2, 3).JoinStringBy(func(i int) string { return string(rune('a' + i - 1)) }, "-")
	if s != "a-b-c" {
		t.Fatalf("JoinStringBy结果错误:%s", s)
	}
	if s := FromT[int]().JoinStringBy(func(int) string { return "x" }, "-"); s != "" {
		t.Fatalf("空流应为空串:%s", s)
	}
}

// 场景: Reduce 任意键形态聚合,双轨通用
func Test_Op_Reduce(t *testing.T) {
	t.Parallel()
	r := FromT("a", "bb").Reduce(func(t string, acc any) any {
		return acc.(int) + len(t)
	}, any(0))
	if r.(int) != 3 {
		t.Fatalf("Reduce聚合错误:%v", r)
	}
}

// 场景: Join 合并保序
func Test_Op_Join(t *testing.T) {
	t.Parallel()
	r := FromT(1, 2).Join(FromT(3), FromT(4, 5)).ToSlice()
	assertSlice(t, r, []int{1, 2, 3, 4, 5})
	assertSlice(t, FromT(1).Join().ToSlice(), []int{1})
}

// 场景: Add/AddIf/AddIfF 追加元素语义
func Test_Op_AddFamily(t *testing.T) {
	t.Parallel()
	assertSlice(t, FromT(1).Add(2, 3).ToSlice(), []int{1, 2, 3})
	assertSlice(t, FromT(1).AddIf(true, 2).ToSlice(), []int{1, 2})
	assertSlice(t, FromT(1).AddIf(false, 2).ToSlice(), []int{1})
	//AddIfF 按元素逐个判断是否追加
	r := FromT(1).AddIfF(func(e int) bool { return e > 0 }, 2, -3, 4).ToSlice()
	assertSlice(t, r, []int{1, 2, 4})
}

// 场景: Limit/Skip/Take/Drop 边界 n=0 与负数
func Test_Op_Limit_Skip_Edge(t *testing.T) {
	t.Parallel()
	src := func() Seq[int] { return FromT(1, 2, 3) }
	assertSlice(t, src().Limit(2).ToSlice(), []int{1, 2})
	assertSlice(t, src().Limit(0).ToSlice(), []int{})
	assertSlice(t, src().Take(0).ToSlice(), []int{})
	assertSlice(t, src().Take(-1).ToSlice(), []int{})
	//负数跳过视为不跳过
	assertSlice(t, src().Skip(-1).ToSlice(), []int{1, 2, 3})
	assertSlice(t, src().Drop(-1).ToSlice(), []int{1, 2, 3})
	assertSlice(t, src().Skip(5).ToSlice(), []int{})
}

// 场景: DropWhile 跳到首次true(含触发元素),永不true为空
func Test_Op_DropWhile(t *testing.T) {
	t.Parallel()
	assertSlice(t, FromT(1, 2, 3, 1).DropWhile(func(i int) bool { return i >= 3 }).ToSlice(), []int{3, 1})
	assertSlice(t, FromT(5, 1, 2).DropWhile(func(i int) bool { return i >= 3 }).ToSlice(), []int{5, 1, 2})
	assertSlice(t, FromT(1, 2).DropWhile(func(int) bool { return false }).ToSlice(), []int{})
}

// 场景: Distinct/DistinctCustomize 去重保序
func Test_Op_Distinct(t *testing.T) {
	t.Parallel()
	assertSlice(t, FromT(1, 2, 1, 3, 2).Distinct().ToSlice(), []int{1, 2, 3})
	seen := map[int]struct{}{}
	r := FromT(1, 2, 1, 3).DistinctCustomize(func(e int) bool {
		if _, ok := seen[e]; ok {
			return true
		}
		seen[e] = struct{}{}
		return false
	}).ToSlice()
	assertSlice(t, r, []int{1, 2, 3})
}

// 场景: Reverse 逆序含空流,Seq与BiSeq
func Test_Op_Reverse(t *testing.T) {
	t.Parallel()
	assertSlice(t, FromT(1, 2, 3).Reverse().ToSlice(), []int{3, 2, 1})
	assertSlice(t, FromT[int]().Reverse().ToSlice(), []int{})
	var got []int
	BiFromSeq(FromT(1, 2, 3), func(v int) (int, int) { return v, v }).Reverse()(func(k, v int) {
		got = append(got, k)
	})
	assertSlice(t, got, []int{3, 2, 1})
}

// 场景: SortCustomize 自定义排序函数
func Test_Op_SortCustomize(t *testing.T) {
	t.Parallel()
	r := FromT("ccc", "a", "bb").SortCustomize(func(s []string) {
		sort.Slice(s, func(i, j int) bool { return len(s[i]) < len(s[j]) })
	}).ToSlice()
	assertSlice(t, r, []string{"a", "bb", "ccc"})
	assertSlice(t, FromT[int]().SortCustomize(func([]int) {}).ToSlice(), []int{})
}

// 场景: Sort 空流与多轮消费
func Test_Op_Sort(t *testing.T) {
	t.Parallel()
	s := FromT(3, 1, 2).Sort(LessT[int])
	assertSlice(t, s.ToSlice(), []int{1, 2, 3})
	//多次消费结果一致
	assertSlice(t, s.ToSlice(), []int{1, 2, 3})
}

// 场景: Cache(init=true) 立即触发源消费且可重复回放
func Test_Op_Cache_Init(t *testing.T) {
	t.Parallel()
	produced := 0
	src := From(func(f func(int)) {
		for i := 0; i < 3; i++ {
			produced++
			f(i)
		}
	})
	s := src.Cache(true)
	//init=true 立即填满缓存
	if produced != 3 {
		t.Fatalf("Cache(true)应立即消费3个,实际%d", produced)
	}
	assertSlice(t, s.ToSlice(), []int{0, 1, 2})
	assertSlice(t, s.ToSlice(), []int{0, 1, 2})
	if produced != 3 {
		t.Fatalf("回放不应再触发生产,实际%d", produced)
	}
}

// 场景: Repeat 显式n重复,无限重复可被Take终止,标准构造源与裸源均需可终止
func Test_Op_Repeat(t *testing.T) {
	t.Parallel()
	assertSlice(t, FromT(1, 2).Repeat(2).ToSlice(), []int{1, 2, 1, 2})
	assertSlice(t, FromT(1, 2).Repeat(0).ToSlice(), []int{})
	//标准构造源吸收Stop哨兵后正常返回,Repeat须识别哨兵停止重启,否则死循环
	msg := awaitResult(awaitTimeout, func() {
		assertSlice(t, FromT(1, 2).Repeat().Take(5).ToSlice(), []int{1, 2, 1, 2, 1})
	})
	if msg != "ok" {
		t.Fatalf("无限Repeat消费异常:%s", msg)
	}
	//裸源无stopRecover,哨兵由Repeat自身吸收
	msg = awaitResult(awaitTimeout, func() {
		raw := Seq[int](func(f func(int)) {
			for {
				f(1)
			}
		})
		assertSlice(t, raw.Repeat().Take(3).ToSlice(), []int{1, 1, 1})
	})
	if msg != "ok" {
		t.Fatalf("裸源无限Repeat消费异常:%s", msg)
	}
}

// 场景: Sync 保证并行消费串行且不丢失
func Test_Op_Sync(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	total := 0
	count := 0
	FromT(1, 2, 3, 4, 5, 6).Parallel(4).Sync().ForEach(func(i int) {
		mu.Lock()
		defer mu.Unlock()
		total += i
		count++
	})
	if total != 21 || count != 6 {
		t.Fatalf("Sync后应恰好消费全部,total=%d count=%d", total, count)
	}
}

// 场景: MapParallel 无顺序保证时全元素完成,order=0消费竞争按语义以Sync保证
func Test_Op_MapParallel_NoOrder(t *testing.T) {
	t.Parallel()
	r := FromT(1, 2, 3, 4, 5, 6, 7, 8).MapParallel(func(v int) any { return v * 2 }).Sync().ToSlice()
	if len(r) != 8 {
		t.Fatalf("元素数应8,实际%d", len(r))
	}
	sum := 0
	for _, v := range r {
		sum += v.(int)
	}
	if sum != 72 {
		t.Fatalf("求和应72,实际%d", sum)
	}
}

// 场景: OnEachF 条件回调与skip变体
func Test_Op_OnEachF(t *testing.T) {
	t.Parallel()
	var hits []int
	FromT(1, 2, 3, 4, 5, 6).OnEachF(func(i int) bool { return i%2 == 0 }, func(i int) {
		hits = append(hits, i)
	}).ForEach(func(int) {})
	assertSlice(t, hits, []int{2, 4, 6})
	//skip=2 跳过前2个的额外执行机会
	hits = nil
	FromT(1, 2, 3, 4, 5, 6).OnEachF(func(i int) bool { return i%2 == 0 }, func(i int) {
		hits = append(hits, i)
	}, 2).ForEach(func(int) {})
	//skip=2时x从-2起,元素4时x=2才首次触发
	assertSlice(t, hits, []int{4, 6})
}

// 场景: OnEachN 每n个触发与skip变体,step非法panic
func Test_Op_OnEachN(t *testing.T) {
	t.Parallel()
	var hits []int
	FromT(1, 2, 3, 4, 5, 6, 7, 8, 9).OnEachN(3, func(i int) {
		hits = append(hits, i)
	}).ForEach(func(int) {})
	assertSlice(t, hits, []int{3, 6, 9})
	hits = nil
	FromT(1, 2, 3, 4, 5, 6, 7, 8, 9).OnEachN(3, func(i int) {
		hits = append(hits, i)
	}, 2).ForEach(func(int) {})
	assertSlice(t, hits, []int{5, 8})
	func() {
		defer func() {
			if recover() == nil {
				t.Error("step<=0应panic")
			}
		}()
		FromT(1).OnEachN(0, func(int) {})
	}()
}

// 场景: OnEachNX 整数倍触发+尾部补触发,skip变体,空流不触发
func Test_Op_OnEachNX(t *testing.T) {
	t.Parallel()
	var hits []int
	FromT(1, 2, 3, 4, 5, 6, 7).OnEachNX(3, func(i int) {
		hits = append(hits, i)
	}).ForEach(func(int) {})
	//3,6整倍触发,末尾余1个补触发7
	assertSlice(t, hits, []int{3, 6, 7})
	hits = nil
	FromT(1, 2, 3).OnEachNX(3, func(i int) {
		hits = append(hits, i)
	}).ForEach(func(int) {})
	//恰好整除不补触发
	assertSlice(t, hits, []int{3})
	hits = nil
	FromT(1, 2).OnEachNX(3, func(i int) {
		hits = append(hits, i)
	}, 1).ForEach(func(int) {})
	//skip=1时2个元素等价于未skip的3个计数,2%3!=0补触发
	assertSlice(t, hits, []int{2})
	hits = nil
	FromT[int]().OnEachNX(3, func(int) {
		t.Error("空流不应触发OnEachNX回调")
	}).ForEach(func(int) {})
}

// 场景: OnBefore/OnAfter 位置边界
func Test_Op_OnBefore_After(t *testing.T) {
	t.Parallel()
	assertOn := func(got, expect []int, name string) {
		t.Helper()
		assertSlice(t, got, expect)
	}
	var hits []int
	FromT(1, 2, 3, 4, 5).OnBefore(2, func(i int) { hits = append(hits, i) }).ForEach(func(int) {})
	assertOn(hits, []int{1, 2}, "OnBefore(2)")
	hits = nil
	FromT(1, 2, 3, 4, 5).OnBefore(0, func(i int) { hits = append(hits, i) }).ForEach(func(int) {})
	assertOn(hits, []int{}, "OnBefore(0)")
	hits = nil
	FromT(1, 2, 3).OnBefore(-1, func(int) { t.Error("负位置不应触发") }).ForEach(func(int) {})
	hits = nil
	FromT(1, 2, 3, 4, 5).OnAfter(2, func(i int) { hits = append(hits, i) }).ForEach(func(int) {})
	assertOn(hits, []int{3, 4, 5}, "OnAfter(2)")
	hits = nil
	FromT(1, 2, 3).OnAfter(0, func(i int) { hits = append(hits, i) }).ForEach(func(int) {})
	assertOn(hits, []int{1, 2, 3}, "OnAfter(0)")
}

// 场景: OnFirst 单次触发,OnLast 末元素与空流nil
func Test_Op_OnFirst_OnLast_Empty(t *testing.T) {
	t.Parallel()
	firstHits := 0
	FromT(1, 2, 3).OnFirst(func(int) { firstHits++ }).ForEach(func(int) {})
	if firstHits != 1 {
		t.Fatalf("OnFirst应触发1次,实际%d", firstHits)
	}
	lastVal := 0
	lastNil := false
	FromT(1, 2, 9).OnLast(func(i *int) {
		if i == nil {
			lastNil = true
			return
		}
		lastVal = *i
	}).ForEach(func(int) {})
	if lastNil || lastVal != 9 {
		t.Fatalf("OnLast应收到9,nil=%v", lastNil)
	}
	FromT[int]().OnLast(func(i *int) {
		if i != nil {
			t.Fatal("空流OnLast参数应为nil")
		}
	}).ForEach(func(int) {})
}

// 场景: RecoverErr 捕获panic值,Stop哨兵穿透不误捕
func Test_Op_RecoverErr(t *testing.T) {
	t.Parallel()
	var captured any
	r := From(func(f func(int)) {
		f(1)
		panic("boom")
	}).RecoverErr(func(a any) { captured = a }).ToSlice()
	assertSlice(t, r, []int{1})
	if captured != "boom" {
		t.Fatalf("应捕获boom,实际:%v", captured)
	}
	//消费端Take的Stop哨兵不应触发RecoverErr
	hit := false
	r = FromIntSeq().RecoverErr(func(any) { hit = true }).Take(3).ToSlice()
	assertSlice(t, r, []int{0, 1, 2})
	if hit {
		t.Fatal("Stop哨兵不应进入RecoverErr")
	}
}

// 场景: RecoverErrWithValue 捕获最后一次元素值
func Test_Op_RecoverErrWithValue(t *testing.T) {
	t.Parallel()
	var lastV int
	var captured any
	From(func(f func(int)) {
		f(1)
		f(2)
		panic("boom")
	}).RecoverErrWithValue(func(v int, a any) {
		lastV = v
		captured = a
	}).ForEach(func(int) {})
	if lastV != 2 || captured != "boom" {
		t.Fatalf("应捕获last=2 boom,实际:%d %v", lastV, captured)
	}
}

// 场景: Finally 正常路径与Stop路径各执行一次
func Test_Op_Finally(t *testing.T) {
	t.Parallel()
	called := 0
	FromT(1, 2).Finally(func() { called++ }).ForEach(func(int) {})
	if called != 1 {
		t.Fatalf("Finally应执行1次,实际%d", called)
	}
	called = 0
	FromIntSeq().Finally(func() { called++ }).Take(2).ToSlice()
	if called != 1 {
		t.Fatalf("Stop路径Finally应执行1次,实际%d", called)
	}
}

// 场景: GroupBy 家族移至 seq_op_pre127_test.go(any键的comparable约束需go1.20,方法形态泛型键在go1.27轨道已有覆盖)

// 场景: MapSliceBy 自定义分块与余量刷出
func Test_Op_MapSliceBy(t *testing.T) {
	t.Parallel()
	var sizes []int
	MapSliceBy(FromT(1, 2, 3, 4, 5, 6, 7), func(_ int, ts []int) bool {
		return len(ts) == 3
	})(func(v any) {
		sizes = append(sizes, len(v.([]int)))
	})
	assertSlice(t, sizes, []int{3, 3, 1})
}

// 场景: FromChan/FromIterator/FromSeq/FromSeq2 生成器
func Test_Op_Generator_Basic(t *testing.T) {
	t.Parallel()
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	assertSlice(t, FromChan(ch).ToSlice(), []int{1, 2, 3})
	assertSlice(t, FromIterator(IteratorInt(0, 3)).ToSlice(), []int{0, 1, 2, 3})
	assertSlice(t, FromSeq(func(y func(int) bool) {
		for i := 0; i < 3; i++ {
			if !y(i) {
				return
			}
		}
	}).ToSlice(), []int{0, 1, 2})
	var ks, vs []int
	FromSeq2(func(y func(int, int) bool) {
		if !y(1, 10) {
			return
		}
		y(2, 20)
	})(func(k, v int) {
		ks = append(ks, k)
		vs = append(vs, v)
	})
	assertSlice(t, ks, []int{1, 2})
	assertSlice(t, vs, []int{10, 20})
}

// 场景: FromSliceRepeat/FromTRepeat/FromTRepeatN 重复生成
func Test_Op_Generator_Repeat(t *testing.T) {
	t.Parallel()
	assertSlice(t, FromSliceRepeat([]int{1, 2}, 3).Take(5).ToSlice(), []int{1, 2, 1, 2, 1})
	assertSlice(t, FromSliceRepeat([]int{1}, 0).Take(2).ToSlice(), []int{1, 1})
	assertSlice(t, FromTRepeatN(2, 7).Take(5).ToSlice(), []int{7, 7})
	assertSlice(t, FromTRepeat(7).Take(3).ToSlice(), []int{7, 7, 7})
}

// 场景: FromIntSeq 起止/步长/降序/无限配合Take,IteratorInt 对应形态
func Test_Op_Generator_Range(t *testing.T) {
	t.Parallel()
	assertSlice(t, FromIntSeq(1, 10).Take(10).ToSlice(), []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	assertSlice(t, FromIntSeq(5, 1).Take(5).ToSlice(), []int{5, 4, 3, 2, 1})
	assertSlice(t, FromIntSeq(3).Take(2).ToSlice(), []int{3, 4})
	assertSlice(t, FromIterator(IteratorInt(5, 1)).ToSlice(), []int{5, 4, 3, 2, 1})
	//三参步长: 首元素为起始值,按步长递进,升序含结束值
	assertSlice(t, FromIntSeq(0, 10, 5).Take(3).ToSlice(), []int{0, 5, 10})
	//步长跨过结束值时不含结束值
	assertSlice(t, FromIntSeq(0, 10, 4).Take(3).ToSlice(), []int{0, 4, 8})
	//降序: 从起始值按负步长递进,含结束值
	assertSlice(t, FromIntSeq(0, 5, -1).Take(6).ToSlice(), []int{5, 4, 3, 2, 1, 0})
	assertSlice(t, FromIterator(IteratorInt(0, 10, 5)).ToSlice(), []int{0, 5, 10})
}

// 场景: FromRandIntSeq 数量与范围
func Test_Op_Generator_Rand(t *testing.T) {
	t.Parallel()
	r := FromRandIntSeq(5, 100).ToSlice()
	if len(r) != 6 {
		t.Fatalf("参数1=5应生成6个元素,实际%d", len(r))
	}
	for _, v := range r {
		if v < 0 || v >= 100 {
			t.Fatalf("元素越界:%d", v)
		}
	}
}

// 场景: FromTreeTV 树带值遍历
func Test_Op_Generator_Tree(t *testing.T) {
	t.Parallel()
	tree := struct {
		v     int
		child []int
	}{
		v:     1,
		child: []int{2, 3},
	}
	childOf := func(n int) Seq[int] {
		if n == 1 {
			return FromSlice(tree.child)
		}
		return FromSlice[int](nil)
	}
	valueOf := func(n int) int { return n * 10 }
	r := FromTreeTV(tree.v, childOf, valueOf).ToSlice()
	assertSlice(t, r, []int{10, 20, 30})
}

// 场景: FromBi/FromBiK/FromBiV 提取单元素流
func Test_Op_FromBi(t *testing.T) {
	t.Parallel()
	bi := BiFromSeq(FromT(1, 2, 3), func(v int) (int, string) { return v, string(rune('a' + v - 1)) })
	assertSlice(t, FromBiK(bi).ToSlice(), []int{1, 2, 3})
	assertSlice(t, FromBiV(bi).ToSlice(), []string{"a", "b", "c"})
	assertSlice(t, FromBi(bi, func(k int, _ string) int { return k * 2 }).ToSlice(), []int{2, 4, 6})
}

// 场景: BiFirst/FirstOrF/Last/LastOrF 命中与空流
func Test_Op_Bi_FirstLast(t *testing.T) {
	t.Parallel()
	bi := func() BiSeq[int, string] {
		return BiFromSeq(FromT(1, 2, 3), func(v int) (int, string) { return v, string(rune('a' + v - 1)) })
	}
	k, v := bi().First()
	if k == nil || v == nil || *k != 1 || *v != "a" {
		t.Fatalf("BiFirst错误:%v %v", k, v)
	}
	ek, ev := BiFromSeq(FromT[int](), func(v int) (int, int) { return v, v }).First()
	if ek != nil || ev != nil {
		t.Fatal("空流BiFirst应返回nil")
	}
	called := false
	fk, fv := bi().FirstOrF(func() (int, string) { called = true; return -1, "z" })
	if called || fk != 1 || fv != "a" {
		t.Fatalf("BiFirstOrF命中错误:%v %v %v", fk, fv, called)
	}
	k2, v2 := bi().Last()
	if *k2 != 3 || *v2 != "c" {
		t.Fatalf("BiLast错误:%v %v", *k2, *v2)
	}
	dk, dv := bi().LastOrF(func() (int, string) { called = true; return -1, "z" })
	if called || dk != 3 || dv != "c" {
		t.Fatalf("BiLastOrF命中错误:%v %v %v", dk, dv, called)
	}
	lk, lv := BiFromSeq(FromT[int](), func(v int) (int, int) { return v, v }).Last()
	if lk != nil || lv != nil {
		t.Fatal("空流BiLast应返回nil")
	}
}

// 场景: BiSeq AnyMatch/AllMatch 含空流
func Test_Op_Bi_Match(t *testing.T) {
	t.Parallel()
	bi := BiFromSeq(FromT(1, 2, 3), func(v int) (int, int) { return v, v })
	if !bi.AnyMatch(func(k, v int) bool { return k == 2 }) {
		t.Fatal("AnyMatch应命中")
	}
	if bi.AnyMatch(func(k, v int) bool { return false }) {
		t.Fatal("AnyMatch不应命中")
	}
	if !bi.AllMatch(func(k, v int) bool { return k < 10 }) {
		t.Fatal("AllMatch应全命中")
	}
	if bi.AllMatch(func(k, v int) bool { return k < 2 }) {
		t.Fatal("AllMatch不应全命中")
	}
	empty := BiFromSeq(FromT[int](), func(v int) (int, int) { return v, v })
	if empty.AnyMatch(func(k, v int) bool { return true }) {
		t.Fatal("空流AnyMatch应为false")
	}
	if !empty.AllMatch(func(k, v int) bool { return false }) {
		t.Fatal("空流AllMatch应为true")
	}
}

// 场景: BiSeq Count/SumBy/SumByFloat64/JoinStringBy/Reduce
func Test_Op_Bi_Aggregate(t *testing.T) {
	t.Parallel()
	bi := BiFromSeq(FromT(1, 2, 3), func(v int) (int, int) { return v, v * 10 })
	if c := bi.Count(); c != 3 {
		t.Fatalf("Count应3,实际%d", c)
	}
	if s := bi.SumBy(func(k, v int) int { return k + v }); s != 6+60 {
		t.Fatalf("SumBy错误:%d", s)
	}
	if s := bi.SumByFloat64(func(k, v int) float64 { return float64(v) }); s != 60 {
		t.Fatalf("SumByFloat64错误:%v", s)
	}
	if s := bi.JoinStringBy(func(k, v int) string { return string(rune('a' + k - 1)) }, ","); s != "a,b,c" {
		t.Fatalf("JoinStringBy错误:%s", s)
	}
	r := bi.Reduce(func(k, v int, acc any) any { return acc.(int) + v }, any(0))
	if r.(int) != 60 {
		t.Fatalf("Reduce错误:%v", r)
	}
}

// 场景: BiJoin/BiAdd/BiAddIf/BiAddIfF
func Test_Op_Bi_AddJoin(t *testing.T) {
	t.Parallel()
	bi := func() BiSeq[int, int] { return BiFromSeq(FromT(1, 2), func(v int) (int, int) { return v, v }) }
	var got []int
	bi().Join(BiFromT(3, 3), BiFromT(4, 4))(func(k, _ int) { got = append(got, k) })
	assertSlice(t, got, []int{1, 2, 3, 4})
	got = nil
	bi().Add(5, 5)(func(k, _ int) { got = append(got, k) })
	assertSlice(t, got, []int{1, 2, 5})
	got = nil
	bi().AddIf(true, 5, 5)(func(k, _ int) { got = append(got, k) })
	assertSlice(t, got, []int{1, 2, 5})
	got = nil
	bi().AddIf(false, 5, 5)(func(k, _ int) { got = append(got, k) })
	assertSlice(t, got, []int{1, 2})
	got = nil
	//AddIfF 条件接收流本身,按Count判断
	bi().AddIfF(func(s BiSeq[int, int]) bool { return s.Count() == 2 }, 5, 5)(func(k, _ int) {
		got = append(got, k)
	})
	assertSlice(t, got, []int{1, 2, 5})
}

// 场景: BiDistinct/BiReverse/BiSort/BiRepeat
func Test_Op_Bi_Structure(t *testing.T) {
	t.Parallel()
	bi := func() BiSeq[int, int] { return BiFromSeq(FromT(1, 2, 1, 3), func(v int) (int, int) { return v, v }) }
	var ks []int
	bi().Distinct()(func(k, _ int) { ks = append(ks, k) })
	assertSlice(t, ks, []int{1, 2, 3})
	ks = nil
	bi().Repeat(2).Take(8)(func(k, _ int) { ks = append(ks, k) })
	assertSlice(t, ks, []int{1, 2, 1, 3, 1, 2, 1, 3})
	ks = nil
	//标准构造源无限BiRepeat()+Take须可终止,Repeat识别停止哨兵后停止重启
	msg := awaitResult(awaitTimeout, func() {
		bi().Repeat().Take(5)(func(k, _ int) { ks = append(ks, k) })
	})
	if msg != "ok" {
		t.Fatalf("无限BiRepeat异常:%s", msg)
	}
	assertSlice(t, ks, []int{1, 2, 1, 3, 1})
	//裸源无stopRecover,哨兵由BiRepeat自身吸收
	ks = nil
	msg = awaitResult(awaitTimeout, func() {
		raw := BiSeq[int, int](func(f func(int, int)) {
			for {
				f(1, 1)
			}
		})
		raw.Repeat().Take(3)(func(k, _ int) { ks = append(ks, k) })
	})
	if msg != "ok" {
		t.Fatalf("裸源无限BiRepeat异常:%s", msg)
	}
	assertSlice(t, ks, []int{1, 1, 1})
	ks = nil
	bi().Sort(func(k1, v1, k2, v2 int) bool { return k1 > k2 })(func(k, _ int) { ks = append(ks, k) })
	assertSlice(t, ks, []int{3, 2, 1, 1})
}

// 场景: BiCache 重复回放与init立即消费
func Test_Op_Bi_Cache(t *testing.T) {
	t.Parallel()
	produced := 0
	src := func() BiSeq[int, int] {
		return BiFromSeq(func(f func(int)) {
			for i := 0; i < 3; i++ {
				produced++
				f(i)
			}
		}, func(v int) (int, int) { return v, v })
	}
	s := src().Cache(true)
	if produced != 3 {
		t.Fatalf("init=true应立即消费3个,实际%d", produced)
	}
	var ks []int
	s(func(k, _ int) { ks = append(ks, k) })
	assertSlice(t, ks, []int{0, 1, 2})
	ks = nil
	s(func(k, _ int) { ks = append(ks, k) })
	assertSlice(t, ks, []int{0, 1, 2})
}

// 场景: BiSync 并行后串行消费不丢失
func Test_Op_Bi_Sync(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	sum := 0
	n := 0
	BiFromSeq(FromT(1, 2, 3, 4), func(v int) (int, int) { return v, v }).
		Parallel(4).Sync()(func(k, v int) {
		mu.Lock()
		defer mu.Unlock()
		sum += k
		n++
	})
	if sum != 10 || n != 4 {
		t.Fatalf("BiSync后应恰好消费全部,sum=%d n=%d", sum, n)
	}
}

// 场景: BiSeq OnBefore/OnAfter/OnFirst/OnLast 含空流
func Test_Op_Bi_Observers(t *testing.T) {
	t.Parallel()
	bi := BiFromSeq(FromT(1, 2, 3), func(v int) (int, int) { return v, v })
	var hits []int
	bi.OnBefore(2, func(k, _ int) { hits = append(hits, k) })(func(k, _ int) { _ = k })
	assertSlice(t, hits, []int{1, 2})
	hits = nil
	bi.OnAfter(1, func(k, _ int) { hits = append(hits, k) })(func(k, _ int) { _ = k })
	assertSlice(t, hits, []int{2, 3})
	hits = nil
	bi.OnFirst(func(k, _ int) { hits = append(hits, k) })(func(k, _ int) { _ = k })
	assertSlice(t, hits, []int{1})
	//OnLast 末元素
	lastK := -1
	bi.OnLast(func(k, v *int) {
		if k == nil || v == nil {
			t.Fatal("非空流OnLast参数不应为nil")
		}
		lastK = *k
	})(func(k, _ int) { _ = k })
	if lastK != 3 {
		t.Fatalf("BiOnLast应收到3,实际%d", lastK)
	}
	//空流OnLast收nil
	BiFromSeq(FromT[int](), func(v int) (int, int) { return v, v }).OnLast(func(k, v *int) {
		if k != nil || v != nil {
			t.Fatal("空流BiOnLast参数应为nil")
		}
	})(func(k, _ int) { _ = k })
}

// 场景: BiOnEachNX 整倍触发+尾部补触发+skip
func Test_Op_Bi_OnEachNX(t *testing.T) {
	t.Parallel()
	bi := BiFromSeq(FromT(1, 2, 3, 4, 5), func(v int) (int, int) { return v, v })
	var hits []int
	bi.OnEachNX(2, func(idx, k, _ int) { hits = append(hits, idx) })(func(k, _ int) { _ = k })
	//idx=2,4整倍触发,尾部余1补触发idx=5
	assertSlice(t, hits, []int{2, 4, 5})
	hits = nil
	bi.OnEachNX(2, func(idx, k, _ int) { hits = append(hits, idx) }, 1)(func(k, _ int) { _ = k })
	//skip=1时5个元素计数为4,4%2==0不补触发
	assertSlice(t, hits, []int{2, 4})
	//空流不传skip时安全
	BiFromSeq(FromT[int](), func(v int) (int, int) { return v, v }).OnEachNX(2, func(idx, k, _ int) {
		t.Error("空流不应触发")
	})(func(k, _ int) { _ = k })
	//空流+skip导致计数非整倍,同样不得触发收尾或panic
	BiFromSeq(FromT[int](), func(v int) (int, int) { return v, v }).OnEachNX(2, func(idx, k, _ int) {
		t.Error("空流+skip不应触发")
	}, 1)(func(k, _ int) { _ = k })
}

// 场景: BiRecoverErr/BiRecoverErrWithValue/BiFinally
func Test_Op_Bi_RecoverFinally(t *testing.T) {
	t.Parallel()
	var captured any
	src := func(f func(int, int)) {
		f(1, 1)
		panic("bi-boom")
	}
	BiFrom(src).RecoverErr(func(a any) { captured = a })(func(k, _ int) { _ = k })
	if captured != "bi-boom" {
		t.Fatalf("应捕获bi-boom,实际:%v", captured)
	}
	var lastK, lastV int
	BiFrom(func(f func(int, int)) {
		f(1, 1)
		f(2, 2)
		panic("bi-boom")
	}).RecoverErrWithValue(func(k, v int, a any) {
		lastK, lastV = k, v
		_ = a
	})(func(k, _ int) { _ = k })
	if lastK != 2 || lastV != 2 {
		t.Fatalf("应捕获last=(2,2),实际:(%d,%d)", lastK, lastV)
	}
	called := 0
	BiFromT(1, 1).Finally(func() { called++ })(func(k, _ int) { _ = k })
	if called != 1 {
		t.Fatalf("BiFinally应执行1次,实际%d", called)
	}
}

// 场景: BiSeq Take(0)/Drop 负数边界与Filter
func Test_Op_Bi_Edge(t *testing.T) {
	t.Parallel()
	bi := func() BiSeq[int, int] { return BiFromSeq(FromT(1, 2, 3), func(v int) (int, int) { return v, v }) }
	if c := bi().Take(0).Count(); c != 0 {
		t.Fatalf("Take(0)应为空,实际%d", c)
	}
	if c := bi().Drop(-1).Count(); c != 3 {
		t.Fatalf("Drop负数应保留全部,实际%d", c)
	}
	if c := bi().Filter(func(k, v int) bool { return k > 1 }).Count(); c != 2 {
		t.Fatalf("Filter后应2个,实际%d", c)
	}
}

// 场景: BiSeq Map 任意形态双值转换,双轨通用
func Test_Op_Bi_MapAny(t *testing.T) {
	t.Parallel()
	var ks, vs []int
	BiFromT(1, 2).Map(func(k, v int) (any, any) { return k * 10, v * 10 })(func(a, b any) {
		ks = append(ks, a.(int))
		vs = append(vs, b.(int))
	})
	assertSlice(t, ks, []int{10})
	assertSlice(t, vs, []int{20})
}

// 场景: BiSeq MapVParallel 无顺序保证时全元素完成
func Test_Op_Bi_MapVParallel_NoOrder(t *testing.T) {
	t.Parallel()
	var n int32
	BiFromSeq(FromT(1, 2, 3, 4), func(v int) (int, int) { return v, v }).
		MapVParallel(func(k, v int) any { return v }, 0, 4)(func(k int, v any) {
		atomic.AddInt32(&n, 1)
	})
	if n != 4 {
		t.Fatalf("MapVParallel应消费4个,实际%d", n)
	}
}
