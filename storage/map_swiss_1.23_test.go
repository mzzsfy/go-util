//go:build !go1.24

package storage

import (
    "math"
    "math/rand"
    "testing"
)

func Equal(t *testing.T, a, b interface{}) {
    t.Helper()
    if a != b {
        t.Errorf("expected %v, got %v", a, b)
    }
}
func True(t *testing.T, a bool) {
    t.Helper()
    if !a {
        t.Errorf("expected true, got false")
    }
}

func Test_SwissMap(t *testing.T) {
    t.Run("strings=1", func(t *testing.T) {
        testSwissMap(t, genStringData(16, 1))
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
    // capacity() behavior depends on |groupSize|
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
        Equal(t, int(c), m1.capacity())
        keys := gen(rand.Intn(int(c)))
        for _, k := range keys {
            m.Put(k, k)
        }
        Equal(t, int(c)-len(keys), m1.capacity())
        Equal(t, int(c), m.Count()+m1.capacity())
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

//go test -run='^\QTestBenchmarkSwissMap\E$/^\Qswiss_w\E$' -cpuprofile=cpu.pprof ./storage
func _TestBenchmarkSwissMap(t *testing.T) {
    n := 300_000
    n1 := 1_000_000
    t.Run("swiss_r", func(t *testing.T) {
        m := makeSwissMap[int, int](1)
        for i := 0; i < n1; i++ {
            m.Put(i, i)
        }
        v := 1000 * n
        ints := make([]int, n1)
        for i := 0; i < n1; i++ {
            ints[i] = rand.Intn(math.MaxInt) % n1
        }
        for i := 0; i < v; i++ {
            m.Get(ints[i%n1])
        }
    })

    t.Run("swiss_w", func(t *testing.T) {
        m := makeSwissMap[int, int](1)
        v := 1000 * n
        ints := make([]int, n1)
        for i := 0; i < n1; i++ {
            ints[i] = rand.Intn(math.MaxInt) % n1
        }
        for i := 0; i < v; i++ {
            m.Put(ints[i%n1], i)
        }
    })

    t.Run("swiss_rw", func(t *testing.T) {
        m := makeSwissMap[int, int](1)
        v := 1000 * n
        ints := make([]int, n1)
        for i := 0; i < n1; i++ {
            ints[i] = rand.Intn(math.MaxInt) % n1
        }
        for i := 0; i < v; i++ {
            if i%2 == 0 {
                m.Get(ints[i%n1])
            } else {
                m.Put(ints[i%n1], i)
            }
        }
    })
}
