package storage

import (
    "math"
    "math/rand"
    "testing"
)

func Equal(t *testing.T, a, b interface{}) {
    if a != b {
        t.Errorf("expected %v, got %v", a, b)
    }
}
func True(t *testing.T, a bool) {
    if !a {
        t.Errorf("expected true, got false")
    }
}

func Test_SwissMap(t *testing.T) {
    t.Run("strings=0", func(t *testing.T) {
        testSwissMap(t, genStringData(16, 0))
    })
    t.Run("strings=100", func(t *testing.T) {
        testSwissMap(t, genStringData(16, 100))
    })
    t.Run("strings=1000", func(t *testing.T) {
        testSwissMap(t, genStringData(16, 1000))
    })
    t.Run("strings=10_000", func(t *testing.T) {
        testSwissMap(t, genStringData(16, 10_000))
    })
    t.Run("strings=100_000", func(t *testing.T) {
        testSwissMap(t, genStringData(16, 100_000))
    })
    t.Run("uint32=0", func(t *testing.T) {
        testSwissMap(t, genUint32Data(0))
    })
    t.Run("uint32=100", func(t *testing.T) {
        testSwissMap(t, genUint32Data(100))
    })
    t.Run("uint32=1000", func(t *testing.T) {
        testSwissMap(t, genUint32Data(1000))
    })
    t.Run("uint32=10_000", func(t *testing.T) {
        testSwissMap(t, genUint32Data(10_000))
    })
    t.Run("uint32=100_000", func(t *testing.T) {
        testSwissMap(t, genUint32Data(100_000))
    })
    t.Run("string capacity", func(t *testing.T) {
        testSwissMapCapacity(t, func(n int) []string {
            return genStringData(16, n)
        })
    })
    t.Run("uint32 capacity", func(t *testing.T) {
        testSwissMapCapacity(t, genUint32Data)
    })
}

func testSwissMap[K comparable](t *testing.T, keys []K) {
    // sanity check
    if len(keys) != len(uniq(keys)) {
        t.Fatalf("keys are not unique")
    }
    t.Run("put", func(t *testing.T) {
        testMapPut(t, keys)
    })
    t.Run("has", func(t *testing.T) {
        testMapHas(t, keys)
    })
    t.Run("get", func(t *testing.T) {
        testMapGet(t, keys)
    })
    t.Run("delete", func(t *testing.T) {
        testMapDelete(t, keys)
    })
    t.Run("clear", func(t *testing.T) {
        testMapClear(t, keys)
    })
    t.Run("iter", func(t *testing.T) {
        testMapIter(t, keys)
    })
    t.Run("grow", func(t *testing.T) {
        testMapGrow(t, keys)
    })
}

func uniq[K comparable](keys []K) []K {
    s := make(map[K]struct{}, len(keys))
    for _, k := range keys {
        s[k] = struct{}{}
    }
    u := make([]K, 0, len(keys))
    for k := range s {
        u = append(u, k)
    }
    return u
}

func genStringData(size, count int) (keys []string) {
    src := rand.New(rand.NewSource(int64(size * count)))
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    r := make([]rune, size*count)
    for i := range r {
        r[i] = letters[src.Intn(len(letters))]
    }
    keys = make([]string, count)
    for i := range keys {
        keys[i] = string(r[:size])
        r = r[size:]
    }
    return
}

func genUint32Data(count int) (keys []uint32) {
    keys = make([]uint32, count)
    var x uint32
    for i := range keys {
        x += (rand.Uint32() % 128) + 1
        keys[i] = x
    }
    return
}

func testMapPut[K comparable](t *testing.T, keys []K) {
    m := makeSwissMap[K, int](uint32(len(keys)))
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
    Equal(t, len(keys), int(m.resident))
}

func testMapHas[K comparable](t *testing.T, keys []K) {
    m := makeSwissMap[K, int](uint32(len(keys)))
    for i, key := range keys {
        m.Put(key, i)
    }
    for _, key := range keys {
        ok := m.Has(key)
        True(t, ok)
    }
}

func testMapGet[K comparable](t *testing.T, keys []K) {
    m := makeSwissMap[K, int](uint32(len(keys)))
    for i, key := range keys {
        m.Put(key, i)
    }
    for i, key := range keys {
        act, ok := m.Get(key)
        True(t, ok)
        Equal(t, i, act)
    }
}

