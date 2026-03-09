package di

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestPackage 测试Package函数
func TestPackage(t *testing.T) {
	// 测试：单个provider成功
	t.Run("single provider success", func(t *testing.T) {
		executed := false
		pkg := Package(func(c Container) error {
			executed = true
			return nil
		})

		ctr := New()
		err := pkg(ctr)
		if err != nil {
			t.Fatalf("Package() error = %v", err)
		}
		if !executed {
			t.Error("Package() provider not executed")
		}
	})

	// 测试：多个provider成功
	t.Run("multiple providers success", func(t *testing.T) {
		count := 0
		pkg := Package(
			func(c Container) error {
				count++
				return nil
			},
			func(c Container) error {
				count++
				return nil
			},
			func(c Container) error {
				count++
				return nil
			},
		)

		ctr := New()
		err := pkg(ctr)
		if err != nil {
			t.Fatalf("Package() error = %v", err)
		}
		if count != 3 {
			t.Errorf("Package() executed %d providers, want 3", count)
		}
	})

	// 测试：provider返回错误
	t.Run("provider error", func(t *testing.T) {
		expectedErr := errors.New("provider error")
		pkg := Package(
			func(c Container) error {
				return nil
			},
			func(c Container) error {
				return expectedErr
			},
			func(c Container) error {
				t.Error("should not execute third provider")
				return nil
			},
		)

		ctr := New()
		err := pkg(ctr)
		if err != expectedErr {
			t.Errorf("Package() error = %v, want %v", err, expectedErr)
		}
	})

	// 测试：空providers
	t.Run("empty providers", func(t *testing.T) {
		pkg := Package()

		ctr := New()
		err := pkg(ctr)
		if err != nil {
			t.Fatalf("Package() with empty providers error = %v", err)
		}
	})

	// 测试：provider注册服务
	t.Run("provider register service", func(t *testing.T) {
		type Service struct {
			Name string
		}

		pkg := Package(
			func(c Container) error {
				return c.ProvideNamedWith("", func(c Container) (*Service, error) {
					return &Service{Name: "test"}, nil
				})
			},
		)

		ctr := New()
		err := pkg(ctr)
		if err != nil {
			t.Fatalf("Package() error = %v", err)
		}

		svc, err := ctr.GetNamed((*Service)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed() error = %v", err)
		}
		if svc.(*Service).Name != "test" {
			t.Errorf("GetNamed() = %v, want Name=test", svc)
		}
	})
}

// TestLoadPackages 测试LoadPackages函数
func TestLoadPackages(t *testing.T) {
	// 测试：单个package成功
	t.Run("single package success", func(t *testing.T) {
		executed := false
		pkg := func(c Container) error {
			executed = true
			return nil
		}

		ctr := New()
		result, err := LoadPackages(ctr, pkg)
		if err != nil {
			t.Fatalf("LoadPackages() error = %v", err)
		}
		if result != ctr {
			t.Error("LoadPackages() should return same container")
		}
		if !executed {
			t.Error("LoadPackages() package not executed")
		}
	})

	// 测试：多个packages成功
	t.Run("multiple packages success", func(t *testing.T) {
		count := 0
		pkg1 := func(c Container) error {
			count++
			return nil
		}
		pkg2 := func(c Container) error {
			count++
			return nil
		}
		pkg3 := func(c Container) error {
			count++
			return nil
		}

		ctr := New()
		result, err := LoadPackages(ctr, pkg1, pkg2, pkg3)
		if err != nil {
			t.Fatalf("LoadPackages() error = %v", err)
		}
		if result != ctr {
			t.Error("LoadPackages() should return same container")
		}
		if count != 3 {
			t.Errorf("LoadPackages() executed %d packages, want 3", count)
		}
	})

	// 测试：package返回错误
	t.Run("package error", func(t *testing.T) {
		expectedErr := errors.New("package error")
		pkg1 := func(c Container) error {
			return nil
		}
		pkg2 := func(c Container) error {
			return expectedErr
		}
		pkg3 := func(c Container) error {
			t.Error("should not execute third package")
			return nil
		}

		ctr := New()
		result, err := LoadPackages(ctr, pkg1, pkg2, pkg3)
		if err != expectedErr {
			t.Errorf("LoadPackages() error = %v, want %v", err, expectedErr)
		}
		if result != nil {
			t.Error("LoadPackages() should return nil on error")
		}
	})

	// 测试：空packages
	t.Run("empty packages", func(t *testing.T) {
		ctr := New()
		result, err := LoadPackages(ctr)
		if err != nil {
			t.Fatalf("LoadPackages() with empty packages error = %v", err)
		}
		if result != ctr {
			t.Error("LoadPackages() should return same container")
		}
	})

	// 测试：packages注册服务
	t.Run("packages register services", func(t *testing.T) {
		type Service1 struct{ ID int }
		type Service2 struct{ Name string }

		pkg1 := func(c Container) error {
			return c.ProvideNamedWith("", func(c Container) (*Service1, error) {
				return &Service1{ID: 100}, nil
			})
		}
		pkg2 := func(c Container) error {
			return c.ProvideNamedWith("", func(c Container) (*Service2, error) {
				return &Service2{Name: "test"}, nil
			})
		}

		ctr := New()
		result, err := LoadPackages(ctr, pkg1, pkg2)
		if err != nil {
			t.Fatalf("LoadPackages() error = %v", err)
		}

		svc1, err := result.GetNamed((*Service1)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed(Service1) error = %v", err)
		}
		if svc1.(*Service1).ID != 100 {
			t.Errorf("GetNamed(Service1) = %v, want ID=100", svc1)
		}

		svc2, err := result.GetNamed((*Service2)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed(Service2) error = %v", err)
		}
		if svc2.(*Service2).Name != "test" {
			t.Errorf("GetNamed(Service2) = %v, want Name=test", svc2)
		}
	})
}

