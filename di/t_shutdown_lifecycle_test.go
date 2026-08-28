package di

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// shutdownHookTimeout 关闭测试的守护超时时长
const shutdownHookTimeout = 5 * time.Second

// concurrentShutdownCount 并发关闭测试的 goroutine 数量
const concurrentShutdownCount = 8

// shutdownFlagService 记录 Shutdown 调用的服务
type shutdownFlagService struct {
	called int32
}

func (s *shutdownFlagService) Shutdown(context.Context) error {
	atomic.StoreInt32(&s.called, 1)
	return nil
}

// destroyFlagService 记录销毁回调的服务
type destroyFlagService struct {
	called int32
}

func (s *destroyFlagService) OnDestroyCallback() error {
	atomic.StoreInt32(&s.called, 1)
	return nil
}

// Test_ShutdownHookGetWithoutDeadlock 关闭钩子内调用 Get 不应死锁
// 钩子内 Get 未缓存服务会触发创建路径加锁,修复前 Shutdown 持写锁执行钩子导致死锁
func Test_ShutdownHookGetWithoutDeadlock(t *testing.T) {
	c := New()
	if err := ProvideNamed(c, "dep", func(c Container) (int, error) { return 42, nil }); err != nil {
		t.Fatalf("ProvideNamed 失败: %v", err)
	}

	hookCalled := make(chan struct{})
	var getErr error
	if err := c.AppendOption(WithContainerShutdown(func(ctx context.Context) error {
		_, getErr = GetNamed[int](c, "dep")
		close(hookCalled)
		return nil
	})); err != nil {
		t.Fatalf("AppendOption 失败: %v", err)
	}

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- c.Shutdown(context.Background())
	}()

	select {
	case <-hookCalled:
	case <-time.After(shutdownHookTimeout):
		t.Fatal("Shutdown 死锁:钩子内 Get 未在时限内完成")
	}

	if err := <-shutdownDone; err != nil {
		t.Errorf("Shutdown 返回错误: %v", err)
	}
	if getErr != nil {
		t.Errorf("钩子内 Get 失败: %v", getErr)
	}
}

// Test_ShutdownConcurrentOnlyOnce 并发 Shutdown 只执行一次钩子
func Test_ShutdownConcurrentOnlyOnce(t *testing.T) {
	c := New()
	var execCount int32
	if err := c.AppendOption(WithContainerShutdown(func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		time.Sleep(50 * time.Millisecond)
		return nil
	})); err != nil {
		t.Fatalf("AppendOption 失败: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrentShutdownCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Shutdown(context.Background())
		}()
	}
	wg.Wait()

	if n := atomic.LoadInt32(&execCount); n != 1 {
		t.Errorf("期望关闭钩子只执行 1 次,实际 %d 次", n)
	}
}

// Test_ServiceLifecycleShutdownCalled 实现 ServiceLifecycle 的服务销毁回调应被执行
func Test_ServiceLifecycleShutdownCalled(t *testing.T) {
	c := New()
	svc := &shutdownFlagService{}
	if err := ProvideValue(c, svc); err != nil {
		t.Fatalf("ProvideValue 失败: %v", err)
	}
	if _, err := Get[*shutdownFlagService](c); err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if err := c.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown 失败: %v", err)
	}
	if atomic.LoadInt32(&svc.called) != 1 {
		t.Error("ServiceLifecycle.Shutdown 未被执行")
	}
}

// Test_DestroyCallbackCalled 实现 DestroyCallback 的服务销毁回调应被执行
func Test_DestroyCallbackCalled(t *testing.T) {
	c := New()
	svc := &destroyFlagService{}
	if err := ProvideValue(c, svc); err != nil {
		t.Fatalf("ProvideValue 失败: %v", err)
	}
	if _, err := Get[*destroyFlagService](c); err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if err := c.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown 失败: %v", err)
	}
	if atomic.LoadInt32(&svc.called) != 1 {
		t.Error("DestroyCallback.OnDestroyCallback 未被执行")
	}
}

// Test_PlainServiceNoDestroyHook 未实现接口且无显式钩子的服务不应注册销毁钩子
func Test_PlainServiceNoDestroyHook(t *testing.T) {
	c := New().(*container)
	if err := ProvideValueNamed(c, "plain", 42); err != nil {
		t.Fatalf("ProvideValueNamed 失败: %v", err)
	}
	if _, err := GetNamed[int](c, "plain"); err != nil {
		t.Fatalf("GetNamed 失败: %v", err)
	}
	if n := len(c.shutdown); n != 0 {
		t.Errorf("普通服务不应注册销毁钩子,实际 %d 个", n)
	}
}

// Test_TransientLifecycleNotRegistered Transient 实例不被容器持有,不应注册销毁钩子
func Test_TransientLifecycleNotRegistered(t *testing.T) {
	c := New().(*container)
	if err := ProvideNamed(c, "transient", func(c Container) (*shutdownFlagService, error) {
		return &shutdownFlagService{}, nil
	}, WithLoadMode(LoadModeTransient)); err != nil {
		t.Fatalf("ProvideNamed 失败: %v", err)
	}
	if _, err := GetNamed[*shutdownFlagService](c, "transient"); err != nil {
		t.Fatalf("GetNamed 失败: %v", err)
	}
	if n := len(c.shutdown); n != 0 {
		t.Errorf("Transient 实例无显式钩子时不应注册销毁钩子,实际 %d 个", n)
	}
}

// Test_ShutdownHookPanicStillCleanup 钩子 panic 时容器仍应完成清理并保持可观测状态
func Test_ShutdownHookPanicStillCleanup(t *testing.T) {
	c := New()
	if err := c.AppendOption(WithContainerShutdown(func(ctx context.Context) error {
		panic("hook panic")
	})); err != nil {
		t.Fatalf("AppendOption 失败: %v", err)
	}

	func() {
		defer func() { _ = recover() }()
		_ = c.Shutdown(context.Background())
	}()

	select {
	case <-c.Done():
	default:
		t.Error("钩子 panic 后容器应完成清理")
	}
}
