package helper

import (
    "testing"
    "time"
)

func TestDoAfterInit_Success(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        AfterInit("test", func() {})
        success := DoAfterInit()
        if !success {
            t.Errorf("DoAfterInit() = %v; want true", success)
        }
    })
    t.Run("panic start", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        AfterInit("test", func() {
            panic("test panic")
        })
        success := DoAfterInit()
        if success {
            t.Errorf("DoAfterInit() = %v; want false", success)
        }
    })
    t.Run("start multiple times", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        AfterInit("test", func() {})
        _ = DoAfterInit()
        defer func() {
            if r := recover(); r == nil {
                t.Errorf("DoAfterInit() did not panic on second call")
            }
        }()
        DoAfterInit()
    })
    t.Run("multiple functions", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        run := 0
        AfterInit("test", func() { run++ })
        AfterInit("test1", func() { run++ })
        success := DoAfterInit()
        if !success {
            t.Errorf("DoAfterInit() = %v; want true", success)
        }
        if run != 2 {
            t.Errorf("run = %v; want 2", run)
        }
    })
    t.Run("multiple functions with order", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        run := 0
        AfterInitOrder("test", func() {
            if run != 2 {
                t.Fatal("run != 2; want 2")
            }
            run++
        }, 3)
        AfterInitOrder("test1", func() {
            if run != 0 {
                t.Fatal("run != 0; want 0")
            }
            run++
        }, 1)
        AfterInitOrder("test1", func() {
            if run != 1 {
                t.Fatal("run != 1; want 1")
            }
            run++
        }, 2)
        success := DoAfterInit()
        if !success {
            t.Errorf("DoAfterInit() = %v; want true", success)
        }
        if run != 3 {
            t.Errorf("run = %v; want 3", run)
        }
    })
    t.Run("multiple functions with order and panic", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        run := 0
        AfterInitOrder("test", func() {
            if run != 2 {
                t.Fatal("run != 2; want 2")
            }
            run++
        }, 3)
        AfterInitOrder("test1", func() {
            if run != 0 {
                t.Fatal("run != 0; want 0")
            }
            run++
        }, 1)
        AfterInitOrder("test1", func() {
            if run != 1 {
                t.Fatal("run != 1; want 1")
            }
            run++
            panic("test panic")
        }, 2)
        success := DoAfterInit()
        if success {
            t.Errorf("DoAfterInit() = %v; want false", success)
        }
        if run != 3 {
            t.Errorf("run = %v; want 3", run)
        }
    })
    t.Run("multiple functions order same", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        run := 0
        AfterInitOrder("test", func() { run++ }, 1)
        AfterInitOrder("test1", func() {
            if run != 2 {
                t.Fatal("run != 1; want 1")
            }
            run++
        }, 2)
        AfterInitOrder("test1", func() { run++ }, 1)
        success := DoAfterInit()
        if !success {
            t.Errorf("DoAfterInit() = %v; want true", success)
        }
        if run != 3 {
            t.Errorf("run = %v; want 3", run)
        }
    })
    t.Run("multiple functions order same1", func(t *testing.T) {
        inited = false
        initCall = &FuncCaller{}
        run := 0
        d := time.Millisecond * 20
        for i := 0; i < 10; i++ {
            AfterInitOrder("test", func() {
                time.Sleep(d)
                run++
            }, 1)
        }
        start := time.Now()
        success := DoAfterInit()
        if !success {
            t.Errorf("DoAfterInit() = %v; want true", success)
        }
        since := time.Since(start)
        if since < d {
            t.Errorf("DoAfterInit() = %v; want > 10ms", since)
        }
        if since > d*2 {
            t.Errorf("DoAfterInit() = %v; want < 20ms", since)
        }
        if run != 10 {
            t.Errorf("run = %v; want 10", run)
        }
    })
}
