package script

import (
	"fmt"
	"strings"
)

// ========== 错误类型定义 ==========
// 本文件定义脚本引擎的错误类型
// 包括编译时错误和运行时错误

// ========== 错误消息格式常量 ==========
// 统一错误消息格式，提高一致性和可维护性

const (
	// ErrorPrefixProblem 问题前缀
	ErrorPrefixProblem = "→ 问题："
	// ErrorPrefixReason 原因前缀
	ErrorPrefixReason = "→ 原因："
	// ErrorPrefixSuggestion 建议前缀
	ErrorPrefixSuggestion = "→ 建议："
	// ErrorPrefixExample 示例前缀
	ErrorPrefixExample = "→ 示例："
	// ErrorPrefixTip 提示前缀
	ErrorPrefixTip = "→ 提示："
)

// CompileError 编译错误
// 包含错误位置（行号、列号）和错误消息
type CompileError struct {
	// Line 错误所在行号
	Line int
	// Column 错误所在列号
	Column int
	// Message 错误消息
	Message string
}

// Error 实现error接口
// 返回包含位置信息的详细错误消息
func (e *CompileError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("编译错误(行:%d, 列:%d): %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("编译错误: %s", e.Message)
}

// NewCompileError 创建编译错误
// 参数: line-行号, column-列号, format-格式化字符串, args-参数
func NewCompileError(line, column int, format string, args ...interface{}) *CompileError {
	return &CompileError{
		Line:    line,
		Column:  column,
		Message: fmt.Sprintf(format, args...),
	}
}

// Pos 位置接口
// 用于从AST节点获取位置信息
type Pos interface {
	Pos() Position
}

// NewCompileErrorFromPos 从位置接口创建编译错误
// 自动从节点提取行号和列号
func NewCompileErrorFromPos(pos Pos, format string, args ...interface{}) *CompileError {
	p := pos.Pos()
	return &CompileError{
		Line:    p.Line,
		Column:  p.Column,
		Message: fmt.Sprintf(format, args...),
	}
}

// ErrCode 运行时错误码
type ErrCode string

const (
	// ErrTypeMismatch 类型不匹配
	ErrTypeMismatch ErrCode = "type_mismatch"
	// ErrIndexOutOfBounds 索引越界
	ErrIndexOutOfBounds ErrCode = "index_out_of_bounds"
	// ErrUndefinedVar 未定义变量
	ErrUndefinedVar ErrCode = "undefined_var"
	// ErrCallStackOverflow 调用栈溢出
	ErrCallStackOverflow ErrCode = "call_stack_overflow"
	// ErrTimeout 执行超时
	ErrTimeout ErrCode = "timeout"
	// ErrStepLimit 指令数超限
	ErrStepLimit ErrCode = "step_limit"
	// ErrPanic panic恢复
	ErrPanic ErrCode = "panic"
	// ErrUnsupportedOp 不支持的操作
	ErrUnsupportedOp ErrCode = "unsupported_op"
	// ErrDivisionByZero 除零
	ErrDivisionByZero ErrCode = "division_by_zero"
)

// RuntimeError 运行时错误
// 包含错误消息和调用栈信息
type RuntimeError struct {
	// Code 错误码, 用于 errors.Is 分类匹配
	Code ErrCode
	// Message 错误消息
	Message string
	// StackTrace 调用栈
	StackTrace []string
}

// Is 支持 errors.Is, 按 Code 匹配
func (e *RuntimeError) Is(target error) bool {
	if t, ok := target.(*RuntimeError); ok {
		return e.Code == t.Code && t.Code != ""
	}
	return false
}

// Error 实现error接口
// 返回包含调用栈的详细错误消息
func (e *RuntimeError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("运行时错误: %s", e.Message))
	if len(e.StackTrace) > 0 {
		sb.WriteString("\n调用栈:\n")
		for _, trace := range e.StackTrace {
			sb.WriteString(fmt.Sprintf("  %s\n", trace))
		}
	}
	return sb.String()
}

// ========== 错误消息构建器 ==========

// ErrorMessageBuilder 错误消息构建器
// 用于构建统一格式的错误消息，减少重复代码
type ErrorMessageBuilder struct {
	problem    string
	reason     string
	suggestion string
	example    string
	tip        string
}

// NewErrorMessage 创建错误消息构建器
func NewErrorMessage(problem string) *ErrorMessageBuilder {
	return &ErrorMessageBuilder{problem: problem}
}

// WithReason 添加原因说明
func (b *ErrorMessageBuilder) WithReason(reason string) *ErrorMessageBuilder {
	b.reason = reason
	return b
}

// WithSuggestion 添加建议
func (b *ErrorMessageBuilder) WithSuggestion(suggestion string) *ErrorMessageBuilder {
	b.suggestion = suggestion
	return b
}

// WithExample 添加示例
func (b *ErrorMessageBuilder) WithExample(example string) *ErrorMessageBuilder {
	b.example = example
	return b
}

// WithTip 添加提示
func (b *ErrorMessageBuilder) WithTip(tip string) *ErrorMessageBuilder {
	b.tip = tip
	return b
}

// Build 构建错误消息
func (b *ErrorMessageBuilder) Build() string {
	var sb strings.Builder
	sb.WriteString(b.problem)

	if b.reason != "" {
		sb.WriteString("\n")
		sb.WriteString(ErrorPrefixReason)
		sb.WriteString(b.reason)
	}

	if b.suggestion != "" {
		sb.WriteString("\n")
		sb.WriteString(ErrorPrefixSuggestion)
		sb.WriteString(b.suggestion)
	}

	if b.example != "" {
		sb.WriteString("\n")
		sb.WriteString(ErrorPrefixExample)
		sb.WriteString(b.example)
	}

	if b.tip != "" {
		sb.WriteString("\n")
		sb.WriteString(ErrorPrefixTip)
		sb.WriteString(b.tip)
	}

	return sb.String()
}
