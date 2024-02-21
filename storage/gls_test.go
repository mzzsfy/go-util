package storage

import (
    "github.com/mzzsfy/go-util/helper"
    "runtime"
    "testing"
)

func init() {
    KnowHowToUseGls = true
}

func Test_itemGet(t *testing.T) {
    t.Run("value exists", func(t *testing.T) {
        item := NewGlsItem[string]()
        item.Set("testValue")
        value, ok := item.Get()
        Equal(t, true, ok)
        Equal(t, "testValue", value)
        item.GlsClean()
    })

    t.Run("value does not exist", func(t *testing.T) {
        nonexistentKey := NewGlsItem[string]()
        _, ok := nonexistentKey.Get()
        Equal(t, false, ok)
        GlsClean()
    })
}

func Test_Set(t *testing.T) {
    t.Run("set new value", func(t *testing.T) {
        item := NewGlsItem[string]()
        item.Set("testValue")
        value, ok := item.Get()
        Equal(t, true, ok)
        Equal(t, "testValue", value)
        GlsClean()
    })

    t.Run("overwrite existing value", func(t *testing.T) {
        item := NewGlsItem[string]()
        item.Set("testValue")
        item.Set("newValue")
        value, ok := item.Get()
        Equal(t, true, ok)
        Equal(t, "newValue", value)
        GlsClean()
    })
}

func Test_Del(t *testing.T) {
    t.Run("delete existing value", func(t *testing.T) {
        item := NewGlsItem[string]()
        item.Set("testValue")
        item.Delete()
        _, ok := item.Get()
        Equal(t, false, ok)
        GlsClean()
    })

    t.Run("delete nonexistent value", func(t *testing.T) {
        nonexistentKey := NewGlsItem[string]()
        _, ok := nonexistentKey.Get()
        Equal(t, false, ok)
        GlsClean()
    })
}

func Test_Clean(t *testing.T) {
    t.Run("clean existing values", func(t *testing.T) {
        item := NewGlsItem[string]()
        item.Set("testValue")
        GlsClean()
        _, ok := item.Get()
        Equal(t, false, ok)
    })

    t.Run("clean when no values exist", func(t *testing.T) {
        item := NewGlsItem[string]()
        GlsClean()
        _, ok := item.Get()
        Equal(t, false, ok)
    })
}

func Test_Get2(t *testing.T) {
    n := 1000
    wg := helper.NewWaitGroup(n)
    item := NewGlsItem[int]()
    f := func(i int) {
        item.Set(i)
        for j := 0; j < 10; j++ {
            value, ok := item.Get()
            Equal(t, true, ok)
            Equal(t, i, value)
            runtime.Gosched()
        }
        GlsClean()
    }
    for i := 0; i < n; i++ {
        i := i
        go func() {
            defer wg.Done()
            f(i)
        }()
    }
    wg.Wait()
}

func Test_check(t *testing.T) {
    defer func() {
        r := recover()
        if r == nil {
            t.Errorf("check should panic")
            return
        } else if _, ok := r.(GlsError); !ok {
            t.Errorf("check should panic with GlsError")
        }
        glsMap.Clear()
    }()
    item := NewGlsItem[string]()
    for i := 0; i < 1000; i++ {
        go func() {
            defer func() { recover() }()
            item.Set("testValue")
        }()
    }
    for i := 0; i < 10_000_000; i++ {
        check()
    }
}
