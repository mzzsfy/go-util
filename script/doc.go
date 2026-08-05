// Package script 提供嵌入式脚本语言引擎
//
// 支持动态脚本执行、变量绑定、外部函数调用和完整控制流语句。
// 源代码经词法分析、语法分析、编译器生成字节码,由虚拟机执行。
//
// 基本用法:
//
//	result, err := script.Eval("10 + 20")
//	fmt.Println(result.Int()) // 30
package script
