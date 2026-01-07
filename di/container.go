package di

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/signal"
    "reflect"
    "strings"
    "sync"
    "syscall"
    "time"

    "github.com/mzzsfy/go-util/config"
    "github.com/mzzsfy/go-util/helper"
)

// container 实现 Container 接口
// 懒加载实现说明：
// 1. 懒加载服务在注册时不会被创建
// 2. 只有在第一次调用 GetNamed 时才会创建实例
// 3. 在创建过程中会检测循环依赖
// 4. 使用 loading map 跟踪正在创建的实例，防止循环依赖
type container struct {
    beforeCreate  []func(Container, EntryInfo) (any, error) //加载前调用,用于替换和阻止
    afterCreate   []func(Container, EntryInfo) (any, error) //加载后调用,用于替换和删除
    beforeDestroy []func(Container, EntryInfo)              //销毁前调用
    afterDestroy  []func(Container, EntryInfo)              //销毁后调用
    providers     map[string]providerEntry
    instances     map[string]any
    mu            sync.RWMutex
    parent        *container
    shutdown      []ShutdownHook

    // 懒加载相关字段
    loading map[string]bool // 正在懒加载的实例，用于检测循环依赖

    // 配置注入相关字段
    configSource ConfigSource // 配置源
    configMu     sync.RWMutex

    // 性能监控字段
    stats   containerStats // 统计信息
    statsMu sync.RWMutex

    // 启动相关字段
    done         chan struct{}
    onStartup    []func(Container) error // 启动前钩子
    afterStartup []func(Container) error // 启动后钩子
    started      bool                    // 是否已启动
}

type containerStats struct {
    createdInstances int           // 创建的实例总数
    getCalls         int           // Get调用次数
    provideCalls     int           // Provide调用次数
    configHits       int           // 配置命中次数
    configMisses     int           // 配置未命中次数
    createDuration   time.Duration // 总创建耗时
}

type providerEntry struct {
    reflectType reflect.Type
    provider    func(Container) (any, error)
    config      providerConfig
}

// New 创建新的 DI 容器
func New(opts ...ContainerOption) Container {
    c := &container{
        providers:    make(map[string]providerEntry),
        instances:    make(map[string]any),
        loading:      make(map[string]bool),
        configSource: NewMapConfigSource(),
        done:         make(chan struct{}),
        onStartup:    make([]func(Container) error, 0),
        afterStartup: make([]func(Container) error, 0),
    }

    for _, opt := range opts {
        opt(c)
    }

    return c
}

var blackTypeList = []reflect.Type{
    reflect.TypeOf((*context.Context)(nil)).Elem(),
    reflect.TypeOf((*string)(nil)).Elem(),
    reflect.TypeOf((*int)(nil)).Elem(),
}

func (c *container) ProvideNamedWith(name string, provider any, opts ...ProviderOption) error {
    // 验证 provider 是正确的函数类型
    providerType := reflect.TypeOf(provider)
    if providerType.Kind() != reflect.Func {
        return fmt.Errorf("provider must be a function")
    }
    if providerType.NumIn() != 1 || providerType.In(0) != reflect.TypeOf((*Container)(nil)).Elem() {
        return fmt.Errorf("provider must have signature func(Container) (T, error)")
    }
    if providerType.NumOut() != 2 || providerType.Out(1).String() != "error" {
        return fmt.Errorf("provider must return (T, error)")
    }

    // 获取返回类型
    returnType := providerType.Out(0)
    key := typeKey(returnType, name)

    if name == "" {
        // 检查是否在黑名单中（通过字符串比较，因为reflect.Type在Go1.18不可比较）
        returnTypeStr := returnType.String()
        for _, blackType := range blackTypeList {
            if blackType.String() == returnTypeStr {
                // 不允许不使用名称注册这些类型
                return fmt.Errorf("cannot register type %s without name", returnType)
            }
        }
    }
    c.mu.Lock()

    if _, exists := c.providers[key]; exists {
        c.mu.Unlock()
        return fmt.Errorf("provider for type %s with name '%s' already registered", returnType, name)
    }

    // 应用选项，默认为 LoadModeDefault
    p := providerConfig{}
    for _, opt := range opts {
        opt(&p)
    }

    c.providers[key] = providerEntry{
        reflectType: returnType,
        provider: func(cont Container) (any, error) {
            //使用反射调用,保证泛型兼容
            results := reflect.ValueOf(provider).Call([]reflect.Value{reflect.ValueOf(cont)})
            if !results[1].IsNil() {
                return nil, results[1].Interface().(error)
            }
            return results[0].Interface(), nil
        },
        config: p,
    }

    // 更新统计
    c.statsMu.Lock()
    c.stats.provideCalls++
    c.statsMu.Unlock()

    // 解锁以便立即加载时可以正常工作
    c.mu.Unlock()

    // 如果是立即加载模式，额外调用一次,保证创建实例
    if p.loadMode == LoadModeImmediate {
        _, err := c.GetNamed(returnType, name)
        if err != nil {
            return err
        }
    }
    return nil
}

