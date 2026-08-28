package config

import (
	"reflect"
	"testing"
)

// Test_ResolveMap_PlaceholderForms 占位符基础形态
func Test_ResolveMap_PlaceholderForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   map[string]any
		key  string
		want any
	}{
		{"引用存在的键", map[string]any{"k": "${name}", "name": "v"}, "k", "v"},
		{"引用不存在的键替换为空", map[string]any{"k": "${missing}"}, "k", ""},
		{"默认值在键缺失时生效", map[string]any{"k": "${missing:fallback}"}, "k", "fallback"},
		{"默认值在键存在时被覆盖", map[string]any{"k": "${name:fallback}", "name": "real"}, "k", "real"},
		{"键为空串时使用默认值", map[string]any{"k": "${name:fallback}", "name": ""}, "k", "fallback"},
		{"带前后缀文本", map[string]any{"k": "pre-${name}-post", "name": "v"}, "k", "pre-v-post"},
		{"单字符串多占位符", map[string]any{"k": "${a}-${b}", "a": "x", "b": "y"}, "k", "x-y"},
		{"引用整数值", map[string]any{"k": "n=${n}", "n": 5}, "k", "n=5"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveMap(tc.in)
			if got[tc.key] != tc.want {
				t.Errorf("结果 %v, 期望 %v", got[tc.key], tc.want)
			}
		})
	}
}

// Test_ResolveMap_ChainedRefs 链式引用经多轮迭代收敛
func Test_ResolveMap_ChainedRefs(t *testing.T) {
	t.Parallel()
	got := ResolveMap(map[string]any{
		"a": "${b}",
		"b": "${c}",
		"c": "deep",
	})
	for k, want := range map[string]any{"a": "deep", "b": "deep", "c": "deep"} {
		if got[k] != want {
			t.Errorf("%s = %v, 期望 %v", k, got[k], want)
		}
	}
}

// Test_ResolveMap_ChainedRefsWithDefault 链式引用末端走默认值
func Test_ResolveMap_ChainedRefsWithDefault(t *testing.T) {
	t.Parallel()
	got := ResolveMap(map[string]any{
		"a": "${b}",
		"b": "${missing:dft}",
	})
	if got["a"] != "dft" || got["b"] != "dft" {
		t.Errorf("链式默认值未收敛: a=%v, b=%v", got["a"], got["b"])
	}
}

// Test_ResolveMap_SelfReference 递归引用自身时使用默认值终止
func Test_ResolveMap_SelfReference(t *testing.T) {
	t.Parallel()
	got := ResolveMap(map[string]any{
		"a": "${a:dft}",
	})
	if got["a"] != "dft" {
		t.Errorf("自引用应取默认值, 实际 %v", got["a"])
	}
}

// Test_ResolveMap_NonStringUntouched 非字符串值保持原样
func Test_ResolveMap_NonStringUntouched(t *testing.T) {
	t.Parallel()
	in := map[string]any{
		"i": 1,
		"f": 1.5,
		"b": true,
		"m": map[string]any{"x": "${y}"},
	}
	got := ResolveMap(in)
	if got["i"] != 1 || got["f"] != 1.5 || got["b"] != true {
		t.Errorf("标量值不应被改动: %v", got)
	}
	// 嵌套 map 不属于单层结构, 不做解析
	if !reflect.DeepEqual(got["m"], map[string]any{"x": "${y}"}) {
		t.Errorf("嵌套 map 不应被解析: %v", got["m"])
	}
}

// Test_ResolveMap_PartialUnresolvable 部分占位符不可解析时保留占位原文
func Test_ResolveMap_PartialUnresolvable(t *testing.T) {
	t.Parallel()
	got := ResolveMap(map[string]any{"k": "a${ok}b${noSuchKey]tail", "ok": "v"})
	if got["k"] != "avb${noSuchKey]tail" {
		t.Errorf("可解析部分应替换, 其余保留原文, 实际 %v", got["k"])
	}
}