func testMapDelete[K comparable](t *testing.T, keys []K) {
    m := makeSwissMap[K, int](uint32(len(keys)))
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

func testMapClear[K comparable](t *testing.T, keys []K) {
    m := makeSwissMap[K, int](0)
    Equal(t, 0, m.Count())
    for i, key := range keys {
        m.Put(key, i)
    }
    Equal(t, len(keys), m.Count())
    m.Clear()
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

    // Assert that the map was actually cleared...
    var k K
    for _, g := range m.groups {
        for i := range g.keys {
            Equal(t, k, g.keys[i])
            Equal(t, 0, g.values[i])
        }
    }
}

func testMapIter[K comparable](t *testing.T, keys []K) {
    m := makeSwissMap[K, int](uint32(len(keys)))
    for i, key := range keys {
        m.Put(key, i)
    }
    visited := make(map[K]uint, len(keys))
    m.Iter(func(k K, v int) (stop bool) {
        visited[k] = 0
        stop = true
        return
    })
    if len(keys) == 0 {
        Equal(t, len(visited), 0)
    } else {
        Equal(t, len(visited), 1)
    }
    for _, k := range keys {
        visited[k] = 0
    }
    m.Iter(func(k K, v int) (stop bool) {
        visited[k]++
        return
    })
    for _, c := range visited {
        Equal(t, c, uint(1))
    }
    // mutate on iter
    m.Iter(func(k K, v int) (stop bool) {
        m.Put(k, -v)
        return
    })
    for i, key := range keys {
        act, ok := m.Get(key)
        True(t, ok)
        Equal(t, -i, act)
    }
}

func testMapGrow[K comparable](t *testing.T, keys []K) {
    n := uint32(len(keys))
    m := makeSwissMap[K, int](n / 10)
    for i, key := range keys {
        m.Put(key, i)
    }
    for i, key := range keys {
        act, ok := m.Get(key)
        True(t, ok)
        Equal(t, i, act)
    }
}

func testSwissMapCapacity[K comparable](t *testing.T, gen func(n int) []K) {
    // Capacity() behavior depends on |groupSize|
    // which varies by processor architecture.
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
        m := makeSwissMap[K, K](c)
        m1 := m
        Equal(t, int(c), m1.Capacity())
        keys := gen(rand.Intn(int(c)))
        for _, k := range keys {
            m.Put(k, k)
        }
        Equal(t, int(c)-len(keys), m1.Capacity())
        Equal(t, int(c), m.Count()+m1.Capacity())
    }
}

func Test_SwissMap_IterDelete(t *testing.T) {
    t.Run("IterDelete=0", func(t *testing.T) {
        testIterDelete(t, 0, 0)
        testIterDelete2(t, 0, 0)
        testIterDelete3(t, 0, 0, 0)
    })
    t.Run("IterDelete=10,1", func(t *testing.T) {
        testIterDelete(t, 10, 1)
        testIterDelete2(t, 10, 1)
        testIterDelete3(t, 10, 1, 3)
    })
    t.Run("IterDelete=100,10", func(t *testing.T) {
        testIterDelete(t, 100, 10)
        testIterDelete2(t, 100, 10)
        testIterDelete3(t, 100, 10, 70)
    })
    t.Run("IterDelete=100,100", func(t *testing.T) {
        testIterDelete(t, 1000, 100)
        testIterDelete2(t, 1000, 100)
        testIterDelete3(t, 1000, 10, 900)
    })
    t.Run("IterDelete=1000,1", func(t *testing.T) {
        testIterDelete(t, 1000, 1)
        testIterDelete2(t, 1000, 1)
        testIterDelete3(t, 1000, 300, 100)
    })
    t.Run("IterDelete=1000,100", func(t *testing.T) {
        testIterDelete(t, 1000, 100)
        testIterDelete2(t, 1000, 100)
        testIterDelete3(t, 1000, 500, 100)
    })
    t.Run("IterDelete=1000,1000", func(t *testing.T) {
        testIterDelete(t, 1000, 1000)
        testIterDelete2(t, 1000, 1000)
        testIterDelete3(t, 1000, 100, 100)
    })
}

func testIterDelete(t *testing.T, all, delete int) {
    t.Helper()
    m := makeSwissMap[int, int](0)
    for i := 0; i < all; i++ {
        m.Put(i, i)
    }
    var calls int
    m.IterDelete(func(k int, v int) (del, stop bool) {
        calls++
        return true, calls == delete
    })
    Equal(t, delete, calls)
    Equal(t, all-delete, m.Count())
}

func testIterDelete2(t *testing.T, all, delete int) {
    t.Helper()
    m := makeSwissMap[int, int](0)
    for i := 0; i < all; i++ {
        m.Put(i, i)
    }
    var calls int
    m.IterDelete(func(k int, v int) (del, stop bool) {
        if calls < delete {
            calls++
            return true, false
        }
        return false, false
    })
    Equal(t, delete, calls)
    Equal(t, all-delete, m.Count())
}

func testIterDelete3(t *testing.T, all, delete, skip int) {
    t.Helper()
    m := makeSwissMap[int, int](0)
    for i := 0; i < all; i++ {
        m.Put(i, i)
    }
    var calls int
    var i int
    m.IterDelete(func(k int, v int) (del, stop bool) {
        i++
        if i > skip && calls < delete {
            calls++
            return true, false
        }
        return false, false
    })
    Equal(t, delete, calls)
    Equal(t, all-delete, m.Count())
}

func Test_NumGroups(t *testing.T) {
    Equal(t, expected(0), numGroups(0))
    Equal(t, expected(1), numGroups(1))
    // max load factor 0.875
    Equal(t, expected(14), numGroups(14))
    Equal(t, expected(15), numGroups(15))
    Equal(t, expected(28), numGroups(28))
    Equal(t, expected(29), numGroups(29))
    Equal(t, expected(56), numGroups(56))
    Equal(t, expected(57), numGroups(57))
}

func expected(x int) (groups uint32) {
    groups = uint32(math.Ceil(float64(x) / float64(maxAvgGroupLoad)))
    if groups == 0 {
        groups = 1
    }
    return
}
