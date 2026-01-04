package di

import (
    "testing"
)

// TestSimple 基本功能测试
func TestSimple(t *testing.T) {
    container := New()

    // 测试1: 注册和获取结构体服务
    type TestService struct {
        Value string
    }

    err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
        return &TestService{Value: "hello"}, nil
    })
    if err != nil {
        t.Fatalf("注册服务失败: %v", err)
    }

    service, err := GetNamed[*TestService](container, "test")
    if err != nil {
        t.Fatalf("获取服务失败: %v", err)
    }
    if service.Value != "hello" {
        t.Errorf("期望'hello', 实际'%s'", service.Value)
    }

    // 测试2: 配置源
    source := NewMapConfigSource()
    source.Set("key1", "value1")
    source.Set("key2", 123)
    container.SetConfigSource(source)

    val1 := container.Value("key1")
    if val1.String() != "value1" {
        t.Errorf("配置获取失败: 期望'value1', 实际'%s'", val1.String())
    }

    val2 := container.Value("key2")
    if val2.Any() != 123 {
        t.Errorf("配置获取失败: 期望123, 实际%v", val2.Any())
    }

    // 测试3: 配置注入
    type ConfigService struct {
        Name string `di.config:"key1"`
        Port int    `di.config:"key2"`
    }

    err = container.ProvideNamedWith("config", func(c Container) (*ConfigService, error) {
        return &ConfigService{}, nil
    })
    if err != nil {
        t.Fatalf("注册配置服务失败: %v", err)
    }

    configService, err := GetNamed[*ConfigService](container, "config")
    if err != nil {
        t.Fatalf("获取配置服务失败: %v", err)
    }

    if configService.Name != "value1" {
        t.Errorf("配置注入失败: 期望'value1', 实际'%s'", configService.Name)
    }
    if configService.Port != 123 {
        t.Errorf("配置注入失败: 期望123, 实际%d", configService.Port)
    }

    t.Log("所有基础测试通过!")
}
