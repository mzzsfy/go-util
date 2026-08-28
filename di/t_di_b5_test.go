package di

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

// --- BDD: mapConfigSource 单元语义 ---

// Given 空配置源
// When Get 未命中的键
// Then 返回非 nil Value, Any 为 nil, Has 为 false
func Test_MapConfigSource_GetMissing(t *testing.T) {
	t.Parallel()
	src := NewMapConfigSource()

	v := src.Get("missing")
	if v == nil {
		t.Fatal("Get 未命中应返回非 nil Value")
	}
	if v.Any() != nil {
		t.Fatalf("未命中 Any=%v, want nil", v.Any())
	}
	if src.Has("missing") {
		t.Fatal("Has 未命中应为 false")
	}
}

// Given 已写入的键值
// When Get 与 Has
// Then 返回原值且存在
func Test_MapConfigSource_SetGetRoundTrip(t *testing.T) {
	t.Parallel()
	src := NewMapConfigSource()

	src.Set("k1", "v1")
	src.Set("k2", 42)

	if !src.Has("k1") || !src.Has("k2") {
		t.Fatal("Has should be true after Set")
	}
	if got := src.Get("k1").String(); got != "v1" {
		t.Fatalf("k1=%q want v1", got)
	}
	if got := src.Get("k2").Any(); got != 42 {
		t.Fatalf("k2=%v want 42", got)
	}

	// 覆盖写
	src.Set("k1", "v2")
	if got := src.Get("k1").String(); got != "v2" {
		t.Fatalf("k1 after overwrite=%q want v2", got)
	}
}

// Given 已有多个键
// When Clear
// Then 全部清空
func Test_MapConfigSource_Clear(t *testing.T) {
	t.Parallel()
	src := NewMapConfigSource()
	src.Set("a", "1")
	src.Set("b", "2")

	src.Clear()

	if src.Has("a") || src.Has("b") {
		t.Fatal("Has should be false after Clear")
	}
	if v := src.Get("a"); v == nil || v.Any() != nil {
		t.Fatalf("Get after Clear = %v, want nil value", v)
	}
}

// Given 并发读写配置源
// When 多 goroutine Set/Get/Has/Clear
// Then 无竞态崩溃
func Test_MapConfigSource_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	src := NewMapConfigSource()

	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				key := string(rune('a' + i%10))
				src.Set(key, i)
				_ = src.Get(key).Any()
				_ = src.Has(key)
				if i%50 == 49 {
					src.Clear()
				}
			}
		}(w)
	}
	wg.Wait()
}

// --- BDD: ChildScope 与加载模式组合 ---

// Given 父容器注册 Transient 服务
// When 子 scope 多次获取
// Then 每次都新建实例
func Test_ChildScope_TransientEachCallNew(t *testing.T) {
	t.Parallel()
	parent := New().(*container)

	type scoped int
	calls := 0
	if err := parent.ProvideNamedWith("scoped", func(c Container) (scoped, error) {
		calls++
		return scoped(calls), nil
	}, WithLoadMode(LoadModeTransient)); err != nil {
		t.Fatalf("ProvideNamedWith() error = %v", err)
	}

	child := parent.CreateChildScope().(*container)
	typ := reflect.TypeOf(scoped(0))
	for want := 1; want <= 3; want++ {
		v, err := child.GetNamed(typ, "scoped")
		if err != nil {
			t.Fatalf("GetNamed() error = %v", err)
		}
		if got := v.(scoped); got != scoped(want) {
			t.Fatalf("transient instance = %d, want %d", got, want)
		}
	}
	if calls != 3 {
		t.Fatalf("constructor calls=%d want=3", calls)
	}

	_ = parent.Shutdown(context.Background())
}

// Given 父容器注册 Singleton 服务
// When 父与子 scope 各自获取
// Then 共享同一实例
func Test_ChildScope_SingletonShared(t *testing.T) {
	t.Parallel()
	parent := New().(*container)

	type shared struct{ n int }
	if err := parent.ProvideNamedWith("shared", func(c Container) (*shared, error) {
		return &shared{n: 7}, nil
	}, WithLoadMode(LoadModeDefault)); err != nil {
		t.Fatalf("ProvideNamedWith() error = %v", err)
	}

	child := parent.CreateChildScope().(*container)
	typ := reflect.TypeOf(&shared{})
	pv, err := parent.GetNamed(typ, "shared")
	if err != nil {
		t.Fatalf("parent GetNamed error = %v", err)
	}
	cv, err := child.GetNamed(typ, "shared")
	if err != nil {
		t.Fatalf("child GetNamed error = %v", err)
	}
	if pv != cv {
		t.Fatal("singleton should be shared between parent and child scope")
	}

	_ = parent.Shutdown(context.Background())
}
