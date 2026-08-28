package storage

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

// Given 带工厂的 GLS key
// When 同一 goroutine 多次 Get
// Then 工厂仅调用一次, 后续返回缓存值
func Test_GlsItemWithFunc_FactoryCalledOnce(t *testing.T) {
	calls := 0
	item := NewGlsItemWithFunc[int](func() int {
		calls++
		return calls
	})
	t.Cleanup(func() {
		item.Delete()
		GlsClean()
	})

	v1, ok1 := item.Get()
	v2, ok2 := item.Get()
	v3, ok3 := item.Get()
	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("Get should always succeed: %v %v %v", ok1, ok2, ok3)
	}
	if v1 != 1 || v2 != 1 || v3 != 1 {
		t.Fatalf("factory should run once, cached: %d %d %d", v1, v2, v3)
	}

	// Set 覆盖缓存值
	item.Set(99)
	if v, ok := item.Get(); !ok || v != 99 {
		t.Fatalf("after Set: %v ok=%v, want 99", v, ok)
	}
	// 工厂不再被调用
	if calls != 1 {
		t.Fatalf("factory calls=%d want=1", calls)
	}
}

// Given 未 Set 的 key
// When GetSimple
// Then 返回零值且不 panic
func Test_GlsKeySimple_GetSimpleZero(t *testing.T) {
	item := NewGlsItem[string]()
	t.Cleanup(func() { GlsClean() })

	if got := item.(KeySimple[string]).GetSimple(); got != "" {
		t.Fatalf("GetSimple without Set = %q, want empty", got)
	}
}

// Given 多个 key 共存于同一 goroutine 子表
// When 分别 Set 后清理单个 key 且 autoClean 开启
// Then 其它 key 保留, 清空后子表释放
func Test_Gls_MultiKey_AutoClean(t *testing.T) {
	k1 := NewGlsItem[int]()
	k2 := NewGlsItem[string]()
	t.Cleanup(func() { GlsClean() })

	k1.Set(1)
	k2.Set("two")

	// 删除 k1 但子表仍有 k2, 不应整体清理
	k1.Delete(true)
	if _, ok := k2.Get(); !ok {
		t.Fatal("k2 should survive after k1.Delete")
	}

	// k2 也删除后子表清空
	k2.Delete(true)
	if _, ok := k1.Get(); ok {
		t.Fatal("k1 should be deleted")
	}
}

// Given GLS key 的类型参数
// When 同一底层 key 以不同类型读取
// Then 各类型 key 独立存储互不干扰
func Test_Gls_TypeSafety(t *testing.T) {
	ki := NewGlsItem[int]()
	ks := NewGlsItem[string]()
	t.Cleanup(func() {
		ki.Delete()
		ks.Delete()
		GlsClean()
	})

	ki.Set(7)
	ks.Set("seven")

	if v, ok := ki.Get(); !ok || v != 7 {
		t.Fatalf("int key = %v ok=%v", v, ok)
	}
	if v, ok := ks.Get(); !ok || v != "seven" {
		t.Fatalf("string key = %v ok=%v", v, ok)
	}
}

// Given 多 goroutine 独立使用同一 key
// When 并发 Set/Get/Clean
// Then 各 goroutine 值隔离, 结束后无残留
func Test_Gls_ConcurrentIsolation(t *testing.T) {
	item := NewGlsItem[int]()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		g := g
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GlsClean()
			for i := 0; i < 50; i++ {
				item.Set(g*100 + i)
				if v, ok := item.Get(); !ok || v != g*100+i {
					t.Errorf("isolated value = %v ok=%v, want %d", v, ok, g*100+i)
					return
				}
			}
		}()
	}
	wg.Wait()
	if n := atomic.LoadInt64(&glsEntryCount); n != 0 {
		t.Fatalf("gls entries leaked: %d", n)
	}
}

// mapMakerMakers 各 Map 实现构造器, 用于差分对拍
var mapMakers = []struct {
	name string
	make func() Map[int, int]
}{
	{"Go", func() Map[int, int] { return MapTypeGo[int, int](0).createMap() }},
	{"Swiss", func() Map[int, int] { return MapTypeSwiss[int, int](4).createMap() }},
	{"SwissConcurrent", func() Map[int, int] { return MapTypeSwissConcurrent[int, int]().createMap() }},
	{"Array", func() Map[int, int] { return MapTypeArray[int, int](2).createMap() }},
	{"ConcurrentWrapper", func() Map[int, int] { return MapTypeConcurrentWrapper(MapTypeArray[int, int](2)).createMap() }},
}

// Given 任一 Map 实现与原生 map 同步执行随机操作序列
// When 每步操作后对拍
// Then Get/Has/Count 与原生 map 完全一致
func Test_MapDifferential_RandomOps(t *testing.T) {
	t.Parallel()
	for _, mm := range mapMakers {
		mm := mm
		t.Run(mm.name, func(t *testing.T) {
			t.Parallel()
			m := mm.make()
			ref := make(map[int]int)
			rnd := rand.New(rand.NewSource(1))

			for step := 0; step < 2000; step++ {
				key := rnd.Intn(64)
				switch rnd.Intn(5) {
				case 0, 1: // 写
					val := rnd.Int()
					m.Put(key, val)
					ref[key] = val
				case 2: // 删
					m.Delete(key)
					delete(ref, key)
				case 3: // 读对拍
					gotVal, gotOk := m.Get(key)
					wantVal, wantOk := ref[key]
					if gotOk != wantOk || (wantOk && gotVal != wantVal) {
						t.Fatalf("step %d Get(%d) = %v,%v want %v,%v", step, key, gotVal, gotOk, wantVal, wantOk)
					}
				case 4: // 存在性与计数对拍
					if m.Has(key) != func() bool { _, ok := ref[key]; return ok }() {
						t.Fatalf("step %d Has(%d) mismatch", step, key)
					}
				}
				if m.Count() != len(ref) {
					t.Fatalf("step %d Count=%d want=%d", step, m.Count(), len(ref))
				}
			}

			// 迭代存在性: 实现迭代到的键集合与原生 map 一致
			seen := make(map[int]bool)
			m.Iter(func(k, v int) (stop bool) {
				wv, wok := ref[k]
				if !wok || wv != v {
					t.Fatalf("Iter got (%d,%d), want value %v exist %v", k, v, wv, wok)
				}
				seen[k] = true
				return false
			})
			if len(seen) != len(ref) {
				t.Fatalf("Iter visited %d keys, want %d", len(seen), len(ref))
			}
			for k := range ref {
				if !seen[k] {
					t.Fatalf("Iter missed key %d", k)
				}
			}
		})
	}
}
