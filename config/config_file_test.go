package config

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
)

// registerJSONParser 仅注册一次 json 解析器, 调用方须为非并行测试
var registerJSONParserOnce sync.Once

func registerJSONParser() {
	registerJSONParserOnce.Do(func() {
		Parser["json"] = func(data []byte) map[string]any {
			r := make(map[string]any)
			if err := json.Unmarshal(data, &r); err != nil {
				return nil
			}
			return r
		}
	})
}

// Test_ParseConfigs2Map_NoFiles 空文件列表返回错误
func Test_ParseConfigs2Map_NoFiles(t *testing.T) {
	registerJSONParser()
	m, err := ParseConfigs2Map()
	if err == nil {
		t.Fatal("空文件列表应返回错误")
	}
	if m != nil {
		t.Errorf("空文件列表不应返回数据, 实际 %v", m)
	}
}

// Test_ParseConfigs2Map_OrderPriority 多文件同名键的合并优先级
// Order 高的覆盖 Order 低的, 与 Spring 式 profile 语义一致
func Test_ParseConfigs2Map_OrderPriority(t *testing.T) {
	registerJSONParser()
	high := &File{Name: "high.json", Order: 10, Data: []byte(`{"a": "fromHigh", "onlyHigh": 1}`)}
	low := &File{Name: "low.json", Order: 1, Data: []byte(`{"a": "fromLow", "b": 2}`)}
	got, err := ParseConfigs2Map(high, low)
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	if got["a"] != "fromHigh" {
		t.Errorf("高 Order 应覆盖低 Order, a=%v", got["a"])
	}
	// JSON 解码数字为 float64
	if got["b"] != float64(2) || got["onlyHigh"] != float64(1) {
		t.Errorf("不同键应全部保留: %v", got)
	}
}

// Test_ParseConfigs2Map_PlaceholderAcrossFiles 占位符可引用其他文件的高优先级值
func Test_ParseConfigs2Map_PlaceholderAcrossFiles(t *testing.T) {
	registerJSONParser()
	high := &File{Name: "high.json", Order: 10, Data: []byte(`{"host": "real-host"}`)}
	low := &File{Name: "low.json", Order: 1, Data: []byte(`{"url": "${host}:80", "fallback": "${missing:fb}"}`)}
	got, err := ParseConfigs2Map(high, low)
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	if got["url"] != "real-host:80" {
		t.Errorf("跨文件占位符应解析为高 Order 值, url=%v", got["url"])
	}
	if got["fallback"] != "fb" {
		t.Errorf("缺失键应取默认值, fallback=%v", got["fallback"])
	}
}

// Test_ParseConfig_UnregisteredType 未注册的文件类型返回 nil
func Test_ParseConfig_UnregisteredType(t *testing.T) {
	registerJSONParser()
	if got := ParseConfig(&File{Name: "a.unknownext", Data: []byte("x")}); got != nil {
		t.Errorf("未注册类型应返回 nil, 实际 %v", got)
	}
	if got := ParseConfig(&File{Name: "noext", Data: []byte("x")}); got != nil {
		t.Errorf("无后缀应返回 nil, 实际 %v", got)
	}
}

// Test_FileTypeHandler_Custom 自定义类型处理器注册后可解析
func Test_FileTypeHandler_Custom(t *testing.T) {
	registerJSONParser()
	origLen := len(FileTypeHandler)
	FileTypeHandler = append(FileTypeHandler, func(name string) string {
		if name == "special" {
			return "json"
		}
		return ""
	})
	defer func() {
		FileTypeHandler = FileTypeHandler[:origLen]
	}()
	got := ParseConfig(&File{Name: "special", Data: []byte(`{"k": "v"}`)})
	if got == nil || got["k"] != "v" {
		t.Errorf("自定义处理器应生效: %v", got)
	}
}

// Test_ParseConfigs_InputSortedByOrder 解析按 Order 降序排列
func Test_ParseConfigs_InputSortedByOrder(t *testing.T) {
	registerJSONParser()
	files := []*File{
		{Name: "a.json", Order: 1, Data: []byte(`{"src": "a"}`)},
		{Name: "b.json", Order: 5, Data: []byte(`{"src": "b"}`)},
		{Name: "c.json", Order: 3, Data: []byte(`{"src": "c"}`)},
	}
	docs := ParseConfigs(files)
	if len(docs) != 3 {
		t.Fatalf("应解析出3份配置, 实际 %d", len(docs))
	}
	if docs[0]["src"] != "b" || docs[1]["src"] != "c" || docs[2]["src"] != "a" {
		t.Errorf("应按 Order 降序输出: %v, %v, %v", docs[0]["src"], docs[1]["src"], docs[2]["src"])
	}
}

// Test_EnvMap 环境变量转 map, 含值中等号切分语义
func Test_EnvMap(t *testing.T) {
	t.Setenv("GO_UTIL_B5_TEST_KEY", "a=b=c")
	env := EnvMap()
	if len(env) == 0 {
		t.Fatal("环境变量 map 不应为空")
	}
	for k, v := range env {
		if k == "" {
			t.Error("环境变量键不应为空")
		}
		if _, ok := v.(string); !ok {
			t.Errorf("环境变量值应为 string, 键 %s 实际 %T", k, v)
		}
	}
	if got := env["GO_UTIL_B5_TEST_KEY"]; got != "a=b=c" {
		t.Errorf("值中首个等号后内容应完整保留, 实际 %v", got)
	}
}

// Test_EnvMap_ContainsPath 常见系统变量存在
func Test_EnvMap_ContainsPath(t *testing.T) {
	env := EnvMap()
	// 跨平台系统均提供 PATH 类变量(Windows 为 Path, Unix 为 PATH)
	_, hasPath := env["PATH"]
	_, hasPathWin := env["Path"]
	if !hasPath && !hasPathWin {
		t.Errorf("应包含 PATH/Path 变量, 实际键数量 %d", len(env))
	}
}

// Test_ParseConfig_JsonDecodeFailure 解析失败返回 nil 不 panic
func Test_ParseConfig_JsonDecodeFailure(t *testing.T) {
	registerJSONParser()
	if got := ParseConfig(&File{Name: "bad.json", Data: []byte("{invalid")}); got != nil {
		t.Errorf("非法 JSON 应返回 nil, 实际 %v", got)
	}
}

// Test_TilingUntiling_JsonRoundTrip JSON 解析后平铺还原保持键值集合
func Test_TilingUntiling_JsonRoundTrip(t *testing.T) {
	registerJSONParser()
	doc := ParseConfig(&File{Name: "r.json", Data: []byte(`{"a": {"b": 1, "c": [true, false]}, "d": "x"}`)})
	if doc == nil {
		t.Fatal("解析不应为 nil")
	}
	tiled := TilingMap(doc)
	// 数组被平铺为下标键, 验证关键路径存在(JSON 数字解码为 float64)
	if tiled["a.b"] != float64(1) || tiled["d"] != "x" {
		t.Errorf("平铺结果缺失: %v", tiled)
	}
	if tiled["a.c.0"] != true || tiled["a.c.1"] != false {
		t.Errorf("数组平铺结果缺失: %v", tiled)
	}
	back, err := UntilingMap(tiled)
	if err != nil {
		t.Fatalf("还原不应报错: %v", err)
	}
	// 数组还原为下标 map, 数值部分与原文档嵌套值一致
	if !reflect.DeepEqual(back["a"].(map[string]any)["b"], doc["a"].(map[string]any)["b"]) {
		t.Errorf("往返后 a.b 不一致: %v", back)
	}
}