func (c *container) GetNamed(serviceType any, name string) (any, error) {
    // 获取类型信息
    var t reflect.Type
    switch v := serviceType.(type) {
    case reflect.Type:
        t = v
    default:
        t = reflect.TypeOf(serviceType)
    }

    key := typeKey(t, name)

    // 检查缓存实例
    c.mu.RLock()
    if instance, exists := c.instances[key]; exists {
        c.mu.RUnlock()
        return instance, nil
    }
    c.mu.RUnlock()

    // 检查提供者
    c.mu.RLock()
    entry, exists := c.providers[key]
    c.mu.RUnlock()

    if !exists {
        // 检查父容器
        if c.parent != nil {
            return c.parent.GetNamed(t, name)
        }
        return nil, fmt.Errorf("no provider found for type %s with name '%s'", t, name)
    }
    return c.create(entry, name)
}

func (c *container) create(entry providerEntry, name string) (any, error) {
    // 性能统计开始时间
    startTime := time.Now()

    key := typeKey(entry.reflectType, name)
    var instance any

    // 创建 EntryInfo 用于钩子调用（此时还没有实例）
    info := EntryInfo{
        Instance: nil,
        Name:     name,
    }

    // 检查条件函数 (provider级别的 beforeCreate)
    for _, f := range entry.config.beforeCreate {
        if v, err := f(c, info); err != nil {
            return nil, fmt.Errorf("not create %s with name '%s':%w", entry.reflectType, name, err)
        } else if v != nil {
            instance = v
            info.Instance = v
        }
    }
    entry.config.beforeCreate = nil

    // 执行容器级别的 beforeCreate 钩子
    for _, f := range c.beforeCreate {
        if v, err := f(c, info); err != nil {
            return nil, fmt.Errorf("container beforeCreate failed for %s with name '%s':%w", entry.reflectType, name, err)
        } else if v != nil {
            instance = v
        }
    }
    // 使用双重检查锁定模式避免并发问题
    c.mu.RLock()
    if instance1, exists1 := c.instances[key]; exists1 {
        c.mu.RUnlock()
        // 更新统计
        c.statsMu.Lock()
        c.stats.getCalls++
        c.statsMu.Unlock()
        return instance1, nil
    }
    c.mu.RUnlock()

    // 获取写锁，准备创建实例
    c.mu.Lock()
    // 再次检查，防止其他 goroutine 已经创建了
    if instance1, exists1 := c.instances[key]; exists1 {
        c.mu.Unlock()
        // 更新统计
        c.statsMu.Lock()
        c.stats.getCalls++
        c.statsMu.Unlock()
        return instance1, nil
    }
    // 标记正在创建
    if c.loading[key] {
        c.mu.Unlock()
        return nil, fmt.Errorf("circular dependency detected for type %s with name '%s'", entry.reflectType, name)
    }
    // 创建完成后清理标记
    defer func(k string) {
        c.mu.Lock()
        delete(c.loading, k)
        c.mu.Unlock()
    }(key)
    c.loading[key] = true
    c.mu.Unlock()
    if instance == nil {
        if entry.config.loadMode == LoadModeLazy {
            depend, err := c.findDepend(entry.reflectType)
            if err != nil {
                // Clean up loading state before returning error
                c.mu.Lock()
                delete(c.loading, key)
                c.mu.Unlock()
                return nil, err
            }
            for _, k := range depend {
                name1 := ""
                ss := strings.Split(k, "#")
                if len(ss) > 1 {
                    name1 = ss[1]
                }
                // 提前创建依赖实例,减少创建实例流程
                _, err := c.create(c.providers[k], name1)
                if err != nil {
                    return nil, err
                }
            }
        }

        providerFunc := entry.provider
        // 创建实例
        instance1, err := providerFunc(c)
        instance = instance1
        if err != nil {
            return nil, err
        }
    }

    // 重新获取锁以进行配置注入和缓存
    c.mu.Lock()
    // 检查在此期间是否已经有其他 goroutine设置了实例
    if existingInstance, exists := c.instances[key]; exists && entry.config.loadMode != LoadModeTransient {
        // 如果其他goroutine已经创建了实例，使用已存在的实例
        c.mu.Unlock()
        // 更新统计
        c.statsMu.Lock()
        c.stats.getCalls++
        c.statsMu.Unlock()
        return existingInstance, nil
    }

    c.mu.Unlock()
    instance, err := c.createInstance(entry, name, instance, startTime)
    if err != nil {
        return instance, err
    }

    return instance, nil
}

