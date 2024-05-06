package helper

import "context"

type StatusKey[T comparable] interface {
    InitValue() T
}

// StatusValue 在程序运行中,方便的更新状态值 DefStatusItem(ctx, k1).Set(11)
type StatusValue[T comparable] interface {
    Value() T
    Set(v T) StatusValue[T]
}

type statusValue[T comparable] struct {
    v T
}

func (t *statusValue[T]) Value() T {
    return t.v
}

func (t *statusValue[T]) Set(v T) StatusValue[T] {
    t.v = v
    return t
}

type statusKey[T comparable] struct {
    f func() T
}

func (t statusKey[T]) InitValue() T {
    return t.f()
}

func DefStatusKeyFn[T comparable](f func() T) StatusKey[T] {
    return &statusKey[T]{f}
}

func DefStatusKeyStatic[T comparable](t T) StatusKey[T] {
    return &statusKey[T]{func() T { return t }}
}

type contextHolder struct {
    keys []any
    ctx  context.Context
}

func NewStatusTraceCtx() StatusHolder {
    return &contextHolder{ctx: context.Background()}
}

func (t *contextHolder) Get(key any) any {
    return t.ctx.Value(key)
}

func (t *contextHolder) Set(key, value any) {
    if value == nil {
        panic("status trace value 不能为 nil")
    }
    a := t.ctx.Value(key)
    if a != nil {
        t.keys = append(t.keys, key)
    }
    t.ctx = context.WithValue(t.ctx, key, value)
}

func (t *contextHolder) Each(f func(key, value any)) {
    for _, k := range t.keys {
        f(k, t.ctx.Value(k))
    }
}

// StatusHolder 状态保存器,可以用context或者gls自定义实现
type StatusHolder interface {
    Get(any) any
    Set(any, any)
}

type EachStatusHolder interface {
    StatusHolder
    Each(func(key any, value any))
}

// DefStatusItem 创建或者获取 StatusValue,因为golang泛型限制,不能使用 StatusHolder.Get(key) 来获取值
func DefStatusItem[T comparable](status StatusHolder, key StatusKey[T]) StatusValue[T] {
    value := status.Get(key)
    if r, ok := value.(StatusValue[T]); ok {
        return r
    }
    r := &statusValue[T]{key.InitValue()}
    status.Set(key, r)
    return r
}

func SaveNewStatusHolder(parent context.Context) context.Context {
    return SaveStatusHolder(parent, NewStatusTraceCtx())
}

const (
    keyStatusTrace _st = iota
)

type _st int8

func SaveStatusHolder(parent context.Context, holder StatusHolder) context.Context {
    return context.WithValue(parent, keyStatusTrace, holder)
}

// StatusHolderFromCtx 从context中获取 StatusHolder,你需要先存入上下文 ctx = SaveNewStatusHolder(ctx)
func StatusHolderFromCtx(ctx context.Context) StatusHolder {
    v := ctx.Value(keyStatusTrace)
    if s, ok := v.(StatusHolder); ok {
        return s
    }
    panic("context 中没有找到 keyStatusTrace")
}

// DefStatusItemFromCtx 从context中获取 StatusValue的便捷方式,你需要先存入上下文 ctx = SaveNewStatusHolder(ctx)
func DefStatusItemFromCtx[T comparable](ctx context.Context, key StatusKey[T]) StatusValue[T] {
    return DefStatusItem(StatusHolderFromCtx(ctx), key)
}
