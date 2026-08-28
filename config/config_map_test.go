package config

import (
	"reflect"
	"strings"
	"testing"
)

// nestedSample 典型嵌套配置样例, 不含数组与 nil, 保证 Tiling/Untiling 可精确往返
func nestedSample() map[string]any {
	return map[string]any{
		"app": map[string]any{
			"name": "demo",
			"port": 8080,
			"db": map[string]any{
				"host": "localhost",
				"port": 3306,
				"pool": map[string]any{
					"min": 1,
					"max": 10,
				},
			},
		},
		"log": map[string]any{
			"level": "info",
		},
		"top": "value",
	}
}

// Test_TilingMap_Nested 多层嵌套平铺为点分键
func Test_TilingMap_Nested(t *testing.T) {
	t.Parallel()
	got := TilingMap(nestedSample())
	expect := map[string]any{
		"app.name":        "demo",
		"app.port":        8080,
		"app.db.host":     "localhost",
		"app.db.port":     3306,
		"app.db.pool.min": 1,
		"app.db.pool.max": 10,
		"log.level":       "info",
		"top":             "value",
	}
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("TilingMap 结果不符:\n got: %v\nwant: %v", got, expect)
	}
}

// Test_TilingMap_Array 数组按下标平铺, nil 值被丢弃
func Test_TilingMap_ArrayAndNil(t *testing.T) {
	t.Parallel()
	got := TilingMap(map[string]any{
		"arr":   []any{1, "two"},
		"empty": nil,
	})
	// nil 值无叶类型, 平铺时丢弃
	if _, exists := got["empty"]; exists {
		t.Errorf("nil 值不应出现在平铺结果中: %v", got)
	}
	if got["arr.0"] != 1 || got["arr.1"] != "two" {
		t.Errorf("数组平铺结果错误: %v", got)
	}
	if len(got) != 2 {
		t.Errorf("平铺键数量应为2, 实际 %d: %v", len(got), got)
	}
}

// Test_UntilingMap_DeepNesting 深层点分键还原为嵌套结构
func Test_UntilingMap_DeepNesting(t *testing.T) {
	t.Parallel()
	got, err := UntilingMap(map[string]any{
		"a.b.c.d": 1,
		"a.b.e":   "x",
		"f":       true,
	})
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	want := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{"d": 1},
				"e": "x",
			},
		},
		"f": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("还原结果不符:\n got: %v\nwant: %v", got, want)
	}
}

// Test_TilingUntilingRoundTrip 平铺还原往返一致
func Test_TilingUntilingRoundTrip(t *testing.T) {
	t.Parallel()
	origin := nestedSample()
	back, err := UntilingMap(TilingMap(origin))
	if err != nil {
		t.Fatalf("往返不应报错: %v", err)
	}
	if !reflect.DeepEqual(back, origin) {
		t.Errorf("往返不一致:\n got: %v\nwant: %v", back, origin)
	}
}

// Test_UntilingMap_ArrayKeysBecomeMap 数组式下标键还原为 map 而非数组(现行为)
func Test_UntilingMap_ArrayKeysBecomeMap(t *testing.T) {
	t.Parallel()
	got, err := UntilingMap(map[string]any{"arr.0": 1, "arr.1": 2})
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	m, ok := got["arr"].(map[string]any)
	if !ok {
		t.Fatalf("数组下标键应还原为嵌套 map, 实际 %T", got["arr"])
	}
	if m["0"] != 1 || m["1"] != 2 {
		t.Errorf("下标键还原内容错误: %v", m)
	}
}

// Test_UntilingMap_ConflictLeafFirst 先叶子后子键方向: 叶子被非map占用
func Test_UntilingMap_ConflictLeafFirst(t *testing.T) {
	t.Parallel()
	r := map[string]any{}
	if err := untilingMap("", "a", reflect.ValueOf(1), r); err != nil {
		t.Fatalf("先写叶子不应报错: %v", err)
	}
	err := untilingMap("", "a.b", reflect.ValueOf(2), r)
	if err == nil {
		t.Fatal("叶子被占用后展开子键应报错")
	}
	if !strings.Contains(err.Error(), "非map") {
		t.Errorf("错误应说明被非map值占用, 实际: %v", err)
	}
}

