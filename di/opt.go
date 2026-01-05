package di

// EntryInfo 包含服务实例和名称信息
type EntryInfo struct {
    Name     string
    Instance any
}

// ProviderOption 服务提供者选项函数
type ProviderOption func(*providerConfig)

// providerConfig 服务提供者配置
type providerConfig struct {
    loadMode      LoadMode                                  // 加载模式
    beforeCreate  []func(Container, EntryInfo) (any, error) //加载前调用,用于替换和阻止
    afterCreate   []func(Container, EntryInfo) (any, error) //加载后调用,用于替换和删除
    beforeDestroy []func(Container, EntryInfo)              //销毁前调用
    afterDestroy  []func(Container, EntryInfo)              //销毁后调用
}

// WithCondition 设置条件函数，只有条件满足时才调用 provider
func WithCondition(condition func(Container) bool) ProviderOption {
    return func(pc *providerConfig) {
        pc.beforeCreate = append(pc.beforeCreate, func(c Container, info EntryInfo) (any, error) {
            if condition(c) {
                return nil, nil
            }
            return nil, ErrorConditionFail
        })
    }
}

// WithLoadMode 设置加载模式
func WithLoadMode(mode LoadMode) ProviderOption {
    return func(pc *providerConfig) {
        pc.loadMode = mode
    }
}

// WithBeforeCreate 设置创建前钩子
func WithBeforeCreate(f func(Container, EntryInfo) (any, error)) ProviderOption {
    return func(pc *providerConfig) {
        pc.beforeCreate = append(pc.beforeCreate, f)
    }
}

// WithAfterCreate 设置创建后钩子
func WithAfterCreate(f func(Container, EntryInfo) (any, error)) ProviderOption {
    return func(pc *providerConfig) {
        pc.afterCreate = append(pc.afterCreate, f)
    }
}

// WithBeforeDestroy 设置销毁前钩子
func WithBeforeDestroy(f func(Container, EntryInfo)) ProviderOption {
    return func(pc *providerConfig) {
        pc.beforeDestroy = append(pc.beforeDestroy, f)
    }
}

// WithAfterDestroy 设置销毁后钩子
func WithAfterDestroy(f func(Container, EntryInfo)) ProviderOption {
    return func(pc *providerConfig) {
        pc.afterDestroy = append(pc.afterDestroy, f)
    }
}

// ContainerOption 容器选项函数
type ContainerOption func(*container)

// WithContainerBeforeCreate 设置容器级别的创建前钩子
func WithContainerBeforeCreate(f func(Container, EntryInfo) (any, error)) ContainerOption {
    return func(c *container) {
        c.beforeCreate = append(c.beforeCreate, f)
    }
}

// WithContainerAfterCreate 设置容器级别的创建后钩子
func WithContainerAfterCreate(f func(Container, EntryInfo) (any, error)) ContainerOption {
    return func(c *container) {
        c.afterCreate = append(c.afterCreate, f)
    }
}

// WithContainerBeforeDestroy 设置容器级别的销毁前钩子
func WithContainerBeforeDestroy(f func(Container, EntryInfo)) ContainerOption {
    return func(c *container) {
        c.beforeDestroy = append(c.beforeDestroy, f)
    }
}

// WithContainerAfterDestroy 设置容器级别的销毁后钩子
func WithContainerAfterDestroy(f func(Container, EntryInfo)) ContainerOption {
    return func(c *container) {
        c.afterDestroy = append(c.afterDestroy, f)
    }
}

// WithOnStart 设置启动前钩子（在Start方法调用时执行）
func WithOnStart(f func(Container) error) ContainerOption {
    return func(c *container) {
        c.onStartup = append(c.onStartup, f)
    }
}

// WithAfterStart 设置启动后钩子（在Start方法调用后执行）
func WithAfterStart(f func(Container) error) ContainerOption {
    return func(c *container) {
        c.afterStartup = append(c.afterStartup, f)
    }
}
