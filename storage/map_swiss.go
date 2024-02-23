package storage

import (
    "github.com/mzzsfy/go-util/unsafe"
    _ "unsafe"
)

func NewDefaultHasher[K comparable]() unsafe.Hasher[K] {
    return unsafe.NewHasher[K]()
}

//go:linkname fastrand runtime.fastrand
func fastrand() uint32

// swissMap is an open-addressing hash map
// based on Abseil's flat_hash_map.
type swissMap[K comparable, V any] struct {
    ctrl     []metadata
    groups   []group[K, V]
    hash     unsafe.Hasher[K]
    resident uint32
    dead     uint32
    limit    uint32
}

func (m *swissMap[K, V]) Get(key K) (V, bool) {
    return m.GetWithHash(key, m.hash.Hash(key))
}

func (m *swissMap[K, V]) GetSimple(key K) (value V) {
    value, _ = m.GetWithHash(key, m.hash.Hash(key))
    return
}

func (m *swissMap[K, V]) Has(key K) bool {
    return m.HasWithHash(key, m.hash.Hash(key))
}

func (m *swissMap[K, V]) Delete(key K) {
    m.DeleteWithHash(key, m.hash.Hash(key))
}

func (m *swissMap[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
    return m.Iter(func(k K, v V) bool {
        del, stop := cb(k, v)
        if del {
            m.Delete(k)
        }
        return stop
    })
}

// Capacity returns the number of additional elements
// the can be added to the Map before resizing.
func (m *swissMap[K, V]) Capacity() int {
    return int(m.limit - m.resident)
}

// metadata is the loByte metadata array for a group.
// find operations first probe the controls bytes
// to filter candidates before matching keys
type metadata [groupSize]int8

// group is a group of 16 key-value pairs
type group[K comparable, V any] struct {
    keys   [groupSize]K
    values [groupSize]V
}

const (
    h2Mask    uint64 = 0x0000_0000_0000_007f
    empty     int8   = -128 // 0b1000_0000
    tombstone int8   = -2   // 0b1111_1110
)

// hiByte is a 57 bit hash prefix
type hiByte uint64

// loByte is a 7 bit hash suffix
type loByte int8

func makeSwissMap[K comparable, V any](size uint32) *swissMap[K, V] {
    groups := numGroups(size)
    m := &swissMap[K, V]{
        ctrl:   make([]metadata, groups),
        groups: make([]group[K, V], groups),
        hash:   NewDefaultHasher[K](),
        limit:  groups * maxAvgGroupLoad,
    }
    for i := range m.ctrl {
        m.ctrl[i] = newEmptyMetadata()
    }
    return m
}

func MapTypeSwiss[K comparable, V any](size uint32) MakeMap[K, V] {
    return MapImpl[K, V](func() Map[K, V] { return makeSwissMap[K, V](size) })
}

func (m *swissMap[K, V]) HasWithHash(key K, hash uint64) (ok bool) {
    hi, lo := splitHash(hash)
    g := probeStart(hi, len(m.groups))
    for { // inlined find loop
        matches := metaMatchH2(&m.ctrl[g], lo)
        for matches != 0 {
            s := nextMatch(&matches)
            if key == m.groups[g].keys[s] {
                ok = true
                return
            }
        }
        // |key| is not in group |g|,
        // stop probing if we see an empty slot
        matches = metaMatchEmpty(&m.ctrl[g])
        if matches != 0 {
            ok = false
            return
        }
        g++ // linear probing
        if g >= uint32(len(m.groups)) {
            g = 0
        }
    }
}

func (m *swissMap[K, V]) GetWithHash(key K, hash uint64) (value V, ok bool) {
    hi, lo := splitHash(hash)
    g := probeStart(hi, len(m.groups))
    for { // inlined find loop
        matches := metaMatchH2(&m.ctrl[g], lo)
        for matches != 0 {
            s := nextMatch(&matches)
            if key == m.groups[g].keys[s] {
                value, ok = m.groups[g].values[s], true
                return
            }
        }
        // |key| is not in group |g|,
        // stop probing if we see an empty slot
        matches = metaMatchEmpty(&m.ctrl[g])
        if matches != 0 {
            ok = false
            return
        }
        g++ // linear probing
        if g >= uint32(len(m.groups)) {
            g = 0
        }
    }
}

// Put attempts to insert |key| and |value|
func (m *swissMap[K, V]) Put(key K, value V) {
    m.PutWithHash(key, value, m.hash.Hash(key))
}

// PutWithHash attempts to insert |key| and |value|
func (m *swissMap[K, V]) PutWithHash(key K, value V, hash uint64) {
    if m.resident >= m.limit {
        m.rehash(m.nextSize())
    }
    hi, lo := splitHash(hash)
    g := probeStart(hi, len(m.groups))
    for { // inlined find loop
        matches := metaMatchH2(&m.ctrl[g], lo)
        for matches != 0 {
            s := nextMatch(&matches)
            if key == m.groups[g].keys[s] { // update
                m.groups[g].keys[s] = key
                m.groups[g].values[s] = value
                return
            }
        }
        // |key| is not in group |g|,
        // stop probing if we see an empty slot
        matches = metaMatchEmpty(&m.ctrl[g])
        if matches != 0 { // insert
            s := nextMatch(&matches)
            m.groups[g].keys[s] = key
            m.groups[g].values[s] = value
            m.ctrl[g][s] = int8(lo)
            m.resident++
            return
        }
        g++ // linear probing
        if g >= uint32(len(m.groups)) {
            g = 0
        }
    }
}

func (m *swissMap[K, V]) DeleteWithHash(key K, hash uint64) (ok bool) {
    hi, lo := splitHash(hash)
    g := probeStart(hi, len(m.groups))
    for {
        matches := metaMatchH2(&m.ctrl[g], lo)
        for matches != 0 {
            s := nextMatch(&matches)
            if key == m.groups[g].keys[s] {
                ok = true
                // optimization: if |m.ctrl[g]| contains any empty
                // metadata bytes, we can physically delete |key|
                // rather than placing a tombstone.
                // The observation is that any probes into group |g|
                // would already be terminated by the existing empty
                // slot, and therefore reclaiming slot |s| will not
                // cause premature termination of probes into |g|.
                if metaMatchEmpty(&m.ctrl[g]) != 0 {
                    m.ctrl[g][s] = empty
                    m.resident--
                } else {
                    m.ctrl[g][s] = tombstone
                    m.dead++
                }
                var k K
                var v V
                m.groups[g].keys[s] = k
                m.groups[g].values[s] = v
                return
            }
        }
        // |key| is not in group |g|,
        // stop probing if we see an empty slot
        matches = metaMatchEmpty(&m.ctrl[g])
        if matches != 0 { // |key| absent
            ok = false
            return
        }
        g++ // linear probing
        if g >= uint32(len(m.groups)) {
            g = 0
        }
    }
}

// Clear removes all elements from the swissMap1.
func (m *swissMap[K, V]) Clear() {
    for i, c := range m.ctrl {
        for j := range c {
            m.ctrl[i][j] = empty
        }
    }
    var k K
    var v V
    for i := range m.groups {
        g := &m.groups[i]
        for i := range g.keys {
            g.keys[i] = k
            g.values[i] = v
        }
    }
    m.resident, m.dead = 0, 0
}

// Iter iterates the elements of the swissMap1, passing them to the callback.
// It guarantees that any key in the swissMap1 will be visited only once, and
// for un-mutated Maps, every key will be visited once. If the swissMap1 is
// Mutated during iteration, mutations will be reflected on return from
// Iter, but the set of keys visited by Iter is non-deterministic.
//
//nolint:gosec
func (m *swissMap[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
    // take a consistent view of the table in case
    // we rehash during iteration
    ctrl, groups := m.ctrl, m.groups
    // pick a random starting group
    g := randIntN(len(groups))
    for n := 0; n < len(groups); n++ {
        for s, c := range ctrl[g] {
            if c == empty || c == tombstone {
                continue
            }
            k, v := groups[g].keys[s], groups[g].values[s]
            if stop := cb(k, v); stop {
                return stop
            }
        }
        g++
        if g >= uint32(len(groups)) {
            g = 0
        }
    }
    return false
}

// Count returns the number of elements in the swissMap1.
func (m *swissMap[K, V]) Count() int {
    return int(m.resident - m.dead)
}

func (m *swissMap[K, V]) nextSize() (n uint32) {
    n = uint32(len(m.groups)) * 2
    if m.dead >= (m.resident / 2) {
        n = uint32(len(m.groups))
    }
    return
}

func (m *swissMap[K, V]) rehash(n uint32) {
    groups, ctrl := m.groups, m.ctrl
    m.groups = make([]group[K, V], n)
    m.ctrl = make([]metadata, n)
    for i := range m.ctrl {
        m.ctrl[i] = newEmptyMetadata()
    }
    m.hash = m.hash.NewSeed()
    m.limit = n * maxAvgGroupLoad
    m.resident, m.dead = 0, 0
    for g := range ctrl {
        for s := range ctrl[g] {
            c := ctrl[g][s]
            if c == empty || c == tombstone {
                continue
            }
            m.Put(groups[g].keys[s], groups[g].values[s])
        }
    }
}

// numGroups returns the minimum number of groups needed to store |n| elems.
func numGroups(n uint32) (groups uint32) {
    groups = (n + maxAvgGroupLoad - 1) / maxAvgGroupLoad
    if groups == 0 {
        groups = 1
    }
    return
}

func newEmptyMetadata() (meta metadata) {
    for i := range meta {
        meta[i] = empty
    }
    return
}

func splitHash(h uint64) (hiByte, loByte) {
    return hiByte(h >> 7), loByte(h & h2Mask)
}

func probeStart(hi hiByte, groups int) uint32 {
    return fastModN(uint32(hi), uint32(groups))
}

// lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
func fastModN(x, n uint32) uint32 {
    return uint32((uint64(x) * uint64(n)) >> 32)
}

// randIntN returns a random number in the interval [0, n).
func randIntN(n int) uint32 {
    return fastModN(fastrand(), uint32(n))
}
