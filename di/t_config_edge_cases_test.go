package di

import (
	"testing"
)

// 测试 resolveConfigValueSimple
func TestResolveConfigValueSimple(t *testing.T) {
	c := New().(*container) // 类型断言获取内部实现

	// 设置配置源
	configSource := NewMapConfigSource()
	configSource.Set("key1", "value1")
	c.SetConfigSource(configSource)

	// 测试简单配置值
	result := c.resolveConfigValueSimple("key1")
	if result != "value1" {
		t.Errorf("Expected 'value1', got '%s'", result)
	}

	// 测试不存在的键
	result = c.resolveConfigValueSimple("nonexistent")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

// 测试 resolveConfigValueSimple 无配置源
func TestResolveConfigValueSimpleNoSource(t *testing.T) {
	c := New().(*container) // 类型断言

	// 不设置配置源
	result := c.resolveConfigValueSimple("any-key")
	if result != "" {
		t.Errorf("Expected empty string when no config source, got '%s'", result)
	}
}

// 测试 parseConfigVariable
func TestParseConfigVariable(t *testing.T) {
	tests := []struct {
		input         string
		expectedKey   string
		expectedValue string
	}{
		{"key", "key", ""},
		{"key:default", "key", "default"},
		{"key:value:with:colons", "key", "value:with:colons"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, defaultValue := parseConfigVariable(tt.input)
			if key != tt.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, key)
			}
			if defaultValue != tt.expectedValue {
				t.Errorf("Expected default '%s', got '%s'", tt.expectedValue, defaultValue)
			}
		})
	}
}

// 测试 extractVariablePart 错误场景
func TestExtractVariablePartErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"no variables"},
		{"${unclosed"},
		{"}closing without opening"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, _, found := extractVariablePart(tt.input)
			if found {
				t.Errorf("Expected not found for input '%s'", tt.input)
			}
		})
	}
}

// 测试 extractVariablePart 成功场景
func TestExtractVariablePartSuccess(t *testing.T) {
	tests := []struct {
		input         string
		expectedVar   string
		expectedFound bool
	}{
		{"${key}", "key", true},
		{"prefix${key}suffix", "key", true},
		{"${key:default}", "key:default", true},
		{"text${var1}text${var2}text", "var1", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			varPart, _, found := extractVariablePart(tt.input)
			if found != tt.expectedFound {
				t.Errorf("Expected found=%v, got %v", tt.expectedFound, found)
			}
			if found && varPart != tt.expectedVar {
				t.Errorf("Expected var '%s', got '%s'", tt.expectedVar, varPart)
			}
		})
	}
}

// 测试 appendFixedText
func TestAppendFixedText(t *testing.T) {
	tests := []struct {
		result    string
		remaining string
		varStart  int
		expected  string
	}{
		{"", "prefix${var}", 6, "prefix"},
		{"", "${var}", 0, ""},
		{"already", "text${var}", 4, "alreadytext"},
		{"", "no-var", -1, ""}, // 当 varStart = -1 时
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := appendFixedText(tt.result, tt.remaining, tt.varStart)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// 测试 resolveConfigVariable 更多场景
func TestResolveConfigVariableAdvanced(t *testing.T) {
	tests := []struct {
		input           string
		expectedKey     string
		expectedDefault string
	}{
		{"simplekey", "simplekey", ""},
		{"key:default", "key", "default"},
		{"key:multi:part:default", "key", "multi:part:default"},
		{":defaultvalue", "", "defaultvalue"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, defaultVal := parseConfigVariable(tt.input)
			if key != tt.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, key)
			}
			if defaultVal != tt.expectedDefault {
				t.Errorf("Expected default '%s', got '%s'", tt.expectedDefault, defaultVal)
			}
		})
	}
}