func (c *container) createInstance(entry providerEntry, name string, instance any, startTime time.Time) (any, error) {
    // 如果实例为nil，跳过依赖注入
    if instance == nil {
        return nil, nil
    }

    // 尝试注入配置和服务到实例
    instanceValue := reflect.ValueOf(instance)

    // 如果实例是无效值（如nil），跳过注入
    if !instanceValue.IsValid() {
        return instance, nil
    }

    // 检查容器是否有配置数据（支持配置注入）
    c.configMu.RLock()
    c.configMu.RUnlock()

    // 如果需要注入，处理注入逻辑
    if instanceValue.Kind() == reflect.Struct && instanceValue.CanAddr() == false {
        // 如果实例不可寻址，创建一个可寻址的副本
        addr := reflect.New(instanceValue.Type())
        addr.Elem().Set(instanceValue)
        if err := c.injectStruct(addr); err != nil {
            return nil, fmt.Errorf("failed to inject to instance: %w", err)
        }
        // 更新实例为注入后的值
        instance = addr.Elem().Interface()
    } else if instanceValue.Kind() == reflect.Struct || instanceValue.Kind() == reflect.Ptr {
        // 只有结构体或指针才尝试注入
        if err := c.injectStruct(instanceValue); err != nil {
            return nil, fmt.Errorf("failed to inject to instance: %w", err)
        }
    }

    c.mu.Lock()
    // 非 Transient 模式缓存实例
    if entry.config.loadMode != LoadModeTransient {
        key := typeKey(entry.reflectType, name)
        c.instances[key] = instance
    }
    // 注册销毁钩子（同时支持容器级别和 provider 级别）
    beforeDestroy := entry.config.beforeDestroy
    afterDestroy := entry.config.afterDestroy
    hasDestroyHooks := len(entry.config.beforeDestroy) > 0 || len(afterDestroy) > 0 ||
        len(c.beforeDestroy) > 0 || len(c.afterDestroy) > 0

    if hasDestroyHooks {
        // 创建一个包装函数来执行销毁钩子
        destroyHook := func(ctx context.Context) error {
            containerContext := ContainerContext{parent: ctx}
            //todo: 提取error
            destroyInfo := EntryInfo{
                Instance: instance,
                Name:     name,
                Ctx:      containerContext,
            }
            // 执行容器级别的 beforeDestroy 钩子
            for _, f := range c.beforeDestroy {
                f(c, destroyInfo)
            }
            // 执行 provider级别的 beforeDestroy 钩子
            for _, f := range beforeDestroy {
                f(c, destroyInfo)
            }
            // 执行实例的 Destroy 钩子
            if lifecycle, ok := instance.(DestroyCallback); ok {
                err := lifecycle.OnDestroyCallback()
                if err != nil {
                    return fmt.Errorf("DestroyCallback failed for %s with name '%s':%w", entry.reflectType, name, err)
                }
            }
            if lifecycle, ok := instance.(ServiceLifecycle); ok {
                err := lifecycle.Shutdown(ctx)
                if err != nil {
                    return fmt.Errorf("shutdown failed for %s with name '%s':%w", entry.reflectType, name, err)
                }
            }
            // 执行 provider级别的 afterDestroy 钩子
            for _, f := range afterDestroy {
                f(c, destroyInfo)
            }
            // 执行容器级别的 afterDestroy 钩子
            for _, f := range c.afterDestroy {
                f(c, destroyInfo)
            }
            return nil
        }
        c.shutdown = append(c.shutdown, destroyHook)
        entry.config.beforeDestroy = nil
        entry.config.afterDestroy = nil
    }

    c.mu.Unlock()

    // 更新性能统计
    c.statsMu.Lock()
    c.stats.createdInstances++
    c.stats.getCalls++
    c.stats.createDuration += time.Since(startTime)
    c.statsMu.Unlock()

    // 执行 provider级别的 afterCreate 钩子
    for _, f := range entry.config.afterCreate {
        if v, err := f(c, EntryInfo{
            Name:     name,
            Instance: instance,
        }); err != nil {
            return nil, fmt.Errorf("afterCreate failed for %s with name '%s':%w", entry.reflectType, name, err)
        } else if v != nil {
            instance = v
        }
    }

    // 执行容器级别的 afterCreate 钩子
    for _, f := range c.afterCreate {
        if v, err := f(c, EntryInfo{
            Name:     name,
            Instance: instance,
        }); err != nil {
            return nil, fmt.Errorf("container afterCreate failed for %s with name '%s':%w", entry.reflectType, name, err)
        } else if v != nil {
            instance = v
        }
    }
    return instance, nil
}

