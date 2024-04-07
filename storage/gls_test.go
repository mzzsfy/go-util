package storage

import (
    "fmt"
    "github.com/mzzsfy/go-util/concurrent"
    "github.com/mzzsfy/go-util/helper"
    "runtime"
    "strconv"
    "sync"
    "testing"
)

func TestMain(m *testing.M) {
    KnowHowToUseGls()
    m.Run()
    knowHowToUseGls = false
}

func Test_itemGet(t *testing.T) {
    t.Run("value exists", func(t *testing.T) {
        item := NewGlsItem[string]()
        defer item.GlsClean()
        item.Set("testValue1")
        item.Set("testValue")
        value, ok := item.Get()
        Equal(t, true, ok)
        Equal(t, "testValue", value)
    })

    t.Run("value does not exist", func(t *testing.T) {
        defer GlsClean()
        nonexistentKey := NewGlsItem[string]()
        _, ok := nonexistentKey.Get()
        Equal(t, false, ok)
    })
}

func Test_Set(t *testing.T) {
    t.Run("set new value", func(t *testing.T) {
        item := NewGlsItem[string]()
        defer GlsClean()
        item.Set("testValue")
        value, ok := item.Get()
        Equal(t, true, ok)
        Equal(t, "testValue", value)
    })

    t.Run("overwrite existing value", func(t *testing.T) {
        item := NewGlsItem[string]()
        defer GlsClean()
        item.Set("testValue")
        item.Set("newValue")
        value, ok := item.Get()
        Equal(t, true, ok)
        Equal(t, "newValue", value)
    })
}

func Test_Del(t *testing.T) {
    t.Run("delete existing value", func(t *testing.T) {
        item := NewGlsItem[string]()
        defer GlsClean()
        item.Set("testValue")
        item.Delete()
        _, ok := item.Get()
        Equal(t, false, ok)
    })

    t.Run("delete nonexistent value", func(t *testing.T) {
        defer GlsClean()
        nonexistentKey := NewGlsItem[string]()
        _, ok := nonexistentKey.Get()
        Equal(t, false, ok)
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
        defer GlsClean()
        item := NewGlsItem[string]()
        GlsClean()
        _, ok := item.Get()
        Equal(t, false, ok)
    })
}

func TestGet2(t *testing.T) {
    n := 1000
    wg := helper.NewWaitGroup(n)
    item := NewGlsItem[int]()
    f := func(i int) {
        defer GlsClean()
        item.Set(i)
        for j := 0; j < 10; j++ {
            value, ok := item.Get()
            Equal(t, true, ok)
            Equal(t, i, value)
            runtime.Gosched()
        }
    }
    for i := 0; i < n; i++ {
        i := i
        go func() {
            defer wg.Done()
            f(i)
        }()
    }
    wg.Wait()
    t.Log("ok")
}

func Test_check(t *testing.T) {
    defer func() {
        defer glsMap.Clean()
        r := recover()
        if r == nil {
            t.Errorf("check should panic")
            t.FailNow()
            return
        } else if _, ok := r.(GlsError); !ok {
            t.Errorf("check should panic with GlsError")
            t.FailNow()
        }
        keyIdGen = 10000
    }()
    item := NewGlsItem[string]()
    wg := helper.NewWaitGroup(1000)
    for i := 0; i < 1000; i++ {
        go func() {
            defer func() { recover(); wg.Done() }()
            item.Set("testValue")
        }()
    }
    wg.Wait()
    for i := 0; i < 10_000_00; i++ {
        check()
    }
}

func BenchmarkGls(b *testing.B) {
    l := 10
    goNum := 1000
    for i := 0; i < 3; i++ {
        b.Run("BenchmarkGls_string", func(b *testing.B) {
            b.SetParallelism(goNum)
            value := "aaa"
            value1 := "bbb"
            fmt.Sprint(value, value1)
            b.RunParallel(func(pb *testing.PB) {
                defer GlsClean()
                var items []Key[string]
                for i := 1; i <= l; i++ {
                    items = append(items, NewGlsItem[string]())
                }
                x := l * 10
                l1 := l
                for pb.Next() {
                    for i := 0; i < x; i++ {
                        items[(i+0)%l1].Set(value)
                        items[(i+1)%l1].Get()
                        items[(i+2)%l1].Delete()
                        items[(i+3)%l1].Get()
                        items[(i+4)%l1].Set(value1)
                        items[(i+5)%l1].Get()
                        items[(i+1)%l1].Set(value)
                        items[(i+6)%l1].Delete()
                        runtime.Gosched()
                    }
                }
            })
        })
        b.Run("BenchmarkGls_int", func(b *testing.B) {
            b.SetParallelism(goNum)
            value := 1
            value1 := 2
            fmt.Sprint(value, value1)
            b.RunParallel(func(pb *testing.PB) {
                defer GlsClean()
                var items []Key[int]
                for i := 1; i <= l; i++ {
                    items = append(items, NewGlsItem[int]())
                }
                x := l * 10
                l1 := l
                for pb.Next() {
                    for i := 0; i < x; i++ {
                        items[(i+0)%l1].Set(value)
                        items[(i+1)%l1].Get()
                        items[(i+2)%l1].Delete()
                        items[(i+3)%l1].Get()
                        items[(i+4)%l1].Set(value1)
                        items[(i+5)%l1].Get()
                        items[(i+1)%l1].Set(value)
                        items[(i+6)%l1].Delete()
                        runtime.Gosched()
                    }
                }
            })
        })
        b.Run("BenchmarkGls_obj", func(b *testing.B) {
            b.SetParallelism(goNum)
            value := struct {
                aaa string
                bbb int
            }{"aaa", 2}
            value1 := struct {
                aaa string
                bbb int
            }{"bbb", 3}
            fmt.Sprint(value, value1)
            b.RunParallel(func(pb *testing.PB) {
                defer GlsClean()
                var items []Key[struct {
                    aaa string
                    bbb int
                }]
                for i := 1; i <= l; i++ {
                    items = append(items, NewGlsItem[struct {
                        aaa string
                        bbb int
                    }]())
                }
                x := l * 10
                l1 := l
                for pb.Next() {
                    for i := 0; i < x; i++ {
                        items[(i+0)%l1].Set(value)
                        items[(i+1)%l1].Get()
                        items[(i+2)%l1].Delete()
                        items[(i+3)%l1].Get()
                        items[(i+4)%l1].Set(value1)
                        items[(i+5)%l1].Get()
                        items[(i+1)%l1].Set(value)
                        items[(i+6)%l1].Delete()
                        runtime.Gosched()
                    }
                }
            })
        })
    }
}

