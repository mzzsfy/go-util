package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// ========== 1. 并发首次 Get 不应误报循环依赖 ==========

// Test_ConcurrentFirstGetNoFalseCircular 并发首次 Get 同一实例应等待创建完成而非误报循环依赖
func Test_ConcurrentFirstGetNoFalseCircular(t *testing.T) {
	c := New()
	var providerCalls int32
	if err := ProvideNamed(c, "slow", func(cont Container) (*slowService, error) {
		atomic.AddInt32(&providerCalls, 1)
		// 拉大创建窗口,让并发 Get 落在创建期间
		time.Sleep(100 * time.Millisecond)
		return &slowService{}, nil
	}); err != nil {
		t.Fatalf("ProvideNamed 失败: %v", err)
	}

	const goroutines = 8
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := GetNamed[*slowService](c, "slow")
			if err != nil {
				errCh <- err
				return
			}
			if v == nil {
				errCh <- fmt.Errorf("期望非 nil 实例")
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("并发 Get 不应失败: %v", err)
	}
	if n := atomic.LoadInt32(&providerCalls); n != 1 {
		t.Errorf("provider 应只执行 1 次,实际 %d 次", n)
	}
}

// slowService 空服务类型
type slowService struct{}

// Test_CircularDependencyStillDetected 同 goroutine 真实循环依赖仍应报错而非死锁
func Test_CircularDependencyStillDetected(t *testing.T) {
	c := New()

	type circA struct{}
	type circB struct{ A *circA }

	if err := ProvideNamed(c, "", func(cont Container) (*circA, error) {
		_, err := Get[*circB](cont)
		if err != nil {
			return nil, err
		}
		return &circA{}, nil
	}); err != nil {
		t.Fatalf("Provide circA 失败: %v", err)
	}
	if err := ProvideNamed(c, "", func(cont Container) (*circB, error) {
		// B 依赖 A,形成 A -> B -> A 循环
		_, err := Get[*circA](cont)
		if err != nil {
			return nil, err
		}
		return &circB{}, nil
	}); err != nil {
		t.Fatalf("Provide circB 失败: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := Get[*circA](c)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("循环依赖应返回错误")
		}
	case <-time.After(shutdownHookTimeout):
		t.Fatal("循环依赖检测死锁")
	}
}

// ========== 2. afterCreate 失败应回滚缓存 ==========

// Test_AfterCreateFailureRollsBackCache afterCreate 失败后半初始化实例不应留缓存
func Test_AfterCreateFailureRollsBackCache(t *testing.T) {
	c := New().(*container)
	var providerCalls int32
	if err := ProvideNamed(c, "", func(cont Container) (*rollbackService, error) {
		atomic.AddInt32(&providerCalls, 1)
		return &rollbackService{}, nil
	}, WithAfterCreate(func(cont Container, info EntryInfo) (any, error) {
		return nil, errors.New("afterCreate failed")
	})); err != nil {
		t.Fatalf("Provide 失败: %v", err)
	}

	_, err := Get[*rollbackService](c)
	if err == nil {
		t.Fatal("afterCreate 失败应返回错误")
	}

	if n := c.GetInstanceCount(); n != 0 {
		t.Errorf("失败后实例不应留在缓存,实际 %d 个", n)
	}
	if n := len(c.shutdown); n != 0 {
		t.Errorf("失败后不应注册销毁钩子,实际 %d 个", n)
	}

	// 再次 Get:应重新创建并再次失败,而非拿到半初始化实例
	_, err = Get[*rollbackService](c)
	if err == nil {
		t.Fatal("第二次 Get 不应拿到半初始化实例")
	}
	if n := atomic.LoadInt32(&providerCalls); n != 2 {
		t.Errorf("第二次 Get 应重新创建,provider 实际执行 %d 次", n)
	}
}

// rollbackService 空服务类型
type rollbackService struct{}

// ========== 3. GetNamedAll 并发注册安全(-race 实证) ==========

// Test_GetNamedAllConcurrentProvide 并发注册与 GetNamedAll 不应产生数据竞争
func Test_GetNamedAllConcurrentProvide(t *testing.T) {
	c := New()

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			_ = ProvideNamed(c, fmt.Sprintf("s%d", i), func(cont Container) (*raceService, error) {
				return &raceService{}, nil
			})
		}
	}()

	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if _, err := c.GetNamedAll(reflect.TypeOf(&raceService{})); err != nil {
			t.Errorf("GetNamedAll 失败: %v", err)
			break
		}
	}
	close(stop)
	wg.Wait()
}

