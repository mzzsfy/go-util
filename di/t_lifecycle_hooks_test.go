package di

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// 测试 executeInstanceDestroy 的所有路径
// 覆盖 DestroyCallback 和 ServiceLifecycle 错误路径

type testDestroyCallback struct {
	err error
}

func (t *testDestroyCallback) OnDestroyCallback() error {
	return t.err
}

type testServiceLifecycle struct {
	err error
}

func (t *testServiceLifecycle) Shutdown(ctx context.Context) error {
	return t.err
}

type testBothLifecycle struct {
	destroyErr  error
	shutdownErr error
}

func (t *testBothLifecycle) OnDestroyCallback() error {
	return t.destroyErr
}

func (t *testBothLifecycle) Shutdown(ctx context.Context) error {
	return t.shutdownErr
}

// TestExecuteInstanceDestroyAllPaths 测试所有销毁路径
func TestExecuteInstanceDestroyAllPaths(t *testing.T) {
	t.Run("DestroyCallback success", func(t *testing.T) {
		c := New().(*container)
		inst := &testDestroyCallback{err: nil}
		err := c.executeInstanceDestroy(context.Background(), inst, nil, "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("DestroyCallback error", func(t *testing.T) {
		c := New().(*container)
		inst := &testDestroyCallback{err: errors.New("destroy failed")}
		err := c.executeInstanceDestroy(context.Background(), inst, nil, "test")
		if err == nil {
			t.Error("Expected error from DestroyCallback")
		}
	})

	t.Run("ServiceLifecycle success", func(t *testing.T) {
		c := New().(*container)
		inst := &testServiceLifecycle{err: nil}
		err := c.executeInstanceDestroy(context.Background(), inst, nil, "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("ServiceLifecycle error", func(t *testing.T) {
		c := New().(*container)
		inst := &testServiceLifecycle{err: errors.New("shutdown failed")}
		err := c.executeInstanceDestroy(context.Background(), inst, nil, "test")
		if err == nil {
			t.Error("Expected error from ServiceLifecycle")
		}
	})

	t.Run("Both lifecycles", func(t *testing.T) {
		c := New().(*container)
		inst := &testBothLifecycle{destroyErr: nil, shutdownErr: nil}
		err := c.executeInstanceDestroy(context.Background(), inst, nil, "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("No lifecycle", func(t *testing.T) {
		c := New().(*container)
		inst := struct{}{}
		err := c.executeInstanceDestroy(context.Background(), inst, nil, "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestGetFromParentOrErrorAllPaths 测试父容器获取的所有路径
func TestGetFromParentOrErrorAllPaths(t *testing.T) {
	t.Run("No parent container", func(t *testing.T) {
		c := New().(*container)
		_, err := c.getFromParentOrError(reflect.TypeOf(""), "")
		if err == nil {
			t.Error("Expected error when no parent container")
		}
	})

	t.Run("Parent container success", func(t *testing.T) {
		parent := New().(*container)
		_ = ProvideValueNamed(parent, "test", 42)
		child := New().(*container)
		child.parent = parent

		inst, err := child.getFromParentOrError(reflect.TypeOf(0), "test")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if inst != 42 {
			t.Errorf("Expected 42, got %v", inst)
		}
	})

	t.Run("Parent container error", func(t *testing.T) {
		parent := New().(*container)
		child := New().(*container)
		child.parent = parent

		_, err := child.getFromParentOrError(reflect.TypeOf(0), "nonexistent")
		if err == nil {
			t.Error("Expected error from parent container")
		}
	})
}

// TestCheckCacheWithWriteLockAllPaths 测试写锁检查缓存的所有路径
func TestCheckCacheWithWriteLockAllPaths(t *testing.T) {
	t.Run("Cache hit", func(t *testing.T) {
		c := New().(*container)
		_ = ProvideValueNamed(c, "test", 42)
		_, _ = GetNamed[int](c, "test") // 触发创建

		key := typeKey(reflect.TypeOf(0), "test")
		inst, found := c.checkCacheWithWriteLock(key)
		if !found {
			t.Error("Expected to find cached instance")
		}
		if inst == nil {
			t.Error("Expected non-nil instance")
		}
	})

	t.Run("Cache miss", func(t *testing.T) {
		c := New().(*container)
		key := typeKey(reflect.TypeOf(0), "nonexistent")
		_, found := c.checkCacheWithWriteLock(key)
		if found {
			t.Error("Expected cache miss")
		}
	})

	t.Run("Loading flag set", func(t *testing.T) {
		c := New().(*container)
		key := "test-key"
		c.loading[key] = true

		_, found := c.checkCacheWithWriteLock(key)
		if found {
			t.Error("Expected not found when loading flag is set")
		}
	})
}
