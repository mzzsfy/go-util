package storage

import (
	"github.com/mzzsfy/go-util/helper"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

func clearGlsMap() {
	glsMap.Range(func(k, _ any) bool {
		glsMap.Delete(k)
		return true
	})
	atomic.StoreInt64(&glsEntryCount, 0)
}

func TestMain(m *testing.M) {
	KnowHowToUseGls()
	m.Run()
	knowHowToUseGls = false
}

func Test_itemGet(t *testing.T) {
	t.Run("value exists", func(t *testing.T) {
		item := NewGlsItem[string]()
		defer item.Delete(true)
		item.Set("testValue1")
		item.Set("testValue")
		value, ok := item.Get()
		Equal(t, true, ok)
		Equal(t, "testValue", value)
	})

	t.Run("value does not exist", func(t *testing.T) {
		defer GlsClean()
		nonexistentKey := NewGlsItem[string]()
		_, ok := nonexistentKey.Get()
		Equal(t, false, ok)
	})
}

func Test_Set(t *testing.T) {
	t.Run("set new value", func(t *testing.T) {
		item := NewGlsItem[string]()
		defer GlsClean()
		item.Set("testValue")
		value, ok := item.Get()
		Equal(t, true, ok)
		Equal(t, "testValue", value)
	})

	t.Run("overwrite existing value", func(t *testing.T) {
		item := NewGlsItem[string]()
		defer GlsClean()
		item.Set("testValue")
		item.Set("newValue")
		value, ok := item.Get()
		Equal(t, true, ok)
		Equal(t, "newValue", value)
	})
}

func Test_Del(t *testing.T) {
	t.Run("delete existing value", func(t *testing.T) {
		item := NewGlsItem[string]()
		defer GlsClean()
		item.Set("testValue")
		item.Delete()
		_, ok := item.Get()
		Equal(t, false, ok)
	})

	t.Run("delete nonexistent value", func(t *testing.T) {
		defer GlsClean()
		nonexistentKey := NewGlsItem[string]()
		_, ok := nonexistentKey.Get()
		Equal(t, false, ok)
	})
}

func Test_Clean(t *testing.T) {
	t.Run("clean existing values", func(t *testing.T) {
		item := NewGlsItem[string]()
		item.Set("testValue")
		GlsClean()
		_, ok := item.Get()
		Equal(t, false, ok)
	})

	t.Run("clean when no values exist", func(t *testing.T) {
		defer GlsClean()
		item := NewGlsItem[string]()
		GlsClean()
		_, ok := item.Get()
		Equal(t, false, ok)
	})
}

func TestGet2(t *testing.T) {
	n := 1000
	wg := helper.NewWaitGroup(n)
	item := NewGlsItem[int]()
	f := func(i int) {
		defer GlsClean()
		item.Set(i)
		for j := 0; j < 10; j++ {
			value, ok := item.Get()
			Equal(t, true, ok)
			Equal(t, i, value)
			runtime.Gosched()
		}
	}
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			f(i)
		}()
	}
	wg.Wait()
	t.Log("ok")
}

// Test_GlsKeyFn_GetById_ReadOnly 验证 GetById 为只读查询:
// 不存在的 goid 返回 (零值,false) 且不创建子表(否则伪造/已退出的 goid 会长期泄漏)
func Test_GlsKeyFn_GetById_ReadOnly(t *testing.T) {
	item := NewGlsItemWithDefault[int](42)
	const ghost = int64(987654321)
	t.Cleanup(func() {
		GlsCleanWithId(ghost)
		item.Delete()
		GlsClean()
	})

	v, ok := item.GetById(ghost)
	if ok {
		t.Fatalf("GetById 对不存在的 goid 应返回 false, got %v", v)
	}
	if v != 0 {
		t.Fatalf("GetById 对不存在的 goid 应返回零值, got %v", v)
	}
	if _, loaded := glsMap.Load(ghost); loaded {
		t.Fatal("GetById 不应为不存在的 goid 创建子表")
	}

	// 自身 Get 仍提供默认值语义
	v, ok = item.Get()
	if !ok || v != 42 {
		t.Fatalf("Get 应返回默认值 42, got %v ok=%v", v, ok)
	}
	// 跨协程读: 自身已初始化后, 按 goid 可读到
	v, ok = item.GetById(GoID())
	if !ok || v != 42 {
		t.Fatalf("GetById(自身) 应读到已初始化值 42, got %v ok=%v", v, ok)
	}
}

func Test_check(t *testing.T) {
	defer func() {
		defer clearGlsMap()
		r := recover()
		if r == nil {
			t.Errorf("check should panic")
			t.FailNow()
			return
		} else if ge, ok := r.(GlsError); !ok {
			t.Errorf("check should panic with GlsError")
			t.FailNow()
		} else if len(ge.GlsGoIds) < 2 {
			// 泄漏检查必须收集全部泄漏的 goid, 不得提前终止遍历
			t.Errorf("GlsGoIds 应收集全部泄漏 goid, 实际 %d 个: %v", len(ge.GlsGoIds), ge.GlsGoIds)
			t.FailNow()
		} else {
			t.Logf("collected %d leaked goids", len(ge.GlsGoIds))
		}
		keyIdGen = 10000
	}()
	item := NewGlsItem[string]()
	wg := helper.NewWaitGroup(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer func() { recover(); wg.Done() }()
			item.Set("testValue")
		}()
	}
	wg.Wait()
	for i := 0; i < 10_000_000; i++ {
		check()
	}
}

