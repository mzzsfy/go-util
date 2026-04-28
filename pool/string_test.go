package pool

import (
    "sync"
    "sync/atomic"
    "testing"
)

// TestStringPool_UseUnUse_Counting 验证同一字符串多次 Use/UnUse 计数正确
func TestStringPool_UseUnUse_Counting(t *testing.T) {
    p := NewStringPool()
    s := "test-string"

    // 第一次 Use, id 应非零
    id1 := p.Use(s)
    if id1 == 0 {
        t.Fatal("第一次 Use 应返回非零 id")
    }

    // 第二次 Use, id 应相同
    id2 := p.Use(s)
    if id2 != id1 {
        t.Fatalf("第二次 Use 应返回相同 id, got %d, want %d", id2, id1)
    }

    // 第三次 Use
    id3 := p.Use(s)
    if id3 != id1 {
        t.Fatalf("第三次 Use 应返回相同 id, got %d, want %d", id3, id1)
    }

    // 两次 UnUse 后, Peek 仍能找到(id还在)
    p.UnUse(s)
    p.UnUse(s)
    if peeked := p.Peek(s); peeked != id1 {
        t.Fatalf("两次 UnUse 后 Peek 应返回 %d, got %d", id1, peeked)
    }

    // 第三次 UnUse 后, 条目应被删除
    p.UnUse(s)
    if peeked := p.Peek(s); peeked != 0 {
        t.Fatalf("三次 UnUse 后 Peek 应返回 0, got %d", peeked)
    }
}

// TestStringPool_UnUse_DeletesEntry 验证最后一次 UnUse 后条目会被删除
func TestStringPool_UnUse_DeletesEntry(t *testing.T) {
    p := NewStringPool()

    // 单次 Use -> 单次 UnUse, 条目应被删除
    p.Use("a")
    p.UnUse("a")
    if p.Peek("a") != 0 {
        t.Fatal("单次 Use/UnUse 后条目应被删除")
    }

    // 5 次 Use -> 5 次 UnUse
    s := "multi"
    var expectedId uint64
    for i := 0; i < 5; i++ {
        got := p.Use(s)
        if expectedId == 0 {
            expectedId = got
        } else if got != expectedId {
            t.Fatalf("Use 应返回相同 id, got %d, want %d", got, expectedId)
        }
    }
    for i := 0; i < 4; i++ {
        p.UnUse(s)
    }
    // 还剩一次引用
    if p.Peek(s) != expectedId {
        t.Fatal("还剩 1 次引用时 Peek 应找到条目")
    }
    p.UnUse(s)
    if p.Peek(s) != 0 {
        t.Fatal("所有引用释放后条目应被删除")
    }

    // 删除后重新 Use 应分配新 id
    newId := p.Use(s)
    if newId == 0 {
        t.Fatal("重新 Use 应返回非零 id")
    }
    if newId == expectedId {
        // 新 id 可以相同(被复用), 这是实现细节, 不强制不同
        // 但不应该为零
    }
    p.UnUse(s)
}

// TestStringPool_Peek_NoSideEffects 验证 Peek 不影响引用计数
func TestStringPool_Peek_NoSideEffects(t *testing.T) {
    p := NewStringPool()
    s := "peek-test"

    // 未 Use 时 Peek 返回 0
    if id := p.Peek(s); id != 0 {
        t.Fatalf("未 Use 时 Peek 应返回 0, got %d", id)
    }

    // Use 一次
    id := p.Use(s)

    // 多次 Peek
    for i := 0; i < 10; i++ {
        if got := p.Peek(s); got != id {
            t.Fatalf("Peek 应返回 %d, got %d", id, got)
        }
    }

    // 一次 UnUse 即可删除(Peek 不增加计数)
    p.UnUse(s)
    if p.Peek(s) != 0 {
        t.Fatal("Peek 不增加计数, 一次 UnUse 后条目应被删除")
    }
}

// TestStringPool_DifferentStrings 不同字符串获得不同 id
func TestStringPool_DifferentStrings(t *testing.T) {
    p := NewStringPool()
    ids := make(map[uint64]string)

    for _, s := range []string{"a", "b", "c", "d", "e"} {
        id := p.Use(s)
        if id == 0 {
            t.Fatalf("Use(%q) 返回零 id", s)
        }
        if prev, exists := ids[id]; exists {
            t.Fatalf("不同字符串 %q 和 %q 获得相同 id %d", prev, s, id)
        }
        ids[id] = s
    }
}

// TestStringPool_ConcurrentUseUnUse 并发场景下引用计数正确
func TestStringPool_ConcurrentUseUnUse(t *testing.T) {
    p := NewStringPool()
    const goroutines = 50
    const opsPerGoroutine = 100
    const key = "concurrent-key"

    var wg sync.WaitGroup
    var useCount int32

    // 并发 Use
    wg.Add(goroutines)
    for i := 0; i < goroutines; i++ {
        go func() {
            defer wg.Done()
            for j := 0; j < opsPerGoroutine; j++ {
                id := p.Use(key)
                if id == 0 {
                    t.Error("Use 不应返回 0")
                    return
                }
                atomic.AddInt32(&useCount, 1)
            }
        }()
    }
    wg.Wait()

    totalUses := atomic.LoadInt32(&useCount)
    if totalUses != goroutines*opsPerGoroutine {
        t.Fatalf("总 Use 次数: %d, want %d", totalUses, goroutines*opsPerGoroutine)
    }

    // 并发 UnUse 相同次数
    wg.Add(goroutines)
    for i := 0; i < goroutines; i++ {
        go func() {
            defer wg.Done()
            for j := 0; j < opsPerGoroutine; j++ {
                p.UnUse(key)
            }
        }()
    }
    wg.Wait()

    // 所有引用释放后, 条目应被删除
    if p.Peek(key) != 0 {
        t.Fatal("并发 UnUse 完成后条目应被删除")
    }
}

// TestStringPool_UnUseNonExistent 对不存在的字符串调用 UnUse 不 panic
func TestStringPool_UnUseNonExistent(t *testing.T) {
    p := NewStringPool()
    // 不应 panic
    p.UnUse("nonexistent")
    p.UnUse("nonexistent")
}
