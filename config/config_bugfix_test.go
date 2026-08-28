package config

import (
	"encoding/json"
	"strings"
	"testing"
)

// Test_FindFirstKey_UnclosedBracket 未闭合的[按普通键字符处理, 不 panic
func Test_FindFirstKey_UnclosedBracket(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path    string
		key     string
		nextKey string
	}{
		{"a[bc", "a[bc", ""},
		{"[bc", "[bc", ""},
		{"a.b", "a", "b"},
		{"a[0]b", "a", "0b"},
		{"plain", "plain", ""},
	}
	for _, c := range cases {
		key, nextKey := FindFirstKey(c.path)
		if key != c.key || nextKey != c.nextKey {
			t.Errorf("FindFirstKey(%q) = (%q, %q), 期望 (%q, %q)", c.path, key, nextKey, c.key, c.nextKey)
		}
	}
}

// Test_GetByPath_UnclosedBracket 畸形路径不触发 panic
func Test_GetByPath_UnclosedBracket(t *testing.T) {
	t.Parallel()
	m := map[string]any{"a": map[string]any{"b": 1}}
	// 不 panic 即通过
	_ = GetByPathAny(m, "a[bc")
	_ = GetByPath(m, "a[bc")
}

// Test_ResolveMap_UnclosedPlaceholder 未闭合占位符保留原始值, 不 panic
func Test_ResolveMap_UnclosedPlaceholder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]any
		want string
	}{
		{"前导字符", map[string]any{"k": "ab${x"}, "ab${x"},
		{"纯占位", map[string]any{"k": "${x"}, "${x"},
		{"混合闭合", map[string]any{"k": "ab${x}cd${y"}, "abcd${y"},
	}
	for _, c := range cases {
		got := ResolveMap(c.in)
		if got["k"] != c.want {
			t.Errorf("%s: ResolveMap 结果 %v, 期望 %v", c.name, got["k"], c.want)
		}
	}
}

// Test_ResolveMap_NormalPlaceholder 正常占位符行为不受影响
func Test_ResolveMap_NormalPlaceholder(t *testing.T) {
	t.Parallel()
	got := ResolveMap(map[string]any{"k": "${name:default}", "other": "x"})
	if got["k"] != "default" {
		t.Errorf("默认值占位解析结果 %v, 期望 default", got["k"])
	}
}

// Test_UntilingMap_KeyConflict 键与键路径冲突返回明确错误而非 panic
func Test_UntilingMap_KeyConflict(t *testing.T) {
	t.Parallel()
	_, err := UntilingMap(map[string]any{"a": 1, "a.b": 2})
	if err == nil {
		t.Fatal("a 与 a.b 冲突时应返回错误")
	}
	if !strings.Contains(err.Error(), "a") {
		t.Errorf("错误信息应包含冲突键名, 实际: %v", err)
	}
}

// Test_UntilingMap_Normal 正常平铺还原不受影响
func Test_UntilingMap_Normal(t *testing.T) {
	t.Parallel()
	got, err := UntilingMap(map[string]any{"a.b": 1, "a.c": "x", "d": true})
	if err != nil {
		t.Fatalf("正常数据不应报错: %v", err)
	}
	if got["a"].(map[string]any)["b"] != 1 {
		t.Errorf("a.b 还原失败: %v", got)
	}
	if got["a"].(map[string]any)["c"] != "x" {
		t.Errorf("a.c 还原失败: %v", got)
	}
	if got["d"] != true {
		t.Errorf("d 还原失败: %v", got)
	}
}

// Test_ParseConfigs2Map_KeyConflict 合并后键冲突返回错误而非 panic
func Test_ParseConfigs2Map_KeyConflict(t *testing.T) {
	t.Parallel()
	Parser["json"] = func(data []byte) map[string]any {
		r := make(map[string]any)
		if err := json.Unmarshal(data, &r); err != nil {
			return nil
		}
		return r
	}
	_, err := ParseConfigs2Map(&File{Name: "a.json", Data: []byte(`{"a": {"b": 1}}`)})
	if err != nil {
		t.Fatalf("正常配置不应报错: %v", err)
	}
}