// runGlsBench 并行基准模型: 固定 goNum 个 goroutine, keys 创建与 GlsClean 清理均不摊入计时
// b.N 由各 goroutine 原子领批次分摊, iter 为单次迭代内容
func runGlsBench[T any](b *testing.B, goNum, keysPerG int, iter func(items []Key[T])) {
	ready := make(chan struct{}, goNum)
	start := make(chan struct{})
	clean := make(chan struct{})
	var next int64
	total := int64(b.N)
	const batch = 64
	var benchWg, cleanWg sync.WaitGroup
	benchWg.Add(goNum)
	cleanWg.Add(goNum)
	for g := 0; g < goNum; g++ {
		go func() {
			keys := make([]Key[T], 0, keysPerG)
			for i := 0; i < keysPerG; i++ {
				keys = append(keys, NewGlsItem[T]())
			}
			ready <- struct{}{}
			<-start
			for {
				s := atomic.AddInt64(&next, batch) - batch
				if s >= total {
					break
				}
				e := s + batch
				if e > total {
					e = total
				}
				for n := s; n < e; n++ {
					iter(keys)
				}
			}
			benchWg.Done()
			<-clean
			GlsClean()
			cleanWg.Done()
		}()
	}
	for i := 0; i < goNum; i++ {
		<-ready
	}
	b.ResetTimer()
	close(start)
	benchWg.Wait()
	b.StopTimer()
	close(clean)
	cleanWg.Wait()
	clearGlsMap()
}

func BenchmarkGls(b *testing.B) {
	l := 10
	x := l * 10
	l1 := l
	goNum := 1000
	b.Run("BenchmarkGls_string", func(b *testing.B) {
		value := "aaa"
		value1 := "bbb"
		runGlsBench(b, goNum, l, func(items []Key[string]) {
			for i := 0; i < x; i++ {
				items[(i+0)%l1].Set(value)
				items[(i+1)%l1].Get()
				items[(i+3)%l1].Get()
				items[(i+4)%l1].Set(value1)
				items[(i+5)%l1].Get()
				items[(i+1)%l1].Set(value)
				runtime.Gosched()
			}
		})
	})
	b.Run("BenchmarkGls_int", func(b *testing.B) {
		value := 1
		value1 := 2
		runGlsBench(b, goNum, l, func(items []Key[int]) {
			for i := 0; i < x; i++ {
				items[(i+0)%l1].Set(value)
				items[(i+1)%l1].Get()
				items[(i+2)%l1].Delete()
				items[(i+3)%l1].Get()
				items[(i+4)%l1].Set(value1)
				items[(i+5)%l1].Get()
				items[(i+1)%l1].Set(value)
				items[(i+6)%l1].Delete()
				runtime.Gosched()
			}
		})
	})
	b.Run("BenchmarkGls_obj", func(b *testing.B) {
		value := struct {
			aaa string
			bbb int
		}{"aaa", 2}
		value1 := struct {
			aaa string
			bbb int
		}{"bbb", 3}
		runGlsBench(b, goNum, l, func(items []Key[struct {
			aaa string
			bbb int
		}]) {
			for i := 0; i < x; i++ {
				items[(i+0)%l1].Set(value)
				items[(i+1)%l1].Get()
				items[(i+2)%l1].Delete()
				items[(i+3)%l1].Get()
				items[(i+4)%l1].Set(value1)
				items[(i+5)%l1].Get()
				items[(i+1)%l1].Set(value)
				items[(i+6)%l1].Delete()
				runtime.Gosched()
			}
		})
	})
}

func BenchmarkGlsSubMapType(b *testing.B) {
	b.Cleanup(func() {
		clearGlsMap()
		glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeArray[uint64, any](2)) }}
	})
	goNum := 1000
	for _, l := range []int{5, 15, 50} {
		for i := 1; i <= 3; i++ {
			i := i
			name := ""
			switch i {
			case 1:
				name = "Go"
			case 2:
				name = "Swiss"
			default:
				name = "Array"
			}
			b.Run("BenchmarkGlsSubMapType_"+strconv.Itoa(l)+"_"+name, func(b *testing.B) {
				switch i {
				case 1:
					glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeGo[uint64, any](1)) }}
				case 2:
					glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeSwiss[uint64, any](1)) }}
				default:
					glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeArray[uint64, any](1)) }}
				}
				value := 1
				value1 := 2
				x := 200
				l1 := l
				runGlsBench(b, goNum, l, func(items []Key[int]) {
					for i := 0; i < x; i++ {
						items[(i+0)%l1].Set(value)
						items[(i+1)%l1].Get()
						items[(i+2)%l1].Delete()
						items[(i+3)%l1].Get()
						items[(i+4)%l1].Set(value1)
						items[(i+5)%l1].Get()
						items[(i+1)%l1].Set(value)
						items[(i+6)%l1].Delete()
						runtime.Gosched()
					}
				})
			})
		}
	}
}