// raceService 空服务类型
type raceService struct{}

// ========== 4. int 到 string 注入不应产生 rune 语义错值 ==========

type intToStringTarget struct {
	Value string `di.config:"port"`
}

// Test_IntToStringInjection 配置注入 int(65) 到 string 字段应得 "65" 而非 "A"
func Test_IntToStringInjection(t *testing.T) {
	c := New()
	source := NewMapConfigSource()
	source.Set("port", 65)
	c.SetConfigSource(source)
	if err := ProvideNamed(c, "", func(cont Container) (*intToStringTarget, error) {
		return &intToStringTarget{}, nil
	}); err != nil {
		t.Fatalf("Provide 失败: %v", err)
	}

	v, err := Get[*intToStringTarget](c)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if v.Value != "65" {
		t.Errorf("期望 \"65\",实际 %q", v.Value)
	}
}

// ========== 5. provider 返回 nil 应返回错误而非 panic ==========

// Test_ProviderReturnsNilError provider 返回 nil 应返回明确错误
func Test_ProviderReturnsNilError(t *testing.T) {
	c := New()
	if err := ProvideNamed(c, "", func(cont Container) (nilResultService, error) {
		return nil, nil
	}); err != nil {
		t.Fatalf("Provide 失败: %v", err)
	}

	v, err := Get[nilResultService](c)
	if err == nil {
		t.Fatal("provider 返回 nil 应返回错误")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("错误信息应说明 provider 返回 nil,实际: %v", err)
	}
	if v != nil {
		t.Errorf("出错时实例应为 nil,实际 %v", v)
	}
}

// nilResultService 接口类型,provider 返回无类型 nil 时触发类型断言 panic 路径
type nilResultService interface{ Marker() }

// ========== 6. 依赖序销毁:被依赖者最后销毁 ==========

type destroyOrderA struct{}
type destroyOrderB struct {
	A *destroyOrderA `di:""`
}

// Test_DestroyOrderLIFO 被依赖者 A 应晚于依赖者 B 销毁
func Test_DestroyOrderLIFO(t *testing.T) {
	c := New()
	var mu sync.Mutex
	var order []string

	if err := ProvideNamed(c, "", func(cont Container) (*destroyOrderA, error) {
		return &destroyOrderA{}, nil
	}, WithAfterDestroy(func(cont Container, info EntryInfo) {
		mu.Lock()
		order = append(order, "A")
		mu.Unlock()
	})); err != nil {
		t.Fatalf("Provide A 失败: %v", err)
	}
	if err := ProvideNamed(c, "", func(cont Container) (*destroyOrderB, error) {
		return &destroyOrderB{}, nil
	}, WithAfterDestroy(func(cont Container, info EntryInfo) {
		mu.Lock()
		order = append(order, "B")
		mu.Unlock()
	})); err != nil {
		t.Fatalf("Provide B 失败: %v", err)
	}

	if _, err := Get[*destroyOrderB](c); err != nil {
		t.Fatalf("Get B 失败: %v", err)
	}

	if err := c.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown 失败: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 {
		t.Fatalf("期望 2 次销毁,实际 %v", order)
	}
	if order[0] != "B" || order[1] != "A" {
		t.Errorf("期望销毁顺序 [B A](依赖者先销毁),实际 %v", order)
	}
}

// ========== 7. CreateChildScope 与 Shutdown 并发安全(-race 实证) ==========

// Test_CreateChildScopeConcurrentShutdown 子容器注册与父容器 Shutdown 并发不应产生数据竞争
func Test_CreateChildScopeConcurrentShutdown(t *testing.T) {
	parent := New()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = parent.Shutdown(context.Background())
	}()
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		child := parent.CreateChildScope()
		_ = child
	}
	<-done
}

// ========== 8. Start 与 AppendStartupOption 并发安全(-race 实证) ==========

// Test_StartConcurrentAppendStartupOption 启动钩子快照/置空与追加选项并发不应产生数据竞争
func Test_StartConcurrentAppendStartupOption(t *testing.T) {
	c := New()
	startDone := make(chan error, 1)
	go func() {
		startDone <- c.Start()
	}()
	for i := 0; i < 100; i++ {
		_ = c.AppendOption(WithContainerOnStart(func(cont Container) error { return nil }))
	}
	if err := <-startDone; err != nil {
		t.Logf("Start 返回错误(可接受): %v", err)
	}
}

