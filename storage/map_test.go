package storage

import (
    "github.com/mzzsfy/go-util/seq"
    "math/rand"
    "strconv"
    "testing"
)

func Test_AllMap(t *testing.T) {
    for _, m := range []struct {
        name string
        m    func() Map[string, int]
        m1   func() Map[uint32, int]
    }{
        {"Go", MapTypeGo[string, int](0).createMap, MapTypeGo[uint32, int](0).createMap},
        {"Swiss", MapTypeSwiss[string, int](0).createMap, MapTypeSwiss[uint32, int](0).createMap},
        {"SwissConcurrent", MapTypeSwissConcurrent[string, int]().createMap, MapTypeSwissConcurrent[uint32, int]().createMap},
        {"Array", MapTypeArray[string, int](0).createMap, MapTypeArray[uint32, int](0).createMap},
    } {
        t.Run(m.name+"_strings=0", func(t *testing.T) {
            testMap(t, genStringData(16, 0), m.m)
        })
        t.Run(m.name+"_strings=100", func(t *testing.T) {
            testMap(t, genStringData(16, 100), m.m)
        })
        t.Run(m.name+"_strings=1000", func(t *testing.T) {
            testMap(t, genStringData(16, 1000), m.m)
        })
        if m.name != "Array" {
            t.Run(m.name+"_strings=10_000", func(t *testing.T) {
                testMap(t, genStringData(16, 10_000), m.m)
            })
            //t.Run(m.name+"_strings=100_000", func(t *testing.T) {
            //    testMap(t, genStringData(16, 100_000), m.m)
            //})
        }
        t.Run(m.name+"_uint32=0", func(t *testing.T) {
            testMap(t, genUint32Data(0), m.m1)
        })
        t.Run(m.name+"_uint32=100", func(t *testing.T) {
            testMap(t, genUint32Data(100), m.m1)
        })
        t.Run(m.name+"_uint32=1000", func(t *testing.T) {
            testMap(t, genUint32Data(1000), m.m1)
        })
        if m.name != "Array" {
            t.Run(m.name+"_uint32=10_000", func(t *testing.T) {
                testMap(t, genUint32Data(10_000), m.m1)
            })
            //t.Run(m.name+"_uint32=100_000", func(t *testing.T) {
            //    testMap(t, genUint32Data(100_000), m.m1)
            //})
        }
        t.Run(m.name+"_string capacity", func(t *testing.T) {
            testMapCapacity(t, func(n int) []string {
                return genStringData(16, n)
            }, m.m)
        })
        t.Run(m.name+"_uint32 capacity", func(t *testing.T) {
            testMapCapacity(t, genUint32Data, m.m1)
        })
    }
}

func BenchmarkMap(b *testing.B) {
    n := 1
    t1 := []struct {
        name string
        m    func() Map[int, int]
    }{
        {"Go", MapTypeGo[int, int](0).createMap},
        {"Swiss", MapTypeSwiss[int, int](0).createMap},
        {"Array", MapTypeArray[int, int](0).createMap},
    }
    for _, n1 := range []int{16, 256, 2048, 65536} {
        for _, m := range t1 {
            b.Run(m.name+"_r_"+strconv.Itoa(n1), func(b *testing.B) {
                m2 := m.m()
                for i := 0; i < n1; i++ {
                    m2.Put(i, i)
                }
                b.ResetTimer()
                v := b.N * n
                for i := 0; i < v; i++ {
                    m2.Get(i % n1)
                }
            })
        }
        for _, m := range t1 {
            b.Run(m.name+"_w_"+strconv.Itoa(n1), func(b *testing.B) {
                m2 := m.m()
                b.ResetTimer()
                v := b.N * n
                for i := 0; i < v; i++ {
                    m2.Put(i%n1, i)
                }
            })
        }
        for _, m := range t1 {
            b.Run(m.name+"_rw_"+strconv.Itoa(n1), func(b *testing.B) {
                m2 := m.m()
                b.ResetTimer()
                v := b.N * n
                for i := 0; i < v; i++ {
                    if i%2 == 0 {
                        m2.Get(i % n1)
                    } else {
                        m2.Put(i%n1, i)
                    }
                }
            })
        }
    }
}

func testMapCapacity[K comparable](t *testing.T, gen func(n int) []K, makeMap func() Map[K, int]) {
    caps := []uint32{
        1 * maxAvgGroupLoad,
        2 * maxAvgGroupLoad,
        3 * maxAvgGroupLoad,
        4 * maxAvgGroupLoad,
        5 * maxAvgGroupLoad,
        10 * maxAvgGroupLoad,
        25 * maxAvgGroupLoad,
        50 * maxAvgGroupLoad,
        100 * maxAvgGroupLoad,
    }
    for _, c := range caps {
        m := makeMap()
        Equal(t, 0, m.Count())
        keys := gen(rand.Intn(int(c)))
        for i, k := range keys {
            m.Put(k, i)
        }
        Equal(t, len(keys), m.Count())
    }
}

