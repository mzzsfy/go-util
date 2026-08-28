package config

import (
	"encoding/json"
	"fmt"
	"testing"
)

// benchConfig 组数, 每组 3 子组, 每子组 3 叶, 共 90 个叶子键
const (
	benchGroups    = 10
	benchSubs      = 3
	benchLeaves    = 3
	benchLeafCount = benchGroups * benchSubs * benchLeaves
)

// makeBenchConfig 生成三层嵌套的典型配置, 叶子键数量在 50-100 量级
func makeBenchConfig() map[string]any {
	m := make(map[string]any, benchGroups)
	for g := 0; g < benchGroups; g++ {
		sub := make(map[string]any, benchSubs)
		for s := 0; s < benchSubs; s++ {
			leaf := make(map[string]any, benchLeaves)
			for l := 0; l < benchLeaves; l++ {
				leaf[fmt.Sprintf("k%d", l)] = l
			}
			sub[fmt.Sprintf("s%d", s)] = leaf
		}
		m[fmt.Sprintf("g%d", g)] = sub
	}
	return m
}

// makeBenchTiled 平铺态基准输入, 注入链式引用与默认值两类占位符
func makeBenchTiled() map[string]any {
	tiled := TilingMap(makeBenchConfig())
	for g := 0; g < benchGroups; g++ {
		// k0 引用 k1, k1 引用缺失键的默认值, 解析需多轮迭代收敛
		tiled[fmt.Sprintf("g%d.s0.k0", g)] = fmt.Sprintf("ref-${g%d.s0.k1}", g)
		tiled[fmt.Sprintf("g%d.s0.k1", g)] = "${missing:dft}"
	}
	return tiled
}

func BenchmarkTilingMap(b *testing.B) {
	b.ReportAllocs()
	src := makeBenchConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TilingMap(src)
	}
}

func BenchmarkUntilingMap(b *testing.B) {
	b.ReportAllocs()
	src := TilingMap(makeBenchConfig())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := UntilingMap(src); err != nil {
			b.Fatalf("UntilingMap: %v", err)
		}
	}
}

// BenchmarkResolveMap 占位符解析, 输入含引用与默认值两类占位符
// ResolveMap 原地修改输入, 每次迭代重建输入, 重建成本计入结果
func BenchmarkResolveMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ResolveMap(makeBenchTiled())
	}
}

// BenchmarkParseConfigs2Map 端到端: 双文件解析合并含占位符
func BenchmarkParseConfigs2Map(b *testing.B) {
	registerJSONParser()
	doc := makeBenchConfig()
	data, err := json.Marshal(doc)
	if err != nil {
		b.Fatalf("序列化基准配置: %v", err)
	}
	high := &File{Name: "bench-high.json", Order: 10, Data: data}
	lowDoc, err := UntilingMap(TilingMap(makeBenchConfig()))
	if err != nil {
		b.Fatalf("构造低阶配置: %v", err)
	}
	// 增加一个跨文件占位符引用
	lowDoc["bench"] = map[string]any{"extra": "${g0.s0.k0}"}
	lowData, err := json.Marshal(lowDoc)
	if err != nil {
		b.Fatalf("序列化低阶配置: %v", err)
	}
	low := &File{Name: "bench-low.json", Order: 1, Data: lowData}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ParseConfigs2Map(high, low); err != nil {
			b.Fatalf("ParseConfigs2Map: %v", err)
		}
	}
}