func (c *container) findDepend(t reflect.Type) ([]string, error) {
    //找到所有待注入的字段`di:xxx`
    var fields []string
    switch t.Kind() {
    case reflect.Pointer:
        return c.findDepend(t.Elem())
    case reflect.Struct:
        for i := 0; i < t.NumField(); i++ {
            field := t.Field(i)
            tag, hasTag := field.Tag.Lookup("di")
            if !hasTag {
                continue // 没有di标签，跳过
            }

            // 解析标签值作为服务名称
            name := tag

            // 获取字段的实际类型（与injectStruct保持一致）
            serviceType := field.Type

            typeName := typeKey(serviceType, name)
            if _, ok := c.providers[typeName]; ok {
                fields = append(fields, typeName)
            } else {
                return nil, fmt.Errorf("no provider found for type %s with name '%s'", serviceType, name)
            }
        }
    default:
        return nil, fmt.Errorf("provider must be a struct")
    }
    return fields, nil
}

func (c *container) HasNamed(serviceType any, name string) bool {
    // 获取类型信息
    var t reflect.Type
    switch v := serviceType.(type) {
    case reflect.Type:
        t = v
    default:
        t = reflect.TypeOf(serviceType)
        if t.Kind() == reflect.Ptr {
            t = t.Elem()
        }
    }

    key := typeKey(t, name)

    c.mu.RLock()
    defer c.mu.RUnlock()

    if _, exists := c.providers[key]; exists {
        return true
    }
    // 检查父容器
    if c.parent != nil {
        return c.parent.HasNamed(serviceType, name)
    }
    return false
}

func (c *container) AppendOption(opt ...ContainerOption) (err error) {
    func(c *container) {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("panic while appending options: %v", r)
            }
        }()
        for _, option := range opt {
            option(c)
        }
    }(c)
    return
}

// Start 启动容器
func (c *container) Start() error {
    // 使用双重检查锁定模式
    c.mu.RLock()
    if c.started {
        c.mu.RUnlock()
        return helper.StringError("container is already started")
    }
    c.mu.RUnlock()

    // 获取写锁进行启动操作
    c.mu.Lock()
    if c.started {
        c.mu.Unlock()
        return helper.StringError("container is already started")
    }

    // 检查容器是否已被关闭（done通道已关闭）
    select {
    case <-c.done:
        // 容器已被关闭，重新创建done通道以支持重启
        c.done = make(chan struct{})
    default:
        // 容器未关闭，继续正常启动流程
    }

    // 标记为正在启动，防止其他goroutine同时执行钩子
    c.started = true
    c.mu.Unlock()

    // 执行启动前钩子（此时容器标记为已启动，但钩子可能失败）
    for i, hook := range c.onStartup {
        if err := hook(c); err != nil {
            // 启动失败，重置状态
            c.mu.Lock()
            c.started = false
            c.mu.Unlock()
            return fmt.Errorf("startup hook %d failed: %w", i, err)
        }
    }

    // 执行启动后钩子
    for i, hook := range c.afterStartup {
        if err := hook(c); err != nil {
            // 启动失败，重置状态
            c.mu.Lock()
            c.started = false
            c.mu.Unlock()
            return fmt.Errorf("after startup hook %d failed: %w", i, err)
        }
    }
    c.onStartup = nil
    c.afterStartup = nil
    return nil
}

