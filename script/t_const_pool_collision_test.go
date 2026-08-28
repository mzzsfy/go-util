package script

import (
	"testing"
)

// strHash31 复现旧实现的字符串累积哈希,用于构造并验证碰撞对
func strHash31(s string) uint64 {
	h := uint64(0)
	for _, ch := range s {
		h = h*31 + uint64(ch)
	}
	return h
}

// Test_constPool_StringCollision 脚本级验证:碰撞字符串不得共享常量
// Given 一对满足 h*31 累积哈希碰撞的字面量
// When 编译运行两者相等比较
// Then 结果为 false
func Test_constPool_StringCollision(t *testing.T) {
	result, err := Eval(`"Aa" == "BB"`)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if result.Bool() {
		t.Fatalf(`"Aa" 与 "BB" 哈希碰撞后被判相等, 得到 %v`, result)
	}
}

// Test_constPool_SameStringReusesCache 同串仍复用缓存路径
// Given 相同字符串两次加入常量池
// When 第二次查找缓存
// Then 命中且返回同一索引
func Test_constPool_SameStringReusesCache(t *testing.T) {
	c := NewCompiler()
	idx1 := c.addConstant(NewValue("Aa"))
	idx2 := c.addConstant(NewValue("Aa"))
	if idx1 != idx2 {
		t.Fatalf("相同字符串应复用常量, 得到 %d 与 %d", idx1, idx2)
	}

	// 脚本级:同一字面量出现两次仍应判等
	result, err := Eval(`"hello" == "hello"`)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !result.Bool() {
		t.Fatalf("相同字符串应判相等")
	}
}

// Test_constPool_HashCollisionPairs 碰撞对抽查:多对哈希碰撞串索引互不相同
func Test_constPool_HashCollisionPairs(t *testing.T) {
	// 碰撞构造规律: h(a,b)=a*31+b, 将首字符加1、次字符减31哈希不变
	pairs := [][2]string{
		{"Aa", "BB"},
		{"aX", "b9"},
		{"Zz", "[["},
	}
	for _, p := range pairs {
		if strHash31(p[0]) != strHash31(p[1]) {
			t.Fatalf("测试前置失败: %q 与 %q 哈希不同, 碰撞对构造错误", p[0], p[1])
		}
	}

	c := NewCompiler()
	for _, p := range pairs {
		idx1 := c.addConstant(NewValue(p[0]))
		idx2 := c.addConstant(NewValue(p[1]))
		if idx1 == idx2 {
			t.Fatalf("碰撞对 %q 与 %q 共享了常量索引 %d", p[0], p[1], idx1)
		}
	}
}