func testMap[K comparable](t *testing.T, keys []K, makeMap func() Map[K, int]) {
    if len(keys) != len(uniq(keys)) {
        t.Fatalf("keys are not unique")
    }
    t.Run("put", func(t *testing.T) {
        testMapPut(t, keys, makeMap())
    })
    t.Run("has", func(t *testing.T) {
        testMapHas(t, keys, makeMap())
    })
    t.Run("get", func(t *testing.T) {
        testMapGet(t, keys, makeMap())
    })
    t.Run("delete", func(t *testing.T) {
        testMapDelete(t, keys, makeMap())
    })
    t.Run("clear", func(t *testing.T) {
        testMapClear(t, keys, makeMap())
    })
    t.Run("iter", func(t *testing.T) {
        testMapIter(t, keys, makeMap())
    })
}

func testMapPut[K comparable](t *testing.T, keys []K, m Map[K, int]) {
    Equal(t, 0, m.Count())
    for i, key := range keys {
        m.Put(key, i)
    }
    Equal(t, len(keys), m.Count())
    // overwrite
    for i, key := range keys {
        m.Put(key, -i)
    }
    Equal(t, len(keys), m.Count())
    for i, key := range keys {
        act, ok := m.Get(key)
        True(t, ok)
        Equal(t, -i, act)
    }
    Equal(t, len(keys), m.Count())
}

func testMapHas[K comparable](t *testing.T, keys []K, m Map[K, int]) {
    for i, key := range keys {
        m.Put(key, i)
    }
    for _, key := range keys {
        ok := m.Has(key)
        True(t, ok)
    }
}

func testMapGet[K comparable](t *testing.T, keys []K, m Map[K, int]) {

    for i, key := range keys {
        m.Put(key, i)
    }
    for i, key := range keys {
        act, ok := m.Get(key)
        True(t, ok)
        Equal(t, i, act)
    }
}

func testMapDelete[K comparable](t *testing.T, keys []K, m Map[K, int]) {

    Equal(t, 0, m.Count())
    for i, key := range keys {
        m.Put(key, i)
    }
    Equal(t, len(keys), m.Count())
    for _, key := range keys {
        m.Delete(key)
        ok := m.Has(key)
        True(t, !ok)
    }
    Equal(t, 0, m.Count())
    // put keys back after deleting them
    for i, key := range keys {
        m.Put(key, i)
    }
    Equal(t, len(keys), m.Count())
}

func testMapClear[K comparable](t *testing.T, keys []K, m Map[K, int]) {

    Equal(t, 0, m.Count())
    for i, key := range keys {
        m.Put(key, i)
    }
    Equal(t, len(keys), m.Count())
    m.Clean()
    Equal(t, 0, m.Count())
    for _, key := range keys {
        ok := m.Has(key)
        True(t, !ok)
        _, ok = m.Get(key)
        True(t, !ok)
    }
    var calls int
    m.Iter(func(k K, v int) (stop bool) {
        calls++
        return
    })
    Equal(t, 0, calls)

    m.Iter(func(k K, v int) (stop bool) {
        t.Errorf("unexpected call to Iter: %v, %v", k, v)
        return false
    })
}

func testMapIter[K comparable](t *testing.T, keys []K, m Map[K, int]) {
    for i, key := range keys {
        m.Put(key, i)
    }
    t.Log("Put ok")
    visited := make(map[K]uint, len(keys))
    m.Iter(func(k K, v int) (stop bool) {
        visited[k] = 0
        stop = true
        return
    })
    t.Log("iter ok")
    if len(keys) == 0 {
        Equal(t, len(visited), 0)
    } else {
        Equal(t, len(visited), 1)
    }
    for _, k := range keys {
        visited[k] = 0
    }
    t.Log("iter2 ok")
    m.Iter(func(k K, v int) (stop bool) {
        visited[k]++
        return
    })
    t.Log("iter3 ok")
    for _, c := range visited {
        Equal(t, c, uint(1))
    }
    t.Log("iter4 ok")
    // mutate on iter
    seq.BiFrom(func(t func(k K, v int)) {
        m.Iter(func(k K, v int) (stop bool) {
            t(k, v)
            return
        })
    }).Cache().ForEach(func(k K, v int) {
        m.Put(k, -v)
    })
    t.Log("iter5 ok")
    for i, key := range keys {
        act, ok := m.Get(key)
        True(t, ok)
        Equal(t, -i, act)
    }
    t.Log("iter6 ok")
}
