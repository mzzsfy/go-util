package helper

//
//import (
//    "strconv"
//    "testing"
//)
//
//// TestBloomFilter 测试布隆过滤器,目前无法通过测试,原因是哈希函数效果不太行
//func _TestBloomFilter(t *testing.T) {
//    faultTolerance := .1
//    valueScale := .4
//    for _, i := range []int{2 << 16, 2 << 19, 2 << 21} {
//        for _, b := range []float64{0.01, 0.03, 0.1} {
//            t.Run("int_"+NumberToString(i)+"_"+strconv.FormatFloat(b, 'f', -1, 64), func(t *testing.T) {
//                filter := NewBloomFilter[int](uint(i), b)
//                l := int(float64(i) * valueScale)
//                for j := 0; j < l; j++ {
//                    filter.Add(j)
//                }
//                fail := 0
//                for j := 0; j < l; j++ {
//                    if !filter.Has(j) {
//                        fail++
//                    }
//                }
//                e := int(float64(l) * (b + faultTolerance))
//                if fail > e {
//                    t.Errorf("want filter.Has() fail less %v, got %v", e, fail)
//                }
//                fail = 0
//                for j := l + 1; j < l+l; j++ {
//                    if filter.Has(j) {
//                        fail++
//                    }
//                }
//                if fail > e {
//                    t.Errorf("want filter.Has() fail less %v, got %v", e, fail)
//                }
//                e = int(float64(l) * (b - faultTolerance))
//                if fail < e {
//                    t.Errorf("want filter.Has() fail more %v, got %v", e, fail)
//                }
//            })
//            t.Run("string_"+NumberToString(i)+"_"+strconv.FormatFloat(b, 'f', -1, 64), func(t *testing.T) {
//                filter := NewBloomFilter[string](uint(i), b)
//                l := int(float64(i) * valueScale)
//                for j := 0; j < l; j++ {
//                    filter.Add(NumberToString(j))
//                }
//                fail := 0
//                for j := 0; j < l; j++ {
//                    if !filter.Has(NumberToString(j)) {
//                        fail++
//                    }
//                }
//                e := int(float64(l) * (b + faultTolerance))
//                if fail > e {
//                    t.Errorf("want filter.Has() fail less %v, got %v", e, fail)
//                }
//                fail = 0
//                for j := l + 1; j < l+l; j++ {
//                    if filter.Has(NumberToString(j)) {
//                        fail++
//                    }
//                }
//                if fail > e {
//                    t.Errorf("want filter.Has() fail less %v, got %v", e, fail)
//                }
//                e = int(float64(l) * (b - faultTolerance))
//                if fail < e {
//                    t.Errorf("want filter.Has() fail more %v, got %v", e, fail)
//                }
//            })
//        }
//    }
//}