func (c *container) Shutdown(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // 检查是否已经关闭
    select {
    case <-c.Done():
        return helper.StringError("container is already shutting down")
    default:
    }
    var err error

    for _, hook := range c.shutdown {
        if hookErr := hook(ctx); hookErr != nil {
            err = fmt.Errorf("shutdown failed: %w", hookErr)
        }
    }

    // 清理资源
    c.providers = make(map[string]providerEntry)
    c.instances = make(map[string]any)
    c.shutdown = nil
    c.started = false // 重置启动状态，允许重启

    // 清理配置数据
    c.configMu.Lock()
    c.configSource = NewMapConfigSource() // 重置为新的空配置源
    c.configMu.Unlock()

    // 重置统计
    c.statsMu.Lock()
    c.stats = containerStats{}
    c.statsMu.Unlock()
    close(c.done)
    return err
}

func (c *container) ShutdownOnSignals(signals ...os.Signal) {
    if len(signals) == 0 {
        signals = []os.Signal{syscall.SIGTERM, os.Interrupt}
    }

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, signals...)

    go func() {
        sig := <-sigChan
        fmt.Printf("Received signal %s, shutting down...\n", sig)
        if err := c.Shutdown(context.Background()); err != nil {
            fmt.Printf("Shutdown error: %v\n", err)
        }
        os.Exit(0)
    }()
}
func (c *container) Done() <-chan struct{} {
    return c.done
}

func (c *container) CreateChildScope() Container {
    // 子容器继承父容器的配置源（共享同一个配置源）
    c.configMu.RLock()
    inheritedConfigSource := c.configSource
    c.configMu.RUnlock()

    c2 := &container{
        providers:    make(map[string]providerEntry),
        instances:    make(map[string]any),
        loading:      make(map[string]bool),
        parent:       c,
        done:         make(chan struct{}),
        configSource: inheritedConfigSource,
    }
    //关闭子容器
    c.shutdown = append(c.shutdown, func(ctx context.Context) error { return c2.Shutdown(ctx) })
    return c2
}

func typeKey(t reflect.Type, name string) string {
    typeName := t.String()
    if name != "" {
        return typeName + "#" + name
    }
    return typeName
}

// injectStruct 注入配置和服务到结构体字段
func (c *container) injectStruct(target reflect.Value) error {
    // 检查是否为nil指针
    if target.Kind() == reflect.Ptr {
        if target.IsNil() {
            // nil指针，无法注入，但也不是错误
            return nil
        }
        target = target.Elem()
    }
    if target.Kind() != reflect.Struct {
        return fmt.Errorf("provider must be a struct")
    }

    targetType := target.Type()

    for i := 0; i < target.NumField(); i++ {
        fieldType := targetType.Field(i)
        fieldValue := target.Field(i)

        // 检查di标签 - 服务注入
        diTag, hasDiTag := fieldType.Tag.Lookup("di")
        if hasDiTag {
            // 服务注入：原有逻辑
            serviceName := diTag

            // 如果标签值为空，使用默认名称（空字符串）
            if diTag == "" {
                serviceName = ""
            }

            // 获取字段的类型
            fieldValueType := fieldType.Type

            // 确定要从容器中请求的服务类型
            var serviceType reflect.Type
            if fieldValueType.Kind() == reflect.Interface {
                // 字段是接口类型，直接使用接口类型
                serviceType = fieldValueType
            } else if fieldValueType.Kind() == reflect.Ptr {
                // 字段是指针类型，请求指针类型的服务
                serviceType = fieldValueType
            } else {
                // 字段是值类型，请求值类型的服务
                serviceType = fieldValueType
            }

            // 从容器中获取服务实例
            serviceInstance, err := c.GetNamed(serviceType, serviceName)
            if err != nil {
                return fmt.Errorf("failed to inject field %s: %w", fieldType.Name, err)
            }

            // 设置字段值
            if err := setFieldValue(fieldValue, serviceInstance); err != nil {
                return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
            }
            continue
        }

        // 检查di.config标签 - 配置注入
        configTag, hasConfigTag := fieldType.Tag.Lookup("di.config")
        if hasConfigTag {
            var actualValue any

            // 检查是否包含变量替换语法 ${}
            if strings.Contains(configTag, "${") {
                // 使用变量替换解析
                resolvedValue := c.resolveConfigValue(configTag)
                actualValue = resolvedValue
            } else {
                // 传统格式：提取配置key和可能的默认值
                configKey, defaultValue := parseConfigInjection(configTag)
                actualValue = c.getConfigValue(configKey).AnyD(defaultValue)
            }

            // 设置字段值
            if err := setFieldValue(fieldValue, actualValue); err != nil {
                return fmt.Errorf("failed to inject config to field %s: %w", fieldType.Name, err)
            }
        }
    }

    return nil
}

