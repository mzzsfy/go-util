// Package seq 提供高性能的泛型链式操作库
//
// 设计灵感来源于Java Stream API，支持并行处理、过滤、映射、排序和聚合操作
// 在热路径中尽可能实现零内存分配
//
// 泛型方法 API 需要 go1.27+;更早版本编译旧版 any 形态 API(*_pre127.go)，两轨源码不兼容，跨版本须按轨改写
//
// 基本用法:
//
//	seq.FromSlice([]int{1, 2, 3}).
//	    Filter(func(i int) bool { return i > 1 }).
//	    Map(func(i int) int { return i * 2 }).
//	    ToSlice()
package seq