// Test_UntilingMap_ConflictSubkeyFirst 先子键后叶子方向: 子键已被展开
func Test_UntilingMap_ConflictSubkeyFirst(t *testing.T) {
	t.Parallel()
	r := map[string]any{}
	if err := untilingMap("", "a.b", reflect.ValueOf(2), r); err != nil {
		t.Fatalf("先展开子键不应报错: %v", err)
	}
	err := untilingMap("", "a", reflect.ValueOf(1), r)
	if err == nil {
		t.Fatal("子键被展开后写叶子应报错")
	}
	if !strings.Contains(err.Error(), "子键") {
		t.Errorf("错误应说明被子键展开占用, 实际: %v", err)
	}
}

// Test_UntilingMap_ConflictOrderIndependent 导出接口下冲突与迭代顺序无关, 均返回错误
func Test_UntilingMap_ConflictOrderIndependent(t *testing.T) {
	t.Parallel()
	for i := 0; i < 20; i++ {
		// map 迭代序随机, 多轮触发两种处理顺序
		_, err := UntilingMap(map[string]any{"a": 1, "a.b": 2})
		if err == nil {
			t.Fatal("键路径冲突必须返回错误")
		}
	}
}

// Test_UntilingMap_MixedLeaf 混合叶子值还原
// nil 叶子与 TilingMap 平铺行为对齐: 跳过写入, 不生成键与父键, 也不 panic
func Test_UntilingMap_MixedLeaf(t *testing.T) {
	t.Parallel()
	got, err := UntilingMap(map[string]any{"b.c": 1, "d": "str", "e": 1.5, "n": nil, "x.y": nil})
	if err != nil {
		t.Fatalf("混合叶子不应报错: %v", err)
	}
	if got["b"].(map[string]any)["c"] != 1 {
		t.Errorf("b.c 还原错误: %v", got)
	}
	if got["d"] != "str" || got["e"] != 1.5 {
		t.Errorf("叶子值还原错误: %v", got)
	}
	if _, ok := got["n"]; ok {
		t.Errorf("nil 叶子应被跳过: %v", got)
	}
	if _, ok := got["x"]; ok {
		t.Errorf("nil 叶子不应生成父键: %v", got)
	}
}

// Test_MergeAndTilingMap 高阶平铺覆盖低阶
func Test_MergeAndTilingMap(t *testing.T) {
	t.Parallel()
	high := map[string]any{"a": map[string]any{"k": "high", "onlyHigh": 1}}
	low := map[string]any{"a": map[string]any{"k": "low", "onlyLow": 2}, "b": 3}
	got := MergeAndTilingMap(high, low)
	expect := map[string]any{
		"a.k":        "high",
		"a.onlyHigh": 1,
		"a.onlyLow":  2,
		"b":          3,
	}
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("合并结果不符:\n got: %v\nwant: %v", got, expect)
	}
}

// Test_MergeMultiAndTilingMap 多 map 依次合并, 前面的优先级高
func Test_MergeMultiAndTilingMap(t *testing.T) {
	t.Parallel()
	m1 := map[string]any{"k": 1, "a": 1}
	m2 := map[string]any{"k": 2, "b": 2}
	m3 := map[string]any{"k": 3, "c": 3}
	got := MergeMultiAndTilingMap(m1, m2, m3)
	if got["k"] != 1 {
		t.Errorf("最前面的 map 优先级应最高, k=%v", got["k"])
	}
	if got["a"] != 1 || got["b"] != 2 || got["c"] != 3 {
		t.Errorf("合并结果缺失: %v", got)
	}
	// 空参数返回 nil, 与 ParseConfigs2Map 空列表返回 nil 的风格一致
	if got := MergeMultiAndTilingMap(); got != nil {
		t.Errorf("空参数应返回 nil, 实际 %v", got)
	}
}