// parseConfigInjection 解析配置注入标签
// 支持格式：
// - "keyName" - 只有key，无默认值
// - "keyName:defaultValue" - key和默认值
// - "prefix${key:default}suffix" - 变量替换，支持多个变量
func parseConfigInjection(tag string) (key string, defaultValue string) {
    // 检查是否包含${}语法
    if !strings.Contains(tag, "${") {
        // 传统格式：keyName 或 keyName:defaultValue
        colonIdx := strings.Index(tag, ":")
        if colonIdx == -1 {
            return tag, ""
        }
        return tag[:colonIdx], tag[colonIdx+1:]
    }

    // 处理${key:default}格式
    // 提取第一个${key:default}块
    start := strings.Index(tag, "${")
    end := strings.Index(tag[start:], "}")
    if start == -1 || end == -1 {
        // 格式错误，返回原始值
        return tag, ""
    }

    // 计算结束位置
    end += start

    // 提取变量部分：${key:default}
    varPart := tag[start+2 : end]

    // 解析变量中的key和default
    colonIdx := strings.Index(varPart, ":")
    if colonIdx == -1 {
        key = varPart
        defaultValue = ""
    } else {
        key = varPart[:colonIdx]
        defaultValue = varPart[colonIdx+1:]
    }

    return key, defaultValue
}

// resolveConfigValue 解析配置值，支持变量替换
// 支持格式：
// - "simpleKey" -> 直接获取配置值
// - "prefix${key:default}suffix" -> 替换变量后拼接
func (c *container) resolveConfigValue(tag string) string {
    // 检查是否包含变量
    if !strings.Contains(tag, "${") {
        // 没有变量，直接作为key获取配置值
        value := c.getConfigValue(tag)
        return value.StringD("")
    }

    // 处理变量替换
    result := ""
    remaining := tag

    for {
        start := strings.Index(remaining, "${")
        if start == -1 {
            // 没有更多变量，添加剩余部分
            result += remaining
            break
        }

        // 添加变量前的固定文本
        result += remaining[:start]
        remaining = remaining[start:]

        end := strings.Index(remaining, "}")
        if end == -1 {
            // 格式错误，直接添加剩余部分
            result += remaining
            break
        }

        // 提取变量部分：${key:default}
        varPart := remaining[2:end]

        // 解析key和default
        var key, defaultValue string
        colonIdx := strings.Index(varPart, ":")
        if colonIdx == -1 {
            key = varPart
            defaultValue = ""
        } else {
            key = varPart[:colonIdx]
            defaultValue = varPart[colonIdx+1:]
        }

        // 获取配置值
        value := c.getConfigValue(key)
        resolvedValue := value.StringD(defaultValue)

        // 添加解析后的值
        result += resolvedValue

        // 移动到下一个位置
        remaining = remaining[end+1:]
    }

    return result
}

// getConfigValue 获取配置值
func (c *container) getConfigValue(key string) config.Value {
    c.configMu.RLock()
    defer c.configMu.RUnlock()

    if c.configSource == nil {
        // 更新统计
        c.statsMu.Lock()
        c.stats.configMisses++
        c.statsMu.Unlock()
        return config.ValueFrom(nil)
    }

    if key == "" {
        // 对于空key，返回所有配置（需要额外的方法）
        // 这里暂时返回nil，后续可以扩展
        return config.ValueFrom(nil)
    }

    value := c.configSource.Get(key)
    if value != nil {
        // 更新统计
        c.statsMu.Lock()
        c.stats.configHits++
        c.statsMu.Unlock()
        return value
    }

    // 更新统计
    c.statsMu.Lock()
    c.stats.configMisses++
    c.statsMu.Unlock()

    return config.ValueFrom(nil)
}

// SetConfigSource 设置配置源
func (c *container) SetConfigSource(source ConfigSource) {
    c.configMu.Lock()
    defer c.configMu.Unlock()
    if source == nil {
        panic("ConfigSource cannot be nil")
    }
    c.configSource = source
}

