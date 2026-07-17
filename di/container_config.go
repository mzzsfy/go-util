// Package di 提供配置注入功能
package di

import (
	"strings"
	"sync/atomic"

	"github.com/mzzsfy/go-util/config"
)

// parseConfigVariable 解析配置变量
// 格式为 "key:defaultValue"，返回键和默认值
func parseConfigVariable(varPart string) (key, defaultValue string) {
	colonIdx := strings.Index(varPart, ":")
	if colonIdx == -1 {
		return varPart, ""
	}
	return varPart[:colonIdx], varPart[colonIdx+1:]
}

// extractVariablePart 从字符串中提取变量部分
// 查找 ${...} 格式的变量
// 返回变量内容、结束位置和是否找到
func extractVariablePart(remaining string) (varPart string, end int, found bool) {
	start := strings.Index(remaining, "${")
	if start == -1 {
		return "", -1, false
	}

	endIdx := strings.Index(remaining[start:], "}")
	if endIdx == -1 {
		return "", -1, false
	}

	end = start + endIdx
	varPart = remaining[start+2 : end]
	return varPart, end, true
}

// resolveConfigValueSimple 简单配置值解析
// 直接从配置源获取值
func (c *container) resolveConfigValueSimple(tag string) string {
	value := c.getConfigValue(tag)
	return value.StringD("")
}

// resolveConfigValueWithVariables 带变量的配置值解析
// 支持 ${key:default} 格式的变量替换
// 使用 strings.Builder 优化字符串拼接性能
func (c *container) resolveConfigValueWithVariables(tag string) string {
	// 预计算容量：至少等于原字符串长度
	var sb strings.Builder
	sb.Grow(len(tag))

	remaining := tag
	for {
		varPart, end, found := extractVariablePart(remaining)
		if !found {
			sb.WriteString(remaining)
			break
		}

		varStart := strings.Index(remaining, "${")
		if varStart > 0 {
			sb.WriteString(remaining[:varStart])
		}

		key, defaultValue := parseConfigVariable(varPart)
		value := c.getConfigValue(key)
		sb.WriteString(value.StringD(defaultValue))

		remaining = remaining[end+1:]
	}

	return sb.String()
}

// resolveConfigValue 配置值解析入口
// 根据是否包含变量选择解析方式
func (c *container) resolveConfigValue(tag string) string {
	if !strings.Contains(tag, "${") {
		return c.resolveConfigValueSimple(tag)
	}
	return c.resolveConfigValueWithVariables(tag)
}

// parseConfigInjection 解析配置注入标签
// 支持简单格式和变量格式
func parseConfigInjection(tag string) (key string, defaultValue string) {
	if !strings.Contains(tag, "${") {
		return parseConfigVariable(tag)
	}

	varPart, _, found := extractVariablePart(tag)
	if !found {
		return tag, ""
	}

	return parseConfigVariable(varPart)
}

// getConfigValue 获取配置值
// 线程安全，带缓存统计
func (c *container) getConfigValue(key string) config.Value {
	c.configMu.RLock()
	defer c.configMu.RUnlock()

	if c.configSource == nil || key == "" {
		c.updateConfigStats(false)
		return config.ValueFrom(nil)
	}

	return c.getConfigFromSource(key)
}

// updateConfigStats 更新配置访问统计
func (c *container) updateConfigStats(hit bool) {
	if hit {
		atomic.AddInt64(&c.stats.configHits, 1)
	} else {
		atomic.AddInt64(&c.stats.configMisses, 1)
	}
}

// getConfigFromSource 从配置源获取值
// 同时更新统计信息
func (c *container) getConfigFromSource(key string) config.Value {
	value := c.configSource.Get(key)
	if value != nil {
		c.updateConfigStats(true)
		return value
	}
	c.updateConfigStats(false)
	return config.ValueFrom(nil)
}