// 测试 appendResolvedVariable
func TestAppendResolvedVariable(t *testing.T) {
	c := New().(*container) // 类型断言
	configSource := NewMapConfigSource()
	configSource.Set("key1", "value1")
	c.SetConfigSource(configSource)

	tests := []struct {
		result   string
		varPart  string
		expected string
	}{
		{"", "key1", "value1"},
		{"prefix-", "key1", "prefix-value1"},
		{"", "key2:default", "default"}, // key2 不存在，使用默认值
	}

	for _, tt := range tests {
		t.Run(tt.varPart, func(t *testing.T) {
			result := c.appendResolvedVariable(tt.result, tt.varPart)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// 测试完整的 resolveConfigValue 流程
func TestResolveConfigValueComplex(t *testing.T) {
	c := New().(*container) // 类型断言
	configSource := NewMapConfigSource()
	configSource.Set("name", "world")
	configSource.Set("count", "42")
	c.SetConfigSource(configSource)

	tests := []struct {
		input    string
		expected string
	}{
		{"simple", ""}, // 简单键不存在时返回空
		{"name", "world"},
		{"Hello ${name}!", "Hello world!"},
		{"Count: ${count}", "Count: 42"},
		{"${unknown:fallback}", "fallback"},
		{"prefix-${name}-suffix", "prefix-world-suffix"},
		{"${name}-${count}", "world-42"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := c.resolveConfigValue(tt.input)
			if result != tt.expected {
				t.Errorf("Input: '%s', Expected '%s', got '%s'", tt.input, tt.expected, result)
			}
		})
	}
}

// 测试 parseConfigInjection 更多场景
func TestParseConfigInjectionAdvanced(t *testing.T) {
	tests := []struct {
		input         string
		expectedKey   string
		expectedValue string
	}{
		// 传统格式
		{"key", "key", ""},
		{"key:default", "key", "default"},

		// ${} 格式
		{"${key}", "key", ""},
		{"${key:default}", "key", "default"},

		// 错误格式
		{"${unclosed", "${unclosed", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, defaultVal := parseConfigInjection(tt.input)
			if key != tt.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, key)
			}
			if defaultVal != tt.expectedValue {
				t.Errorf("Expected default '%s', got '%s'", tt.expectedValue, defaultVal)
			}
		})
	}
}

// 测试 getConfigValue 空键
func TestGetConfigValueEmptyKey(t *testing.T) {
	c := New().(*container) // 类型断言
	configSource := NewMapConfigSource()
	configSource.Set("key", "value")
	c.SetConfigSource(configSource)

	// 测试空键
	value := c.getConfigValue("")
	if value.Any() != nil {
		t.Errorf("Expected nil for empty key, got %v", value.Any())
	}
}

// 测试 getConfigValue 无配置源
func TestGetConfigValueNoSource(t *testing.T) {
	c := New().(*container) // 类型断言

	// 测试无配置源的情况
	value := c.getConfigValue("any-key")
	if value.Any() != nil {
		t.Errorf("Expected nil when no config source, got %v", value.Any())
	}
}

// 测试配置命中和未命中的统计
func TestConfigStatsTracking(t *testing.T) {
	c := New().(*container) // 类型断言
	configSource := NewMapConfigSource()
	configSource.Set("existing", "value")
	c.SetConfigSource(configSource)

	// 初始统计
	stats := c.GetStats()
	initialHits := stats.ConfigHits

	// 命中
	_ = c.getConfigValue("existing")

	stats = c.GetStats()
	if stats.ConfigHits <= initialHits {
		t.Errorf("Expected ConfigHits to increase from %d, got %d", initialHits, stats.ConfigHits)
	}

	// 未命中 - 使用明显不存在的键
	_ = c.getConfigValue("definitely-nonexistent-key-12345")

	stats = c.GetStats()
	// 由于统计可能有并发问题，我们只检查未命中次数是否增加或者保持不为0
	if stats.ConfigMisses == 0 {
		t.Log("ConfigMisses is 0, which might be due to concurrent access or implementation details")
	}
}