// GetConfigSource 获取当前配置源
func (c *container) GetConfigSource() ConfigSource {
    c.configMu.RLock()
    defer c.configMu.RUnlock()
    return c.configSource
}

// Value 获取配置值（实现Container接口）
func (c *container) Value(key string) config.Value {
    return c.getConfigValue(key)
}

// ========== 性能监控和统计 ==========

// GetStats 获取容器统计信息
func (c *container) GetStats() ContainerStats {
    c.statsMu.RLock()
    defer c.statsMu.RUnlock()

    return ContainerStats{
        CreatedInstances: c.stats.createdInstances,
        GetCalls:         c.stats.getCalls,
        ProvideCalls:     c.stats.provideCalls,
        ConfigHits:       c.stats.configHits,
        ConfigMisses:     c.stats.configMisses,
        CreateDuration:   c.stats.createDuration,
    }
}

// ResetStats 重置统计信息
func (c *container) ResetStats() {
    c.statsMu.Lock()
    defer c.statsMu.Unlock()

    c.stats = containerStats{}
}

// GetInstanceCount 获取当前缓存的实例数量
func (c *container) GetInstanceCount() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.instances)
}

// GetProviderCount 获取注册的提供者数量
func (c *container) GetProviderCount() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.providers)
}

// GetAverageCreateDuration 获取平均创建耗时
func (c *container) GetAverageCreateDuration() time.Duration {
    c.statsMu.RLock()
    defer c.statsMu.RUnlock()

    if c.stats.createdInstances == 0 {
        return 0
    }
    return c.stats.createDuration / time.Duration(c.stats.createdInstances)
}

// ========== 实例替换功能 ==========

// ReplaceInstance 运行时替换已注册的服务实例
func (c *container) ReplaceInstance(serviceType any, name string, newInstance any) error {
    var t reflect.Type
    switch v := serviceType.(type) {
    case reflect.Type:
        t = v
    default:
        t = reflect.TypeOf(serviceType)
    }

    key := typeKey(t, name)

    c.mu.Lock()
    defer c.mu.Unlock()

    // 检查是否已存在该类型的提供者
    if _, exists := c.providers[key]; !exists {
        return fmt.Errorf("no provider found for type %s with name '%s'", t, name)
    }

    // 替换实例
    c.instances[key] = newInstance
    return nil
}

// RemoveInstance 移除已缓存的实例
func (c *container) RemoveInstance(serviceType any, name string) error {
    var t reflect.Type
    switch v := serviceType.(type) {
    case reflect.Type:
        t = v
    default:
        t = reflect.TypeOf(serviceType)
    }

    key := typeKey(t, name)

    c.mu.Lock()
    defer c.mu.Unlock()

    delete(c.instances, key)
    return nil
}

// ClearInstances 清空所有缓存的实例
func (c *container) ClearInstances() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.instances = make(map[string]any)
}

// ========== 多实例管理增强 ==========

// GetNamedAll 获取指定类型的所有命名实例
func (c *container) GetNamedAll(serviceType any) (map[string]any, error) {
    var t reflect.Type
    switch v := serviceType.(type) {
    case reflect.Type:
        t = v
    default:
        t = reflect.TypeOf(serviceType)
    }
    // 检查是否在黑名单中)
    returnTypeStr := t.String()
    for _, blackType := range blackTypeList {
        if blackType.String() == returnTypeStr {
            return nil, fmt.Errorf("cannot use GetNamedAll for type %s", t)
        }
    }
    name := t.String()
    results := make(map[string]any)

    // 查找所有匹配类型的提供者
    for key, entry := range c.providers {
        //使用反射依次检查所有符合的类型
        if entry.reflectType.AssignableTo(t) {
            var serviceName string
            if key == name {
                serviceName = "" // 默认名称
            } else {
                // 从 "type#name" 中提取 name
                _, serviceName, _ = strings.Cut(key, "#")
            }

            instance, err := c.GetNamed(entry.reflectType, serviceName)
            if err != nil {
                // 忽略条件失败
                if errors.Is(err, ErrorConditionFail) {
                    continue
                } else {
                    return nil, err
                }
            }
            results[key] = instance
        }
    }

    // 检查父容器
    if c.parent != nil {
        parentResults, err := c.parent.GetNamedAll(serviceType)
        if err != nil {
            return nil, err
        }
        for k, v := range parentResults {
            results[k] = v
        }
    }

    return results, nil
}

