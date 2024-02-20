package storage

import (
    "github.com/mzzsfy/go-util/helper"
    "runtime"
    "strconv"
    "testing"
)

func init() {
    KnowHowToUseGls = true
}

func Test_GOGet(t *testing.T) {
    t.Run("value exists", func(t *testing.T) {
        GOSet("testKey", "testValue")
        value, ok := GOGet("testKey")
        Equal(t, true, ok)
        Equal(t, "testValue", value)
        GOClean()
    })

    t.Run("value does not exist", func(t *testing.T) {
        _, ok := GOGet("nonexistentKey")
        Equal(t, false, ok)
        GOClean()
    })
}

func Test_GOSet(t *testing.T) {
    t.Run("set new value", func(t *testing.T) {
        GOSet("testKey", "testValue")
        value, ok := GOGet("testKey")
        Equal(t, true, ok)
        Equal(t, "testValue", value)
        GOClean()
    })

    t.Run("overwrite existing value", func(t *testing.T) {
        GOSet("testKey", "testValue")
        GOSet("testKey", "newValue")
        value, ok := GOGet("testKey")
        Equal(t, true, ok)
        Equal(t, "newValue", value)
        GOClean()
    })
}

func Test_GODel(t *testing.T) {
    t.Run("delete existing value", func(t *testing.T) {
        GOSet("testKey", "testValue")
        GODel("testKey")
        _, ok := GOGet("testKey")
        Equal(t, false, ok)
        GOClean()
    })

    t.Run("delete nonexistent value", func(t *testing.T) {
        GODel("nonexistentKey")
        _, ok := GOGet("nonexistentKey")
        Equal(t, false, ok)
        GOClean()
    })
}

func Test_GOClean(t *testing.T) {
    t.Run("clean existing values", func(t *testing.T) {
        GOSet("testKey", "testValue")
        GOClean()
        _, ok := GOGet("testKey")
        Equal(t, false, ok)
    })

    t.Run("clean when no values exist", func(t *testing.T) {
        GOClean()
        _, ok := GOGet("testKey")
        Equal(t, false, ok)
    })
}

func Test_GOGet2(t *testing.T) {
    n := 1000
    wg := helper.NewWaitGroup(n)
    f := func(i int) {
        v := strconv.Itoa(i)
        GOSet("testKey", v)
        for j := 0; j < 10; j++ {
            value, ok := GOGet("testKey")
            Equal(t, true, ok)
            Equal(t, v, value)
            runtime.Gosched()
        }
        GOClean()
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
    for i := 0; i < 100; i++ {
        go func() {
            defer func() { recover() }()
            GOSet("testKey", "testValue")
        }()
    }
    for i := 0; i < 1_000_000; i++ {
        check()
    }
}
