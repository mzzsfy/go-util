package helper

import (
    "context"
    "math/rand"
    "strconv"
    "testing"
)

var k1 = DefStatusKeyStatic(int64(1))
var k2 = DefStatusKeyStatic(int64(1))
var k3 = DefStatusKeyFn(func() string {
    return strconv.Itoa(rand.Int())
})

func TestGetItem(t *testing.T) {
    ctx := NewStatusTraceCtx()
    DefItem(ctx, k1).Set(11)
    DefItem(ctx, k2).Set(12)
    DefItem(ctx, k3).Set("abc")
    if DefItem(ctx, k1).Value() != 11 {
        t.Errorf("Expected 11, got %d", DefItem(ctx, k1).Value())
    }
    if DefItem(ctx, k2).Value() != 12 {
        t.Errorf("Expected 12, got %d", DefItem(ctx, k2).Value())
    }
    if DefItem(ctx, k3).Value() != "abc" {
        t.Errorf("Expected abc, got %s", DefItem(ctx, k3).Value())
    }
}

func TestGetItem1(t *testing.T) {
    ctx := SaveNewStatusHolder(context.Background())
    DefItemFromCtx(ctx, k1).Set(11)
    DefItemFromCtx(ctx, k2).Set(12)
    DefItemFromCtx(ctx, k3).Set("abc")
    if DefItemFromCtx(ctx, k1).Value() != 11 {
        t.Errorf("Expected 11, got %d", DefItemFromCtx(ctx, k1).Value())
    }
    if DefItemFromCtx(ctx, k2).Value() != 12 {
        t.Errorf("Expected 12, got %d", DefItemFromCtx(ctx, k2).Value())
    }
    if DefItemFromCtx(ctx, k3).Value() != "abc" {
        t.Errorf("Expected abc, got %s", DefItemFromCtx(ctx, k3).Value())
    }
}