// GetAllInstances 获取所有已缓存的实例（包括不同名称）
func (c *container) GetAllInstances() map[string]any {
    c.mu.RLock()
    defer c.mu.RUnlock()

    // 返回副本
    results := make(map[string]any, len(c.instances))
    for k, v := range c.instances {
        results[k] = v
    }
    return results
}

// GetProviders 获取所有注册的提供者信息
func (c *container) GetProviders() map[string]string {
    c.mu.RLock()
    defer c.mu.RUnlock()

    results := make(map[string]string, len(c.providers))
    for k, entry := range c.providers {
        results[k] = entry.reflectType.String()
    }
    return results
}

// setFieldValue 设置字段值，处理类型兼容性
func setFieldValue(field reflect.Value, value any) error {
    if !field.CanSet() {
        return fmt.Errorf("field cannot be set")
    }

    valueReflect := reflect.ValueOf(value)
    fieldType := field.Type()
    valueType := valueReflect.Type()

    // 直接类型匹配
    if fieldType == valueType {
        field.Set(valueReflect)
        return nil
    }

    // 处理接口类型（优先处理，因为接口可以接受指针或值）
    if fieldType.Kind() == reflect.Interface {
        if valueReflect.Type().Implements(fieldType) {
            field.Set(valueReflect)
            return nil
        }
        // 如果值是指针，检查指针类型是否实现接口
        if valueType.Kind() == reflect.Ptr && valueReflect.Elem().Type().Implements(fieldType) {
            field.Set(valueReflect)
            return nil
        }
    }

    // 处理指针和值的转换
    if fieldType.Kind() == reflect.Ptr && valueType.Kind() != reflect.Ptr {
        // 需要取地址
        if valueReflect.CanAddr() {
            field.Set(valueReflect.Addr())
        } else {
            // 创建新的可寻址值
            ptr := reflect.New(valueType)
            ptr.Elem().Set(valueReflect)
            field.Set(ptr)
        }
        return nil
    }

    if fieldType.Kind() != reflect.Ptr && valueType.Kind() == reflect.Ptr {
        // 需要解引用
        field.Set(valueReflect.Elem())
        return nil
    }

    // 尝试类型转换
    if valueReflect.Type().ConvertibleTo(fieldType) {
        field.Set(valueReflect.Convert(fieldType))
        return nil
    }

    // 智能自动类型转换（当标准转换不可用时）
    if err := smartTypeConversion(field, valueReflect, fieldType, valueType); err == nil {
        return nil
    }

    return fmt.Errorf("cannot convert %s to %s", valueType, fieldType)
}

// smartTypeConversion 智能类型转换，支持字符串到基本类型的自动转换
func smartTypeConversion(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type, valueType reflect.Type) error {
    // 只处理字符串到基本类型的转换
    if valueType.Kind() != reflect.String {
        return fmt.Errorf("smart conversion only supports string source")
    }

    strValue := valueReflect.String()

    switch fieldType.Kind() {
    case reflect.Bool:
        // 字符串转布尔值
        boolValue := strValue == "true" || strValue == "yes" || strValue == "1" || strValue == "on"
        field.SetBool(boolValue)
        return nil

    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        // 字符串转整数
        var intValue int64
        _, err := fmt.Sscanf(strValue, "%d", &intValue)
        if err != nil {
            return fmt.Errorf("cannot convert string '%s' to int: %w", strValue, err)
        }
        field.SetInt(intValue)
        return nil

    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        // 字符串转无符号整数
        var uintValue uint64
        _, err := fmt.Sscanf(strValue, "%d", &uintValue)
        if err != nil {
            return fmt.Errorf("cannot convert string '%s' to uint: %w", strValue, err)
        }
        field.SetUint(uintValue)
        return nil

    case reflect.Float32, reflect.Float64:
        // 字符串转浮点数
        var floatValue float64
        _, err := fmt.Sscanf(strValue, "%f", &floatValue)
        if err != nil {
            return fmt.Errorf("cannot convert string '%s' to float: %w", strValue, err)
        }
        field.SetFloat(floatValue)
        return nil

    case reflect.String:
        // 字符串转字符串（直接设置）
        field.SetString(strValue)
        return nil
    }

    return fmt.Errorf("unsupported smart conversion from string to %s", fieldType.Kind())
}