// ========== 9. Lazy 依赖应沿父链解析 ==========

type lazyDepService struct{}
type lazyDependentService struct {
	Dep *lazyDepService `di:""`
}

// Test_LazyDependencyResolvedFromParent 子容器懒加载服务依赖父容器 provider 应可解析
func Test_LazyDependencyResolvedFromParent(t *testing.T) {
	parent := New()
	if err := Provide(parent, func(cont Container) (*lazyDepService, error) {
		return &lazyDepService{}, nil
	}); err != nil {
		t.Fatalf("Provide 父容器服务失败: %v", err)
	}

	child := parent.CreateChildScope()
	if err := Provide(child, func(cont Container) (*lazyDependentService, error) {
		return &lazyDependentService{}, nil
	}, WithLoadMode(LoadModeLazy)); err != nil {
		t.Fatalf("Provide 子容器服务失败: %v", err)
	}

	v, err := Get[*lazyDependentService](child)
	if err != nil {
		t.Fatalf("子容器懒加载服务应能解析父容器依赖: %v", err)
	}
	if v.Dep == nil {
		t.Fatal("依赖未注入")
	}
}

// ========== 10. Shutdown 后禁止复活 ==========

// Test_StartAfterShutdownForbidden Shutdown 后 Start 应返回错误
func Test_StartAfterShutdownForbidden(t *testing.T) {
	c := New()
	hookCalls := 0
	if err := c.AppendOption(WithContainerOnStart(func(cont Container) error {
		hookCalls++
		return nil
	})); err != nil {
		t.Fatalf("AppendOption 失败: %v", err)
	}

	if err := c.Start(); err != nil {
		t.Fatalf("第一次 Start 失败: %v", err)
	}
	if err := c.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown 失败: %v", err)
	}

	err := c.Start()
	if err == nil {
		t.Fatal("Shutdown 后 Start 应返回错误,不允许复活")
	}
	if hookCalls != 1 {
		t.Errorf("启动钩子不应重复执行,实际 %d 次", hookCalls)
	}
	select {
	case <-c.Done():
	default:
		t.Error("Shutdown 后 Done 通道应保持关闭")
	}
}

// ========== 11. 信号触发 Shutdown 出错时退出码非零 ==========

// Test_ShutdownOnSignalsErrorExitCode Shutdown 失败时退出码应为非零
func Test_ShutdownOnSignalsErrorExitCode(t *testing.T) {
	c := New()
	env := newSignalTestEnv(t)
	if err := c.AppendOption(WithContainerShutdown(func(ctx context.Context) error {
		return errors.New("hook failure")
	})); err != nil {
		t.Fatalf("AppendOption 失败: %v", err)
	}

	c.ShutdownOnSignals()
	env.sendSignal(syscall.SIGTERM)

	if code := env.waitExit(); code == 0 {
		t.Error("Shutdown 出错时退出码应为非零")
	}
}

// ========== 12. 多个关闭钩子失败应收集全部错误 ==========

// Test_ShutdownCollectsAllHookErrors 多个钩子失败时错误应全部保留
func Test_ShutdownCollectsAllHookErrors(t *testing.T) {
	c := New()
	errA := errors.New("hook-a failed")
	errB := errors.New("hook-b failed")
	errC := errors.New("hook-c failed")
	if err := c.AppendOption(
		WithContainerShutdown(func(ctx context.Context) error { return errA }),
		WithContainerShutdown(func(ctx context.Context) error { return errB }),
		WithContainerShutdown(func(ctx context.Context) error { return errC }),
	); err != nil {
		t.Fatalf("AppendOption 失败: %v", err)
	}

	err := c.Shutdown(context.Background())
	if err == nil {
		t.Fatal("Shutdown 应返回错误")
	}
	for _, want := range []error{errA, errB, errC} {
		// go1.18 的 errors.Is 不支持 Unwrap() []error, 兼容退化为信息包含检查
		if !errors.Is(err, want) && !strings.Contains(err.Error(), want.Error()) {
			t.Errorf("聚合错误应包含 %v,实际: %v", want, err)
		}
	}
}

// ========== 13. 销毁回调错误分支在原测试文件中修正(Test_CreateDestroyHookInstanceDestroyError) ==========
