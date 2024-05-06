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
    DefStatusItem(ctx, k1).Set(11)
    DefStatusItem(ctx, k2).Set(12)
    DefStatusItem(ctx, k3).Set("abc")
    if DefStatusItem(ctx, k1).Value() != 11 {
        t.Errorf("Expected 11, got %d", DefStatusItem(ctx, k1).Value())
    }
    if DefStatusItem(ctx, k2).Value() != 12 {
        t.Errorf("Expected 12, got %d", DefStatusItem(ctx, k2).Value())
    }
    if DefStatusItem(ctx, k3).Value() != "abc" {
        t.Errorf("Expected abc, got %s", DefStatusItem(ctx, k3).Value())
    }
}

func TestGetItem1(t *testing.T) {
    ctx := SaveNewStatusHolder(context.Background())
    DefStatusItemFromCtx(ctx, k1).Set(11)
    DefStatusItemFromCtx(ctx, k2).Set(12)
    DefStatusItemFromCtx(ctx, k3).Set("abc")
    if DefStatusItemFromCtx(ctx, k1).Value() != 11 {
        t.Errorf("Expected 11, got %d", DefStatusItemFromCtx(ctx, k1).Value())
    }
    if DefStatusItemFromCtx(ctx, k2).Value() != 12 {
        t.Errorf("Expected 12, got %d", DefStatusItemFromCtx(ctx, k2).Value())
    }
    if DefStatusItemFromCtx(ctx, k3).Value() != "abc" {
        t.Errorf("Expected abc, got %s", DefStatusItemFromCtx(ctx, k3).Value())
    }
}