func BenchmarkGlsSubMapType(b *testing.B) {
    b.Cleanup(func() {
        glsMap.Clean()
        keyIdGen = 10000
        glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeArray[uint32, any](2)) }}
    })
    goNum := 1000
    for _, l := range []int{5, 15, 50} {
        for i := 1; i <= 3; i++ {
            i := i
            name := ""
            switch i {
            case 1:
                name = "Go"
            case 2:
                name = "Swiss"
            default:
                name = "Array"
            }
            b.Run("BenchmarkGlsSubMapType_"+strconv.Itoa(l)+"_"+name, func(b *testing.B) {
                b.SetParallelism(goNum)
                switch i {
                case 1:
                    glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeGo[uint32, any](1)) }}
                case 2:
                    glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeSwiss[uint32, any](1)) }}
                default:
                    glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeArray[uint32, any](1)) }}
                }
                value := 1
                value1 := 2
                fmt.Sprint(value, value1)
                b.RunParallel(func(pb *testing.PB) {
                    defer GlsClean()
                    var items []Key[int]
                    for i := 1; i <= l; i++ {
                        items = append(items, NewGlsItem[int]())
                    }
                    x := 200
                    l1 := l
                    for pb.Next() {
                        for i := 0; i < x; i++ {
                            items[(i+0)%l1].Set(value)
                            items[(i+1)%l1].Get()
                            items[(i+2)%l1].Delete()
                            items[(i+3)%l1].Get()
                            items[(i+4)%l1].Set(value1)
                            items[(i+5)%l1].Get()
                            items[(i+1)%l1].Set(value)
                            items[(i+6)%l1].Delete()
                            runtime.Gosched()
                        }
                    }
                })
            })
        }
    }
}

func BenchmarkGlsLock(b *testing.B) {
    b.Cleanup(func() {
        glsMap.Clean()
        glsLock = &sync.RWMutex{}
        glsMap = NewMap(MapTypeSwiss[int64, Map[uint32, any]](1))
        keyIdGen = 10000
    })
    goNum := 1000
    l := 5
    for i := 1; i <= 3; i++ {
        i := i
        name := ""
        switch i {
        case 1:
            name = "RWMutex"
        case 2:
            name = "CasRwLock"
        default:
            name = "ConcurrentMap"
        }
        b.Run("BenchmarkGlsSubMapType_"+name, func(b *testing.B) {
            b.SetParallelism(goNum)
            switch i {
            case 1:
                glsLock = &sync.RWMutex{}
                glsMap = NewMap(MapTypeSwiss[int64, Map[uint32, any]](1))
            case 2:
                glsLock = &concurrent.CasRwLocker{}
                glsMap = NewMap(MapTypeSwiss[int64, Map[uint32, any]](1))
            default:
                glsLock = concurrent.NoLock{}
                glsMap = NewMap(MapTypeSwissConcurrent[int64, Map[uint32, any]]())
            }
            value := 1
            value1 := 2
            fmt.Sprint(value, value1)
            b.RunParallel(func(pb *testing.PB) {
                defer GlsClean()
                var items []Key[int]
                for i := 1; i <= l; i++ {
                    items = append(items, NewGlsItem[int]())
                }
                x := l * 10
                l1 := l
                for pb.Next() {
                    for i := 0; i < x; i++ {
                        items[(i+0)%l1].Set(value)
                        items[(i+1)%l1].Get()
                        items[(i+2)%l1].Delete()
                        items[(i+3)%l1].Get()
                        items[(i+4)%l1].Set(value1)
                        items[(i+5)%l1].Get()
                        items[(i+1)%l1].Set(value)
                        items[(i+6)%l1].Delete()
                        runtime.Gosched()
                    }
                }
            })
        })
    }
}
