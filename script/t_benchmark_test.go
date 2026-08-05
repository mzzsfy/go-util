package script

import (
	"fmt"
	"strings"
	"testing"
)

// ========== Lexer 基准测试 ==========

func BenchmarkLexerSimple(b *testing.B) {
	code := "x := 10 + 20"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(code)
		lexer.Tokenize()
	}
}

func BenchmarkLexerComplex(b *testing.B) {
	code := `
		x := 10 + 20 * 30
		y := x / 5
		if x > 100 {
			y = y + 1
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(code)
		lexer.Tokenize()
	}
}

func BenchmarkLexerString(b *testing.B) {
	code := fmt.Sprintf(`s := "%s"`, strings.Repeat("hello world ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(code)
		lexer.Tokenize()
	}
}

// ========== Parser 基准测试 ==========

func BenchmarkParserSimple(b *testing.B) {
	code := "x := 10 + 20"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		parser.Compile(code)
	}
}

func BenchmarkParserComplex(b *testing.B) {
	code := `
		x := 10 + 20 * 30
		y := x / 5
		if x > 100 {
			y = y + 1
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		parser.Compile(code)
	}
}

func BenchmarkParserFunction(b *testing.B) {
	code := `
		fn add(a, b) {
			return a + b
		}
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		parser.Compile(code)
	}
}

// ========== VM 基准测试 ==========

func BenchmarkVMArithmetic(b *testing.B) {
	code := "x := 10 + 20 * 30 / 5"
	script, _ := NewParser().Compile(code)
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}

func BenchmarkVMComparison(b *testing.B) {
	code := "x := 10 < 20 && 30 > 15"
	script, _ := NewParser().Compile(code)
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}

func BenchmarkVMStringConcat(b *testing.B) {
	code := `x := "hello" + " " + "world"`
	script, _ := NewParser().Compile(code)
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}

func BenchmarkVMArrayAccess(b *testing.B) {
	code := `
		arr := [1, 2, 3, 4, 5]
		x := arr[2]
	`
	script, _ := NewParser().Compile(code)
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}

func BenchmarkVMBranch(b *testing.B) {
	code := `
		x := 10
		if x > 5 {
			y := 1
		} else {
			y := 2
		}
	`
	script, _ := NewParser().Compile(code)
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}

func BenchmarkVMExternalCall(b *testing.B) {
	code := `
		#fn add(int, int)=>int
		x := add(10, 20)
	`
	script, _ := NewParser().Compile(code)
	ctx := NewContext()
	ctx.BindFunc("add", func(a, b int) int {
		return a + b
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}

// ========== Value 创建基准测试 ==========

func BenchmarkNewValueInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewValue(123)
	}
}

func BenchmarkNewValueString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewValue("hello world")
	}
}

func BenchmarkNewValueArray(b *testing.B) {
	elements := []Value{
		NewValue(1),
		NewValue(2),
		NewValue(3),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewValue(elements)
	}
}

func BenchmarkNewValueMap(b *testing.B) {
	m := &MapValue{
		Pairs:   make(map[string]Value),
		KeyType: TypeString,
	}
	m.Pairs["key"] = NewValue(123)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewValue(m)
	}
}

// ========== 运行时操作基准测试 ==========

func BenchmarkRuntimeAdd(b *testing.B) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)
	a := NewValue(10)
	c := NewValue(20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.add(a, c)
	}
}

func BenchmarkRuntimeStringConcat(b *testing.B) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)
	a := NewValue("hello")
	c := NewValue("world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.add(a, c)
	}
}

func BenchmarkRuntimeArrayAccess(b *testing.B) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)
	arr := NewValue([]Value{
		NewValue(1),
		NewValue(2),
		NewValue(3),
	})
	idx := NewValue(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.indexAccess(arr, idx)
	}
}

func BenchmarkRuntimeTypeConversion(b *testing.B) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)
	val := NewValue("123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.toInt(val)
	}
}

// ========== 完整流程基准测试 ==========

func BenchmarkFullPipeline(b *testing.B) {
	code := `
		x := 10
		y := 20
		z := x + y
		if z > 25 {
			result := z * 2
		}
	`
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		script, _ := parser.Compile(code)
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}

func BenchmarkFullPipelineWithExternal(b *testing.B) {
	code := `
		#fn multiply(int, int)=>int
		x := multiply(10, 20)
		y := x + 5
	`
	ctx := NewContext()
	ctx.BindFunc("multiply", func(a, b int) int {
		return a * b
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		script, _ := parser.Compile(code)
		vm := getVMFromPool(ctx, DefaultMaxCallDepth)
		vm.Run(script)
		returnVMToPool(vm)
	}
}
