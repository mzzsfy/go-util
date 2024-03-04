package storage

//
//import "github.com/mzzsfy/go-util/unsafe"
//
//type swissMap1[K comparable, V any] struct {
//    ctrl     []metadata
//    groups   []group[K, V]
//    hash     unsafe.Hasher[K]
//    resident uint32
//    dead     uint32
//    limit    uint32
//}
//
//func (m *swissMap1[K, V]) Get(key K) (value V, ok bool) {
//    return m.GetWithHash(key, m.hash.Hash(key))
//}
//
//func (m *swissMap1[K, V]) Has(key K) (ok bool) {
//    return m.HasWithHash(key, m.hash.Hash(key))
//}
//
//func (m *swissMap1[K, V]) Delete(key K) (ok bool) {
//    return m.DeleteWithHash(key, m.hash.Hash(key))
//}
//
//func (m *swissMap1[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
//    return m.Iter(func(k K, v V) bool {
//        del, stop := cb(k, v)
//        if del {
//            m.Delete(k)
//        }
//        return stop
//    })
//}
//
//func (m *swissMap1[K, V]) Capacity() int {
//    return int(m.limit - m.resident)
//}
//
//func makeSwissMap1[K comparable, V any](size uint32) *swissMap1[K, V] {
//    groups := numGroups(size)
//    m := &swissMap1[K, V]{
//        ctrl:   make([]metadata, groups),
//        groups: make([]group[K, V], groups),
//        hash:   NewDefaultHasher[K](),
//        limit:  groups * maxAvgGroupLoad,
//    }
//    for i := range m.ctrl {
//        m.ctrl[i] = newEmptyMetadata()
//    }
//    return m
//}
//
//func (m *swissMap1[K, V]) HasWithHash(key K, hash uint64) (ok bool) {
//    hi, lo := splitHash(hash)
//    g := probeStart(hi, len(m.groups))
//    for { // inlined find loop
//        matches := metaMatchH2(&m.ctrl[g], lo)
//        for matches != 0 {
//            s := nextMatch(&matches)
//            if key == m.groups[g].keys[s] {
//                ok = true
//                return
//            }
//        }
//        // |key| is not in group |g|,
//        // stop probing if we see an empty slot
//        matches = metaMatchEmpty(&m.ctrl[g])
//        if matches != 0 {
//            ok = false
//            return
//        }
//        g++ // linear probing
//        if g >= uint32(len(m.groups)) {
//            g = 0
//        }
//    }
//}
//
//func (m *swissMap1[K, V]) GetWithHash(key K, hash uint64) (value V, ok bool) {
//    hi, lo := splitHash(hash)
//    g := probeStart(hi, len(m.groups))
//    for { // inlined find loop
//        matches := metaMatchH2(&m.ctrl[g], lo)
//        for matches != 0 {
//            s := nextMatch(&matches)
//            if key == m.groups[g].keys[s] {
//                value, ok = m.groups[g].values[s], true
//                return
//            }
//        }
//        // |key| is not in group |g|,
//        // stop probing if we see an empty slot
//        matches = metaMatchEmpty(&m.ctrl[g])
//        if matches != 0 {
//            ok = false
//            return
//        }
//        g++ // linear probing
//        if g >= uint32(len(m.groups)) {
//            g = 0
//        }
//    }
//}
//
//func (m *swissMap1[K, V]) Put(key K, value V) {
//    m.putWithHash(key, value, m.hash.Hash(key))
//}
//
//func (m *swissMap1[K, V]) putWithHash(key K, value V, hash uint64) {
//    if m.resident >= m.limit {
//        m.rehash(m.nextSize())
//    }
//    hi, lo := splitHash(hash)
//    g := probeStart(hi, len(m.groups))
//    for { // inlined find loop
//        matches := metaMatchH2(&m.ctrl[g], lo)
//        for matches != 0 {
//            s := nextMatch(&matches)
//            if key == m.groups[g].keys[s] { // update
//                m.groups[g].keys[s] = key
//                m.groups[g].values[s] = value
//                return
//            }
//        }
//        // |key| is not in group |g|,
//        // stop probing if we see an empty slot
//        matches = metaMatchEmpty(&m.ctrl[g])
//        if matches != 0 { // insert
//            s := nextMatch(&matches)
//            m.groups[g].keys[s] = key
//            m.groups[g].values[s] = value
//            m.ctrl[g][s] = int8(lo)
//            m.resident++
//            return
//        }
//        g++ // linear probing
//        if g >= uint32(len(m.groups)) {
//            g = 0
//        }
//    }
//}
//
//func (m *swissMap1[K, V]) DeleteWithHash(key K, hash uint64) (ok bool) {
//    hi, lo := splitHash(hash)
//    g := probeStart(hi, len(m.groups))
//    for {
//        matches := metaMatchH2(&m.ctrl[g], lo)
//        for matches != 0 {
//            s := nextMatch(&matches)
//            if key == m.groups[g].keys[s] {
//                ok = true
//                // optimization: if |m.ctrl[g]| contains any empty
//                // metadata bytes, we can physically delete |key|
//                // rather than placing a tombstone.
//                // The observation is that any probes into group |g|
//                // would already be terminated by the existing empty
//                // slot, and therefore reclaiming slot |s| will not
//                // cause premature termination of probes into |g|.
//                if metaMatchEmpty(&m.ctrl[g]) != 0 {
//                    m.ctrl[g][s] = empty
//                    m.resident--
//                } else {
//                    m.ctrl[g][s] = tombstone
//                    m.dead++
//                }
//                var k K
//                var v V
//                m.groups[g].keys[s] = k
//                m.groups[g].values[s] = v
//                return
//            }
//        }
//        // |key| is not in group |g|,
//        // stop probing if we see an empty slot
//        matches = metaMatchEmpty(&m.ctrl[g])
//        if matches != 0 { // |key| absent
//            ok = false
//            return
//        }
//        g++ // linear probing
//        if g >= uint32(len(m.groups)) {
//            g = 0
//        }
//    }
//}
//
//func (m *swissMap1[K, V]) Clean() {
//    for i, c := range m.ctrl {
//        for j := range c {
//            m.ctrl[i][j] = empty
//        }
//    }
//    var k K
//    var v V
//    for i := range m.groups {
//        g := &m.groups[i]
//        for i := range g.keys {
//            g.keys[i] = k
//            g.values[i] = v
//        }
//    }
//    m.resident, m.dead = 0, 0
//}
//
////nolint:gosec
//func (m *swissMap1[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
//    // take a consistent view of the table in case
//    // we rehash during iteration
//    ctrl, groups := m.ctrl, m.groups
//    // pick a random starting group
//    g := randIntN(len(groups))
//    for n := 0; n < len(groups); n++ {
//        for s, c := range ctrl[g] {
//            if c == empty || c == tombstone {
//                continue
//            }
//            k, v := groups[g].keys[s], groups[g].values[s]
//            if stop := cb(k, v); stop {
//                return stop
//            }
//        }
//        g++
//        if g >= uint32(len(groups)) {
//            g = 0
//        }
//    }
//    return false
//}
//
//func (m *swissMap1[K, V]) Count() int {
//    return int(m.resident - m.dead)
//}
//
//func (m *swissMap1[K, V]) nextSize() (n uint32) {
//    n = uint32(len(m.groups)) * 2
//    if m.dead >= (m.resident / 2) {
//        n = uint32(len(m.groups))
//    }
//    return
//}
//
//func (m *swissMap1[K, V]) rehash(n uint32) {
//    groups, ctrl := m.groups, m.ctrl
//    m.groups = make([]group[K, V], n)
//    m.ctrl = make([]metadata, n)
//    for i := range m.ctrl {
//        m.ctrl[i] = newEmptyMetadata()
//    }
//    m.hash = m.hash.NewSeed()
//    m.limit = n * maxAvgGroupLoad
//    m.resident, m.dead = 0, 0
//    for g := range ctrl {
//        for s := range ctrl[g] {
//            c := ctrl[g][s]
//            if c == empty || c == tombstone {
//                continue
//            }
//            m.Put(groups[g].keys[s], groups[g].values[s])
//        }
//    }
//}
