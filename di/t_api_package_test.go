package di

import (
	"context"
	"testing"
	"time"
)

// 测试泛型 API 函数
func TestGenericAPIProvide(t *testing.T) {
	c := New()

	// 测试 Provide（使用命名服务）
	err := ProvideNamed(c, "test", func(c Container) (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("ProvideNamed failed: %v", err)
	}

	// 测试 ProvideNamed
	err = ProvideNamed(c, "named", func(c Container) (int, error) {
		return 100, nil
	})
	if err != nil {
		t.Fatalf("ProvideNamed failed: %v", err)
	}

	// 测试 ProvideValue
	err = ProvideValueNamed(c, "value", 123)
	if err != nil {
		t.Fatalf("ProvideValueNamed failed: %v", err)
	}

	// 测试 ProvideValueNamed
	err = ProvideValueNamed(c, "namedValue", 456)
	if err != nil {
		t.Fatalf("ProvideValueNamed failed: %v", err)
	}
}

func TestGenericAPIGet(t *testing.T) {
	c := New()

	// 注册服务（使用命名）
	_ = ProvideValueNamed(c, "test-int", 42)
	_ = ProvideValueNamed(c, "named-int", 100)

	// 测试 GetNamed
	intVal, err := GetNamed[int](c, "test-int")
	if err != nil {
		t.Fatalf("GetNamed failed: %v", err)
	}
	if intVal != 42 {
		t.Errorf("Expected 42, got %d", intVal)
	}

	// 测试 MustGetNamed
	mustIntVal := MustGetNamed[int](c, "named-int")
	if mustIntVal != 100 {
		t.Errorf("Expected 100, got %d", mustIntVal)
	}
}

func TestGenericAPIMustGetPanic(t *testing.T) {
	c := New()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for MustGet with non-existent service")
		}
	}()

	_ = MustGetNamed[int](c, "non-existent") // 应该 panic
}

func TestGenericAPIMustGetNamedPanic(t *testing.T) {
	c := New()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for MustGetNamed with non-existent service")
		}
	}()

	_ = MustGetNamed[int](c, "non-existent") // 应该 panic
}

func TestGenericAPIHas(t *testing.T) {
	c := New()

	// 测试未注册时
	if HasNamed[int](c, "test") {
		t.Error("Expected HasNamed to return false for non-existent service")
	}

	// 注册服务（使用命名）
	_ = ProvideValueNamed(c, "test", 42)

	// 测试已注册时
	if !HasNamed[int](c, "test") {
		t.Error("Expected HasNamed to return true for registered service")
	}
}

func TestGenericAPIGetNamedAll(t *testing.T) {
	c := New()

	// 注册多个同类型服务（使用 float64，不在黑名单中）
	_ = ProvideValueNamed(c, "first", 1.1)
	_ = ProvideValueNamed(c, "second", 2.2)
	_ = ProvideValueNamed(c, "third", 3.3)

	// 获取所有命名实例
	results, err := GetNamedAll[float64](c)
	if err != nil {
		t.Fatalf("GetNamedAll failed: %v", err)
	}

	// 验证结果
	if len(results) < 3 {
		t.Errorf("Expected at least 3 results, got %d", len(results))
	}
}

func TestGenericAPIGetNamedAllError(t *testing.T) {
	c := New()

	// 测试黑名单类型
	_, err := GetNamedAll[string](c)
	if err == nil {
		t.Error("Expected error for blacklisted type")
	}
}

// 测试 ContainerContext
func TestContainerContext(t *testing.T) {
	// 测试没有父 context
	ctx := ContainerContext{}

	deadline, ok := ctx.Deadline()
	if ok {
		t.Error("Expected no deadline")
	}
	if !deadline.IsZero() {
		t.Error("Expected zero deadline")
	}

	if ctx.Done() != nil {
		t.Error("Expected nil Done channel")
	}

	if ctx.Err() != nil {
		t.Error("Expected nil error")
	}

	if ctx.Value("key") != nil {
		t.Error("Expected nil value")
	}
}

func TestContainerContextWithParent(t *testing.T) {
	// 创建带超时的父 context
	parentCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx := ContainerContext{
		parent: parentCtx,
		err:    context.Canceled,
	}

	// 测试 Deadline
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("Expected deadline to be set")
	}
	if deadline.IsZero() {
		t.Error("Expected non-zero deadline")
	}

	// 测试 Done
	if ctx.Done() == nil {
		t.Error("Expected non-nil Done channel")
	}

	// 测试 Err
	if ctx.Err() != context.Canceled {
		t.Error("Expected Canceled error")
	}

	// 测试 Value
	key := struct{}{}
	childCtx := context.WithValue(parentCtx, key, "value")
	ctxWithValue := ContainerContext{parent: childCtx}
	if ctxWithValue.Value(key) != "value" {
		t.Error("Expected to get value from parent")
	}
}

func TestContainerContextValue(t *testing.T) {
	// 测试带值的父 context
	parentCtx := context.WithValue(context.Background(), "test-key", "test-value")
	ctx := ContainerContext{parent: parentCtx}

	// 测试获取值
	val := ctx.Value("test-key")
	if val != "test-value" {
		t.Errorf("Expected 'test-value', got %v", val)
	}

	// 测试不存在的 key
	val = ctx.Value("non-existent")
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

// 测试全局容器
func TestGlobalContainer(t *testing.T) {
	gc := GlobalContainer()
	if gc == nil {
		t.Error("GlobalContainer should not be nil")
	}

	// 测试多次调用返回同一实例（虽然可能不是同一个，但不应该为 nil）
	gc2 := GlobalContainer()
	if gc2 == nil {
		t.Error("Second GlobalContainer call should not be nil")
	}
}

// 测试错误场景
func TestGenericAPIGetError(t *testing.T) {
	c := New()

	// 测试获取不存在的服务
	_, err := GetNamed[int](c, "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent service")
	}
}

// 测试 GetNamedAll 的错误路径
func TestGetNamedAllWithBlacklist(t *testing.T) {
	c := New()

	// 尝试获取黑名单类型（string 在黑名单中）
	_, err := GetNamedAll[string](c)
	if err == nil {
		t.Error("Expected error for blacklisted type")
	}
}

// 测试空名称的 GetNamed
func TestGenericGetNamedEmptyName(t *testing.T) {
	c := New()

	// 使用命名服务
	_ = ProvideValueNamed(c, "test", 42)

	val, err := GetNamed[int](c, "test")
	if err != nil {
		t.Fatalf("GetNamed failed: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}
