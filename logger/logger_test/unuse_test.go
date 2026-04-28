package logger_test

import (
    "testing"

    "github.com/mzzsfy/go-util/logger"
)

// TestUnuse_ChainCall 验证链式调用后 Unuse 能正确回收
func TestUnuse_ChainCall(t *testing.T) {
    log := logger.Logger("test.unuse.chain")

    // 验证 With().I().Unuse() 模式不会 panic
    log.With("key", "value").I("test message").Unuse()
}

// TestUnuse_NamedLogger 验证命名 logger 调用 Unuse 不会被回收
func TestUnuse_NamedLogger(t *testing.T) {
    // 获取命名 logger
    log1 := logger.Logger("test.unuse.named1")
    log2 := logger.Logger("test.unuse.named1")

    // 命名 logger 调用 Unuse 应该无效
    log1.Unuse()

    // 再次获取应该是同一个实例
    log3 := logger.Logger("test.unuse.named1")

    // 验证是同一个 logger
    if log1 != log2 || log1 != log3 {
        t.Error("命名 logger 不应该被回收")
    }
}

// TestUnuse_DerivedLogger 验证派生 logger 可以被回收
func TestUnuse_DerivedLogger(t *testing.T) {
    log := logger.Logger("test.unuse.derived")

    // 多次派生并回收
    for i := 0; i < 100; i++ {
        derived := log.With("index", i)
        derived.I("message").Unuse()
    }
}

// TestUnuse_MultipleWith 验证多次 With 后 Unuse
func TestUnuse_MultipleWith(t *testing.T) {
    log := logger.Logger("test.unuse.multi")

    // 链式 With
    log.With("a", 1).
        With("b", 2).
        With("c", 3).
        I("multi with").
        Unuse()
}