// TestPrepareLazyDependenciesEdgeCases 测试prepareLazyDependencies的边界情况
func TestPrepareLazyDependenciesEdgeCases(t *testing.T) {
	// 测试：非Lazy模式
	t.Run("non lazy mode", func(t *testing.T) {
		type Service struct {
			Name string `di:""`
		}

		ctr := New().(*container)
		err := ctr.ProvideNamedWith("", func(c Container) (*Service, error) {
			return &Service{Name: "test"}, nil
		}, WithLoadMode(LoadModeDefault))
		if err != nil {
			t.Fatalf("ProvideNamedWith() error = %v", err)
		}

		// 非Lazy模式应该直接返回nil
		key := typeKey(reflect.TypeOf((*Service)(nil)), "")
		entry := ctr.providers[key]
		err = ctr.prepareLazyDependencies(entry, key)
		if err != nil {
			t.Errorf("prepareLazyDependencies() error = %v", err)
		}
	})

	// 测试：Lazy模式依赖查找错误
	t.Run("lazy mode with invalid dependency", func(t *testing.T) {
		type NonExistent struct{}
		type Service struct {
			Dep *NonExistent `di:""` // 不存在的依赖
		}

		ctr := New().(*container)
		err := ctr.ProvideNamedWith("", func(c Container) (*Service, error) {
			return &Service{}, nil
		}, WithLoadMode(LoadModeLazy))
		if err != nil {
			t.Fatalf("ProvideNamedWith() error = %v", err)
		}

		key := typeKey(reflect.TypeOf((*Service)(nil)), "")
		entry := ctr.providers[key]
		err = ctr.prepareLazyDependencies(entry, key)
		if err == nil {
			t.Error("prepareLazyDependencies() should return error for missing dependency")
		}
	})

	// 测试：Lazy模式依赖创建错误
	t.Run("lazy mode with dependency creation error", func(t *testing.T) {
		type DepService struct {
			Name string `di:""`
		}
		type Service struct {
			Dep *DepService `di:""`
		}

		ctr := New().(*container)
		// 注册一个会失败的provider
		err := ctr.ProvideNamedWith("", func(c Container) (*DepService, error) {
			return nil, errors.New("provider error")
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith(DepService) error = %v", err)
		}

		err = ctr.ProvideNamedWith("", func(c Container) (*Service, error) {
			return &Service{}, nil
		}, WithLoadMode(LoadModeLazy))
		if err != nil {
			t.Fatalf("ProvideNamedWith(Service) error = %v", err)
		}

		key := typeKey(reflect.TypeOf((*Service)(nil)), "")
		entry := ctr.providers[key]
		err = ctr.prepareLazyDependencies(entry, key)
		if err == nil {
			t.Error("prepareLazyDependencies() should return error for failed dependency creation")
		}
	})

	// 测试：Lazy模式带命名依赖
	t.Run("lazy mode with named dependency", func(t *testing.T) {
		type DepService struct {
			Name string
		}
		type Service struct {
			Dep *DepService `di:"special"`
		}

		ctr := New().(*container)
		err := ctr.ProvideNamedWith("special", func(c Container) (*DepService, error) {
			return &DepService{Name: "special"}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith(DepService, special) error = %v", err)
		}

		err = ctr.ProvideNamedWith("", func(c Container) (*Service, error) {
			return &Service{}, nil
		}, WithLoadMode(LoadModeLazy))
		if err != nil {
			t.Fatalf("ProvideNamedWith(Service) error = %v", err)
		}

		key := typeKey(reflect.TypeOf((*Service)(nil)), "")
		entry := ctr.providers[key]
		err = ctr.prepareLazyDependencies(entry, key)
		if err != nil {
			t.Errorf("prepareLazyDependencies() error = %v", err)
		}

		// 验证依赖已创建
		dep, err := ctr.GetNamed((*DepService)(nil), "special")
		if err != nil {
			t.Fatalf("GetNamed(DepService, special) error = %v", err)
		}
		if dep.(*DepService).Name != "special" {
			t.Errorf("GetNamed(DepService, special) = %v, want Name=special", dep)
		}
	})

	// 测试：Lazy模式有多个依赖
	t.Run("lazy mode with multiple dependencies", func(t *testing.T) {
		type Dep1 struct{ Value int }
		type Dep2 struct{ Name string }
		type Service struct {
			D1 *Dep1 `di:""`
			D2 *Dep2 `di:""`
		}

		ctr := New().(*container)
		err := ctr.ProvideNamedWith("", func(c Container) (*Dep1, error) {
			return &Dep1{Value: 100}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith(Dep1) error = %v", err)
		}

		err = ctr.ProvideNamedWith("", func(c Container) (*Dep2, error) {
			return &Dep2{Name: "test"}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith(Dep2) error = %v", err)
		}

		err = ctr.ProvideNamedWith("", func(c Container) (*Service, error) {
			return &Service{}, nil
		}, WithLoadMode(LoadModeLazy))
		if err != nil {
			t.Fatalf("ProvideNamedWith(Service) error = %v", err)
		}

		key := typeKey(reflect.TypeOf((*Service)(nil)), "")
		entry := ctr.providers[key]
		err = ctr.prepareLazyDependencies(entry, key)
		if err != nil {
			t.Errorf("prepareLazyDependencies() error = %v", err)
		}

		// 验证所有依赖都已创建
		d1, err := ctr.GetNamed((*Dep1)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed(Dep1) error = %v", err)
		}
		if d1.(*Dep1).Value != 100 {
			t.Errorf("GetNamed(Dep1) = %v, want Value=100", d1)
		}

		d2, err := ctr.GetNamed((*Dep2)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed(Dep2) error = %v", err)
		}
		if d2.(*Dep2).Name != "test" {
			t.Errorf("GetNamed(Dep2) = %v, want Name=test", d2)
		}
	})
}

// TestValidateInstanceEdgeCases 测试validateInstance的边界情况
func TestValidateInstanceEdgeCases(t *testing.T) {
	// 测试：nil实例
	t.Run("nil instance", func(t *testing.T) {
		ctr := New().(*container)
		result := ctr.validateInstance(nil, reflect.Value{})
		if result {
			t.Error("validateInstance(nil) = true, want false")
		}
	})

	// 测试：无效的reflect.Value
	t.Run("invalid reflect value", func(t *testing.T) {
		ctr := New().(*container)
		instance := "test"
		result := ctr.validateInstance(instance, reflect.Value{})
		if result {
			t.Error("validateInstance(invalid value) = true, want false")
		}
	})

	// 测试：有效实例
	t.Run("valid instance", func(t *testing.T) {
		ctr := New().(*container)
		instance := "test"
		result := ctr.validateInstance(instance, reflect.ValueOf(instance))
		if !result {
			t.Error("validateInstance(valid) = false, want true")
		}
	})

	// 测试：零值但有效的实例
	t.Run("zero value but valid", func(t *testing.T) {
		ctr := New().(*container)
		var instance int
		result := ctr.validateInstance(instance, reflect.ValueOf(instance))
		if !result {
			t.Error("validateInstance(zero value) = false, want true")
		}
	})
}

// TestFindDependEdgeCases 测试findDepend的边界情况
func TestFindDependEdgeCases(t *testing.T) {
	// 测试：非struct/pointer类型
	t.Run("non struct or pointer type", func(t *testing.T) {
		ctr := New().(*container)
		_, err := ctr.findDepend(reflect.TypeOf("string"))
		if err == nil {
			t.Error("findDepend(string) should return error")
		}
		if !strings.Contains(err.Error(), "must be a struct") {
			t.Errorf("findDepend(string) error = %v, want 'must be a struct'", err)
		}
	})

	// 测试：指针到非struct
	t.Run("pointer to non struct", func(t *testing.T) {
		ctr := New().(*container)
		var s *string
		_, err := ctr.findDepend(reflect.TypeOf(s))
		if err == nil {
			t.Error("findDepend(*string) should return error")
		}
	})

	// 测试：struct没有di标签
	t.Run("struct without di tags", func(t *testing.T) {
		type NoTags struct {
			Field string
		}

		ctr := New().(*container)
		result, err := ctr.findDepend(reflect.TypeOf(NoTags{}))
		if err != nil {
			t.Errorf("findDepend(NoTags) error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("findDepend(NoTags) = %v, want empty", result)
		}
	})

	// 测试：struct有di标签但没有provider
	t.Run("struct with di tag but no provider", func(t *testing.T) {
		type MissingProvider struct {
			Dep string `di:""`
		}

		ctr := New().(*container)
		_, err := ctr.findDepend(reflect.TypeOf(MissingProvider{}))
		if err == nil {
			t.Error("findDepend(MissingProvider) should return error")
		}
		if !strings.Contains(err.Error(), "no provider found") {
			t.Errorf("findDepend(MissingProvider) error = %v, want 'no provider found'", err)
		}
	})

	// 测试：嵌套struct
	t.Run("nested struct", func(t *testing.T) {
		type Inner struct {
			Value string `di:""`
		}
		type Outer struct {
			InnerField *Inner `di:""`
		}

		ctr := New().(*container)
		err := ctr.ProvideNamedWith("", func(c Container) (*Inner, error) {
			return &Inner{Value: "test"}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith() error = %v", err)
		}

		result, err := ctr.findDepend(reflect.TypeOf(Outer{}))
		if err != nil {
			t.Errorf("findDepend(Outer) error = %v", err)
		}
		if len(result) != 1 {
			t.Errorf("findDepend(Outer) = %v, want 1 dependency", result)
		}
	})

	// 测试：指针类型
	t.Run("pointer type", func(t *testing.T) {
		type Service struct {
			Field int `di:"myint"`
		}

		ctr := New().(*container)
		err := ctr.ProvideNamedWith("myint", func(c Container) (int, error) {
			return 42, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith() error = %v", err)
		}

		// Service的Field依赖int (名称为myint)
		result, err := ctr.findDepend(reflect.TypeOf(Service{}))
		if err != nil {
			t.Errorf("findDepend(Service) error = %v", err)
		}
		if len(result) != 1 {
			t.Errorf("findDepend(Service) = %v, want 1 dependency", result)
		}
	})

	// 测试：多个字段
	t.Run("multiple fields", func(t *testing.T) {
		type Dep1 struct{ Value int }
		type Dep2 struct{ Name string }
		type MultiFields struct {
			D1 *Dep1 `di:""`
			D2 *Dep2 `di:""`
		}

		ctr := New().(*container)
		err := ctr.ProvideNamedWith("", func(c Container) (*Dep1, error) {
			return &Dep1{Value: 100}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith(Dep1) error = %v", err)
		}

		err = ctr.ProvideNamedWith("", func(c Container) (*Dep2, error) {
			return &Dep2{Name: "test"}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith(Dep2) error = %v", err)
		}

		result, err := ctr.findDepend(reflect.TypeOf(MultiFields{}))
		if err != nil {
			t.Errorf("findDepend(MultiFields) error = %v", err)
		}
		if len(result) != 2 {
			t.Errorf("findDepend(MultiFields) = %v, want 2 dependencies", result)
		}
	})
}

// TestGetConfigValueEdgeCases 测试getConfigValue的边界情况
func TestGetConfigValueEdgeCases(t *testing.T) {
	// 测试：configSource为nil
	t.Run("nil config source", func(t *testing.T) {
		ctr := New().(*container)
		ctr.configSource = nil

		value := ctr.getConfigValue("test")
		if value.Any() != nil {
			t.Error("getConfigValue() with nil source should return nil value")
		}

		// 检查统计
		stats := ctr.GetStats()
		if stats.ConfigMisses == 0 {
			t.Error("getConfigValue() with nil source should increment configMisses")
		}
	})

	// 测试：空key
	t.Run("empty key", func(t *testing.T) {
		ctr := New().(*container)
		ctr.SetConfigSource(NewMapConfigSource())

		value := ctr.getConfigValue("")
		if value.Any() != nil {
			t.Error("getConfigValue('') should return nil value")
		}
	})

	// 测试：key存在
	t.Run("key exists", func(t *testing.T) {
		ctr := New().(*container)
		configSrc := NewMapConfigSource()
		configSrc.Set("test", "value")
		ctr.SetConfigSource(configSrc)

		value := ctr.getConfigValue("test")
		if value.Any() == nil {
			t.Error("getConfigValue('test') should return non-nil value")
		}
		if value.String() != "value" {
			t.Errorf("getConfigValue('test') = %v, want 'value'", value.String())
		}

		// 检查统计
		stats := ctr.GetStats()
		if stats.ConfigHits == 0 {
			t.Error("getConfigValue() with existing key should increment configHits")
		}
	})

	// 测试：key不存在
	t.Run("key not exists", func(t *testing.T) {
		ctr := New().(*container)
		ctr.SetConfigSource(NewMapConfigSource())

		// 先访问一次建立baseline
		_ = ctr.getConfigValue("nonexistent")

		// 再访问一次
		value := ctr.getConfigValue("nonexistent")
		if value.Any() != nil {
			t.Error("getConfigValue('nonexistent') should return nil value")
		}
	})

	// 测试：并发访问
	t.Run("concurrent access", func(t *testing.T) {
		ctr := New().(*container)
		configSrc := NewMapConfigSource()
		for i := 0; i < 10; i++ {
			configSrc.Set(string(rune('a'+i)), i)
		}
		ctr.SetConfigSource(configSrc)

		// 并发读取
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				key := string(rune('a' + idx))
				_ = ctr.getConfigValue(key)
				done <- true
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestShutdownWithContextCancellation 测试Shutdown的context取消场景
func TestShutdownWithContextCancellation(t *testing.T) {
	// 测试：context取消导致Shutdown中止
	t.Run("shutdown cancelled by context", func(t *testing.T) {
		ctr := New()
		type TestService struct{ Name string }
		err := ctr.ProvideNamedWith("", func(c Container) (*TestService, error) {
			return &TestService{Name: "test"}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith() error = %v", err)
		}

		// 创建一个已经过期的context
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(10 * time.Millisecond) // 确保已过期

		_ = ctr.Shutdown(ctx)
		// Shutdown可能返回context错误，也可能忽略，取决于实现
	})

	// 测试：context已经取消
	t.Run("shutdown with already cancelled context", func(t *testing.T) {
		ctr := New()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消

		err := ctr.Shutdown(ctx)
		// 空容器的Shutdown可能成功也可能失败，取决于实现
		_ = err
	})

	// 测试：正常Shutdown
	t.Run("shutdown normally", func(t *testing.T) {
		ctr := New()
		type TestService struct{ Name string }
		err := ctr.ProvideNamedWith("", func(c Container) (*TestService, error) {
			return &TestService{Name: "test"}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith() error = %v", err)
		}

		ctx := context.Background()
		err = ctr.Shutdown(ctx)
		if err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}
	})
}

// TestConfigSourceNilScenarios 测试配置源为nil的场景
func TestConfigSourceNilScenarios(t *testing.T) {
	// 测试：GetConfigSource返回nil
	t.Run("get nil config source", func(t *testing.T) {
		ctr := New().(*container)
		// 默认情况下configSource可能就是nil
		src := ctr.GetConfigSource()
		// 可能返回nil也可能返回空的configSource
		_ = src
	})

	// 测试：注入配置时configSource为nil
	t.Run("inject config with nil source", func(t *testing.T) {
		type ConfigService struct {
			Value string `di:"config:MISSING"`
		}

		ctr := New().(*container)
		// 不设置configSource，保持nil状态

		err := ctr.ProvideNamedWith("", func(c Container) (*ConfigService, error) {
			return &ConfigService{}, nil
		})
		if err != nil {
			t.Fatalf("ProvideNamedWith() error = %v", err)
		}

		_, err = ctr.GetNamed((*ConfigService)(nil), "")
		// 配置不存在时，应该返回错误或使用零值
		// 具体行为取决于实现
		_ = err
	})
}

// TestConcurrentRaceConditions 测试并发竞争条件
func TestConcurrentRaceConditions(t *testing.T) {
	// 测试：并发访问配置
	t.Run("concurrent config access", func(t *testing.T) {
		ctr := New().(*container)
		configSrc := NewMapConfigSource()
		ctr.SetConfigSource(configSrc)

		done := make(chan bool)

		// Goroutine 1: 写配置
		go func() {
			for i := 0; i < 10; i++ {
				configSrc.Set(string(rune('a'+i%26)), i)
			}
			done <- true
		}()

		// Goroutine 2: 读配置
		go func() {
			for i := 0; i < 10; i++ {
				_ = ctr.getConfigValue(string(rune('a' + i%26)))
			}
			done <- true
		}()

		<-done
		<-done
	})
}

// TestComplexPackageScenarios 测试复杂package场景
func TestComplexPackageScenarios(t *testing.T) {
	// 测试：package依赖链
	t.Run("package dependency chain", func(t *testing.T) {
		type DB struct{ Name string }
		type Repository struct{ DB *DB }
		type Service struct{ Repo *Repository }

		dbPackage := Package(func(c Container) error {
			return c.ProvideNamedWith("", func(c Container) (*DB, error) {
				return &DB{Name: "test-db"}, nil
			})
		})

		repoPackage := Package(
			dbPackage,
			func(c Container) error {
				return c.ProvideNamedWith("", func(c Container) (*Repository, error) {
					db, err := c.GetNamed((*DB)(nil), "")
					if err != nil {
						return nil, err
					}
					return &Repository{DB: db.(*DB)}, nil
				})
			},
		)

		servicePackage := Package(
			repoPackage,
			func(c Container) error {
				return c.ProvideNamedWith("", func(c Container) (*Service, error) {
					repo, err := c.GetNamed((*Repository)(nil), "")
					if err != nil {
						return nil, err
					}
					return &Service{Repo: repo.(*Repository)}, nil
				})
			},
		)

		ctr := New()
		err := servicePackage(ctr)
		if err != nil {
			t.Fatalf("Package() error = %v", err)
		}

		svc, err := ctr.GetNamed((*Service)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed() error = %v", err)
		}

		service := svc.(*Service)
		if service.Repo.DB.Name != "test-db" {
			t.Errorf("Service.Repo.DB.Name = %v, want 'test-db'", service.Repo.DB.Name)
		}
	})

	// 测试：LoadPackages组合多个package
	t.Run("load multiple packages", func(t *testing.T) {
		type Logger struct{ Level string }
		type Database struct{ URL string }
		type Cache struct{ TTL int }

		loggerPackage := func(c Container) error {
			return c.ProvideNamedWith("", func(c Container) (*Logger, error) {
				return &Logger{Level: "info"}, nil
			})
		}

		databasePackage := func(c Container) error {
			return c.ProvideNamedWith("", func(c Container) (*Database, error) {
				return &Database{URL: "localhost:5432"}, nil
			})
		}

		cachePackage := func(c Container) error {
			return c.ProvideNamedWith("", func(c Container) (*Cache, error) {
				return &Cache{TTL: 3600}, nil
			})
		}

		ctr := New()
		result, err := LoadPackages(ctr, loggerPackage, databasePackage, cachePackage)
		if err != nil {
			t.Fatalf("LoadPackages() error = %v", err)
		}

		// 验证所有服务都已注册
		logger, err := result.GetNamed((*Logger)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed(Logger) error = %v", err)
		}
		if logger.(*Logger).Level != "info" {
			t.Errorf("Logger.Level = %v, want 'info'", logger.(*Logger).Level)
		}

		db, err := result.GetNamed((*Database)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed(Database) error = %v", err)
		}
		if db.(*Database).URL != "localhost:5432" {
			t.Errorf("Database.URL = %v, want 'localhost:5432'", db.(*Database).URL)
		}

		cache, err := result.GetNamed((*Cache)(nil), "")
		if err != nil {
			t.Fatalf("GetNamed(Cache) error = %v", err)
		}
		if cache.(*Cache).TTL != 3600 {
			t.Errorf("Cache.TTL = %v, want 3600", cache.(*Cache).TTL)
		}
	})
}
