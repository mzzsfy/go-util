package script

import (
	"testing"
)

// ========== AST节点与解析器结构深度测试 ==========
// 通过直接检查parseProgram返回的AST结构和间接行为验证,
// 深度测试各类AST节点的字段正确性和解析器行为

// parseAST 解析源码并返回Program AST, 断言无错误
func parseAST(t *testing.T, source string) *Program {
	t.Helper()
	parser := NewParser()
	parser.initLexer(source)
	program := parser.parseProgram()
	if len(parser.errors) > 0 {
		t.Fatalf("解析失败 [%s]: %v", source, parser.errors[0])
	}
	return program
}

// parseASTWithErrs 解析源码并返回Program AST和错误列表
func parseASTWithErrs(t *testing.T, source string) (*Program, []*CompileError) {
	t.Helper()
	parser := NewParser()
	parser.initLexer(source)
	program := parser.parseProgram()
	return program, parser.errors
}

// stmtAs 将program的第idx条语句断言为目标类型
func stmtAs[T Stmt](t *testing.T, program *Program, idx int) T {
	t.Helper()
	if idx >= len(program.Statements) {
		t.Fatalf("语句索引越界: %d >= %d", idx, len(program.Statements))
	}
	result, ok := program.Statements[idx].(T)
	if !ok {
		t.Fatalf("语句[%d]类型断言失败: 期望 %T, 实际 %T", idx, *new(T), program.Statements[idx])
	}
	return result
}

// ========== 字面量表达式AST验证 ==========

func Test_ASTNode_IntExpr(t *testing.T) {
	t.Run("整数字面量AST结构", func(t *testing.T) {
		program := parseAST(t, "42")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit, ok := stmt.Expr.(*LiteralExpr)
		if !ok {
			t.Fatalf("期望 *LiteralExpr, 得到 %T", stmt.Expr)
		}
		if lit.Type != LiteralInt {
			t.Errorf("期望 LiteralInt, 得到 %v", lit.Type)
		}
		if lit.Value.(int) != 42 {
			t.Errorf("期望 42, 得到 %v", lit.Value)
		}
	})
	t.Run("十六进制整数字面量", func(t *testing.T) {
		program := parseAST(t, "0xFF")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit := stmt.Expr.(*LiteralExpr)
		if lit.Value.(int) != 255 {
			t.Errorf("期望 255, 得到 %v", lit.Value)
		}
	})
	t.Run("整数值正确传递到执行", func(t *testing.T) {
		runIntTest(t, "12345", 12345)
	})
}

func Test_ASTNode_FloatExpr(t *testing.T) {
	t.Run("浮点数字面量AST结构", func(t *testing.T) {
		program := parseAST(t, "3.14")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit, ok := stmt.Expr.(*LiteralExpr)
		if !ok {
			t.Fatalf("期望 *LiteralExpr, 得到 %T", stmt.Expr)
		}
		if lit.Type != LiteralFloat {
			t.Errorf("期望 LiteralFloat, 得到 %v", lit.Type)
		}
		if lit.Value.(float64) != 3.14 {
			t.Errorf("期望 3.14, 得到 %v", lit.Value)
		}
	})
	t.Run("浮点数值正确传递到执行", func(t *testing.T) {
		runFloatTest(t, "2.5", 2.5)
	})
}

func Test_ASTNode_StringExpr(t *testing.T) {
	t.Run("字符串字面量AST结构", func(t *testing.T) {
		program := parseAST(t, `"hello"`)
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit, ok := stmt.Expr.(*LiteralExpr)
		if !ok {
			t.Fatalf("期望 *LiteralExpr, 得到 %T", stmt.Expr)
		}
		if lit.Type != LiteralString {
			t.Errorf("期望 LiteralString, 得到 %v", lit.Type)
		}
		if lit.Value.(string) != "hello" {
			t.Errorf("期望 hello, 得到 %v", lit.Value)
		}
	})
	t.Run("空字符串字面量", func(t *testing.T) {
		program := parseAST(t, `""`)
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit := stmt.Expr.(*LiteralExpr)
		if lit.Value.(string) != "" {
			t.Errorf("期望空字符串, 得到 %v", lit.Value)
		}
	})
}

func Test_ASTNode_BoolExpr(t *testing.T) {
	t.Run("布尔字面量AST结构", func(t *testing.T) {
		program := parseAST(t, "true")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit, ok := stmt.Expr.(*LiteralExpr)
		if !ok {
			t.Fatalf("期望 *LiteralExpr, 得到 %T", stmt.Expr)
		}
		if lit.Type != LiteralBool {
			t.Errorf("期望 LiteralBool, 得到 %v", lit.Type)
		}
		if lit.Value.(bool) != true {
			t.Errorf("期望 true, 得到 %v", lit.Value)
		}
	})
	t.Run("false字面量", func(t *testing.T) {
		program := parseAST(t, "false")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit := stmt.Expr.(*LiteralExpr)
		if lit.Value.(bool) != false {
			t.Errorf("期望 false, 得到 %v", lit.Value)
		}
	})
}

func Test_ASTNode_NilExpr(t *testing.T) {
	t.Run("nil字面量AST结构", func(t *testing.T) {
		program := parseAST(t, "nil")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		lit, ok := stmt.Expr.(*LiteralExpr)
		if !ok {
			t.Fatalf("期望 *LiteralExpr, 得到 %T", stmt.Expr)
		}
		if lit.Type != LiteralNil {
			t.Errorf("期望 LiteralNil, 得到 %v", lit.Type)
		}
		if lit.Value != nil {
			t.Errorf("期望 nil, 得到 %v", lit.Value)
		}
	})
}

// ========== 标识符与复合类型AST验证 ==========

func Test_ASTNode_IdentExpr(t *testing.T) {
	t.Run("标识符AST结构", func(t *testing.T) {
		program := parseAST(t, "x := 1\nx")
		// 第二条语句是标识符表达式
		stmt := stmtAs[*ExprStmt](t, program, 1)
		ident, ok := stmt.Expr.(*IdentExpr)
		if !ok {
			t.Fatalf("期望 *IdentExpr, 得到 %T", stmt.Expr)
		}
		if ident.Name != "x" {
			t.Errorf("期望 Name=x, 得到 %s", ident.Name)
		}
	})
	t.Run("标识符变量引用执行", func(t *testing.T) {
		runIntTest(t, "x := 99\nx", 99)
	})
}

func Test_ASTNode_ArrayExpr(t *testing.T) {
	t.Run("数组元素顺序正确", func(t *testing.T) {
		program := parseAST(t, "[10, 20, 30]")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		arr, ok := stmt.Expr.(*ArrayExpr)
		if !ok {
			t.Fatalf("期望 *ArrayExpr, 得到 %T", stmt.Expr)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("期望 3 个元素, 得到 %d", len(arr.Elements))
		}
		vals := []int{}
		for _, el := range arr.Elements {
			lit := el.(*LiteralExpr)
			vals = append(vals, lit.Value.(int))
		}
		if vals[0] != 10 || vals[1] != 20 || vals[2] != 30 {
			t.Errorf("期望 [10,20,30], 得到 %v", vals)
		}
	})
	t.Run("空数组AST结构", func(t *testing.T) {
		program := parseAST(t, "[]")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		arr := stmt.Expr.(*ArrayExpr)
		if len(arr.Elements) != 0 {
			t.Errorf("期望 0 个元素, 得到 %d", len(arr.Elements))
		}
	})
}

func Test_ASTNode_MapExpr(t *testing.T) {
	t.Run("Map键值对正确", func(t *testing.T) {
		program := parseAST(t, `{"a": 1, "b": 2}`)
		stmt := stmtAs[*ExprStmt](t, program, 0)
		m, ok := stmt.Expr.(*MapExpr)
		if !ok {
			t.Fatalf("期望 *MapExpr, 得到 %T", stmt.Expr)
		}
		if len(m.Pairs) != 2 {
			t.Fatalf("期望 2 对, 得到 %d", len(m.Pairs))
		}
		k0 := m.Pairs[0].Key.(*LiteralExpr).Value.(string)
		v0 := m.Pairs[0].Value.(*LiteralExpr).Value.(int)
		if k0 != "a" || v0 != 1 {
			t.Errorf("期望 a:1, 得到 %s:%v", k0, v0)
		}
		k1 := m.Pairs[1].Key.(*LiteralExpr).Value.(string)
		v1 := m.Pairs[1].Value.(*LiteralExpr).Value.(int)
		if k1 != "b" || v1 != 2 {
			t.Errorf("期望 b:2, 得到 %s:%v", k1, v1)
		}
	})
	t.Run("空Map AST结构", func(t *testing.T) {
		program := parseAST(t, "{}")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		m := stmt.Expr.(*MapExpr)
		if len(m.Pairs) != 0 {
			t.Errorf("期望 0 对, 得到 %d", len(m.Pairs))
		}
	})
}

// ========== 运算符AST验证 ==========

func Test_ASTNode_BinaryExpr(t *testing.T) {
	t.Run("加法运算符AST结构", func(t *testing.T) {
		program := parseAST(t, "1 + 2")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		bin, ok := stmt.Expr.(*BinaryExpr)
		if !ok {
			t.Fatalf("期望 *BinaryExpr, 得到 %T", stmt.Expr)
		}
		if bin.Operator != "+" {
			t.Errorf("期望运算符 +, 得到 %s", bin.Operator)
		}
		left := bin.Left.(*LiteralExpr).Value.(int)
		right := bin.Right.(*LiteralExpr).Value.(int)
		if left != 1 || right != 2 {
			t.Errorf("期望 1+2, 得到 %d+%d", left, right)
		}
	})
	t.Run("比较运算符AST", func(t *testing.T) {
		program := parseAST(t, "1 < 2")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		bin := stmt.Expr.(*BinaryExpr)
		if bin.Operator != "<" {
			t.Errorf("期望运算符 <, 得到 %s", bin.Operator)
		}
	})
	t.Run("逻辑运算符AST", func(t *testing.T) {
		program := parseAST(t, "true && false")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		bin := stmt.Expr.(*BinaryExpr)
		if bin.Operator != "&&" {
			t.Errorf("期望运算符 &&, 得到 %s", bin.Operator)
		}
	})
}

func Test_ASTNode_UnaryExpr(t *testing.T) {
	t.Run("负号一元运算符AST", func(t *testing.T) {
		program := parseAST(t, "-5")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		un, ok := stmt.Expr.(*UnaryExpr)
		if !ok {
			t.Fatalf("期望 *UnaryExpr, 得到 %T", stmt.Expr)
		}
		if un.Operator != "-" {
			t.Errorf("期望运算符 -, 得到 %s", un.Operator)
		}
		operand := un.Operand.(*LiteralExpr).Value.(int)
		if operand != 5 {
			t.Errorf("期望操作数 5, 得到 %d", operand)
		}
	})
	t.Run("逻辑非一元运算符AST", func(t *testing.T) {
		program := parseAST(t, "!true")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		un := stmt.Expr.(*UnaryExpr)
		if un.Operator != "!" {
			t.Errorf("期望运算符 !, 得到 %s", un.Operator)
		}
	})
}

func Test_ASTNode_NestedBinary(t *testing.T) {
	t.Run("嵌套运算AST结构", func(t *testing.T) {
		// 1 + 2 * 3 -> BinaryExpr{1, +, BinaryExpr{2, *, 3}}
		program := parseAST(t, "1 + 2 * 3")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		outer := stmt.Expr.(*BinaryExpr)
		if outer.Operator != "+" {
			t.Fatalf("期望外层运算符 +, 得到 %s", outer.Operator)
		}
		inner, ok := outer.Right.(*BinaryExpr)
		if !ok {
			t.Fatalf("期望内层 *BinaryExpr, 得到 %T", outer.Right)
		}
		if inner.Operator != "*" {
			t.Errorf("期望内层运算符 *, 得到 %s", inner.Operator)
		}
	})
	t.Run("括号改变嵌套结构", func(t *testing.T) {
		// (1 + 2) * 3 -> BinaryExpr{BinaryExpr{1,+,2}, *, 3}
		program := parseAST(t, "(1 + 2) * 3")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		outer := stmt.Expr.(*BinaryExpr)
		if outer.Operator != "*" {
			t.Fatalf("期望外层运算符 *, 得到 %s", outer.Operator)
		}
		inner, ok := outer.Left.(*BinaryExpr)
		if !ok {
			t.Fatalf("期望左子树 *BinaryExpr, 得到 %T", outer.Left)
		}
		if inner.Operator != "+" {
			t.Errorf("期望内层运算符 +, 得到 %s", inner.Operator)
		}
	})
}

func Test_ASTNode_Precedence(t *testing.T) {
	t.Run("乘法优先级高于加法", func(t *testing.T) {
		// 乘法应嵌套在加法的右子树中
		program := parseAST(t, "1 + 2 * 3")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		top := stmt.Expr.(*BinaryExpr)
		if top.Operator != "+" {
			t.Fatalf("期望顶层 +, 得到 %s", top.Operator)
		}
		if _, ok := top.Right.(*BinaryExpr); !ok {
			t.Error("期望右子树为乘法子表达式")
		}
	})
	t.Run("AND优先级高于OR", func(t *testing.T) {
		// true || false && false -> OR(true, AND(false, false))
		program := parseAST(t, "true || false && false")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		top := stmt.Expr.(*BinaryExpr)
		if top.Operator != "||" {
			t.Fatalf("期望顶层 ||, 得到 %s", top.Operator)
		}
		if _, ok := top.Right.(*BinaryExpr); !ok {
			t.Error("期望右子树为AND子表达式")
		}
	})
}

func Test_ASTNode_LeftAssoc(t *testing.T) {
	t.Run("减法左结合AST结构", func(t *testing.T) {
		// 10 - 3 - 2 -> BinaryExpr{BinaryExpr{10,-,3}, -, 2}
		program := parseAST(t, "10 - 3 - 2")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		top := stmt.Expr.(*BinaryExpr)
		if top.Operator != "-" {
			t.Fatalf("期望顶层 -, 得到 %s", top.Operator)
		}
		left, ok := top.Left.(*BinaryExpr)
		if !ok {
			t.Fatalf("期望左子树为减法子表达式, 得到 %T", top.Left)
		}
		if left.Operator != "-" {
			t.Errorf("期望左子树运算符 -, 得到 %s", left.Operator)
		}
	})
	t.Run("除法左结合AST结构", func(t *testing.T) {
		// 100 / 5 / 2 -> BinaryExpr{BinaryExpr{100,/,5}, /, 2}
		program := parseAST(t, "100 / 5 / 2")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		top := stmt.Expr.(*BinaryExpr)
		left, ok := top.Left.(*BinaryExpr)
		if !ok {
			t.Fatalf("期望左子树为除法子表达式")
		}
		if left.Operator != "/" {
			t.Errorf("期望左子树运算符 /, 得到 %s", left.Operator)
		}
	})
}

// ========== 语句AST验证 ==========

func Test_ASTNode_VarDeclStmt(t *testing.T) {
	t.Run("变量声明AST结构", func(t *testing.T) {
		program := parseAST(t, "x := 42")
		stmt := stmtAs[*VarDeclStmt](t, program, 0)
		if stmt.Name != "x" {
			t.Errorf("期望 Name=x, 得到 %s", stmt.Name)
		}
		if stmt.TypeAnnot != nil {
			t.Errorf("期望无类型注解, 得到 %T", stmt.TypeAnnot)
		}
		lit, ok := stmt.Value.(*LiteralExpr)
		if !ok {
			t.Fatalf("期望 Value 为 *LiteralExpr, 得到 %T", stmt.Value)
		}
		if lit.Value.(int) != 42 {
			t.Errorf("期望 42, 得到 %v", lit.Value)
		}
	})
	t.Run("变量声明执行验证", func(t *testing.T) {
		runIntTest(t, "y := 7\ny", 7)
	})
}

func Test_ASTNode_ExprStmt(t *testing.T) {
	t.Run("表达式语句AST结构", func(t *testing.T) {
		program := parseAST(t, "1 + 2")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		if _, ok := stmt.Expr.(*BinaryExpr); !ok {
			t.Fatalf("期望 *BinaryExpr, 得到 %T", stmt.Expr)
		}
	})
	t.Run("字面量表达式语句", func(t *testing.T) {
		program := parseAST(t, "42")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		if _, ok := stmt.Expr.(*LiteralExpr); !ok {
			t.Fatalf("期望 *LiteralExpr, 得到 %T", stmt.Expr)
		}
	})
}

func Test_ASTNode_IfStmt(t *testing.T) {
	t.Run("if语句AST结构", func(t *testing.T) {
		program := parseAST(t, "if true { 1 }")
		stmt := stmtAs[*IfStmt](t, program, 0)
		if stmt.Condition == nil {
			t.Fatal("期望 Condition 非空")
		}
		if stmt.Then == nil {
			t.Fatal("期望 Then 块非空")
		}
		if len(stmt.Then.Statements) != 1 {
			t.Errorf("期望 Then 有 1 条语句, 得到 %d", len(stmt.Then.Statements))
		}
		if stmt.Else != nil {
			t.Errorf("期望 Else 为空")
		}
	})
	t.Run("if-else语句AST结构", func(t *testing.T) {
		program := parseAST(t, "if true { 1 } else { 2 }")
		stmt := stmtAs[*IfStmt](t, program, 0)
		if stmt.Else == nil {
			t.Fatal("期望 Else 非空")
		}
		elseBlock, ok := stmt.Else.(*BlockStmt)
		if !ok {
			t.Fatalf("期望 Else 为 *BlockStmt, 得到 %T", stmt.Else)
		}
		if len(elseBlock.Statements) != 1 {
			t.Errorf("期望 Else 有 1 条语句, 得到 %d", len(elseBlock.Statements))
		}
	})
	t.Run("else-if链AST结构", func(t *testing.T) {
		program := parseAST(t, "if true { 1 } else if false { 2 }")
		stmt := stmtAs[*IfStmt](t, program, 0)
		elseIf, ok := stmt.Else.(*IfStmt)
		if !ok {
			t.Fatalf("期望 Else 为 *IfStmt, 得到 %T", stmt.Else)
		}
		if elseIf.Condition == nil {
			t.Error("期望 else-if 的 Condition 非空")
		}
	})
}

func Test_ASTNode_ForStmt(t *testing.T) {
	t.Run("无限循环AST结构", func(t *testing.T) {
		program := parseAST(t, "for { break }")
		stmt := stmtAs[*ForStmt](t, program, 0)
		if stmt.Mode != ForWhile {
			t.Errorf("期望 ForWhile, 得到 %v", stmt.Mode)
		}
		if stmt.Body == nil {
			t.Fatal("期望 Body 非空")
		}
	})
	t.Run("条件循环AST结构", func(t *testing.T) {
		program := parseAST(t, "for true { break }")
		stmt := stmtAs[*ForStmt](t, program, 0)
		if stmt.Mode != ForWhile {
			t.Errorf("期望 ForWhile, 得到 %v", stmt.Mode)
		}
		if stmt.Cond == nil {
			t.Fatal("期望 Cond 非空")
		}
	})
	t.Run("标准for循环AST结构", func(t *testing.T) {
		program := parseAST(t, "for i := 0; i < 3; i = i + 1 { break }")
		stmt := stmtAs[*ForStmt](t, program, 0)
		if stmt.Mode != ForStandard {
			t.Errorf("期望 ForStandard, 得到 %v", stmt.Mode)
		}
		if stmt.Init == nil {
			t.Error("期望 Init 非空")
		}
		if stmt.Cond == nil {
			t.Error("期望 Cond 非空")
		}
		if stmt.Post == nil {
			t.Error("期望 Post 非空")
		}
	})
	t.Run("range循环AST结构", func(t *testing.T) {
		program := parseAST(t, "for k := range [1,2,3] { break }")
		stmt := stmtAs[*ForStmt](t, program, 0)
		if stmt.Mode != ForRange {
			t.Errorf("期望 ForRange, 得到 %v", stmt.Mode)
		}
		if stmt.Init == nil {
			t.Error("期望 Init 非空")
		}
	})
}

func Test_ASTNode_FuncDeclStmt(t *testing.T) {
	t.Run("函数定义AST结构", func(t *testing.T) {
		program := parseAST(t, "fn add(a, b) { return a + b }")
		stmt := stmtAs[*FuncDeclStmt](t, program, 0)
		if stmt.Name != "add" {
			t.Errorf("期望 Name=add, 得到 %s", stmt.Name)
		}
		if len(stmt.Params) != 2 {
			t.Fatalf("期望 2 个参数, 得到 %d", len(stmt.Params))
		}
		if stmt.Params[0].Name != "a" || stmt.Params[1].Name != "b" {
			t.Errorf("期望参数 a,b, 得到 %s,%s", stmt.Params[0].Name, stmt.Params[1].Name)
		}
		if stmt.Body == nil {
			t.Fatal("期望 Body 非空")
		}
	})
	t.Run("无参函数定义", func(t *testing.T) {
		program := parseAST(t, "fn f() { return 1 }")
		stmt := stmtAs[*FuncDeclStmt](t, program, 0)
		if len(stmt.Params) != 0 {
			t.Errorf("期望 0 个参数, 得到 %d", len(stmt.Params))
		}
	})
}

func Test_ASTNode_ReturnStmt(t *testing.T) {
	t.Run("return语句AST结构", func(t *testing.T) {
		program := parseAST(t, "fn f() { return 42 }")
		fnStmt := stmtAs[*FuncDeclStmt](t, program, 0)
		ret := fnStmt.Body.Statements[0].(*ReturnStmt)
		if ret.Value == nil {
			t.Fatal("期望返回值非空")
		}
		lit := ret.Value.(*LiteralExpr)
		if lit.Value.(int) != 42 {
			t.Errorf("期望 42, 得到 %v", lit.Value)
		}
	})
	t.Run("空return语句AST", func(t *testing.T) {
		program := parseAST(t, "fn f() { return }")
		fnStmt := stmtAs[*FuncDeclStmt](t, program, 0)
		ret := fnStmt.Body.Statements[0].(*ReturnStmt)
		if ret.Value != nil {
			t.Errorf("期望返回值为 nil, 得到 %T", ret.Value)
		}
	})
}

func Test_ASTNode_BreakStmt(t *testing.T) {
	t.Run("break语句AST结构", func(t *testing.T) {
		program := parseAST(t, "for { break }")
		forStmt := stmtAs[*ForStmt](t, program, 0)
		brk, ok := forStmt.Body.Statements[0].(*BreakStmt)
		if !ok {
			t.Fatalf("期望 *BreakStmt, 得到 %T", forStmt.Body.Statements[0])
		}
		_ = brk
	})
}

func Test_ASTNode_ContinueStmt(t *testing.T) {
	t.Run("continue语句AST结构", func(t *testing.T) {
		program := parseAST(t, "for { continue }")
		forStmt := stmtAs[*ForStmt](t, program, 0)
		cont, ok := forStmt.Body.Statements[0].(*ContinueStmt)
		if !ok {
			t.Fatalf("期望 *ContinueStmt, 得到 %T", forStmt.Body.Statements[0])
		}
		_ = cont
	})
}

func Test_ASTNode_ThrowStmt(t *testing.T) {
	t.Run("throw语句AST结构", func(t *testing.T) {
		program := parseAST(t, `throw "error"`)
		stmt := stmtAs[*ThrowStmt](t, program, 0)
		if stmt.Value == nil {
			t.Fatal("期望 Value 非空")
		}
		lit := stmt.Value.(*LiteralExpr)
		if lit.Value.(string) != "error" {
			t.Errorf("期望 error, 得到 %v", lit.Value)
		}
	})
}

// ========== 解析器行为验证 ==========

func Test_ASTNode_ProgramStructure(t *testing.T) {
	t.Run("Parse返回Program结构", func(t *testing.T) {
		program := parseAST(t, "1\n2\n3")
		if program == nil {
			t.Fatal("期望 Program 非空")
		}
		if len(program.Statements) != 3 {
			t.Errorf("期望 3 条语句, 得到 %d", len(program.Statements))
		}
	})
	t.Run("Program的Pos方法", func(t *testing.T) {
		program := parseAST(t, "42")
		pos := program.Pos()
		if pos.Line != 1 {
			t.Errorf("期望行号 1, 得到 %d", pos.Line)
		}
	})
}

func Test_ASTNode_StatementOrder(t *testing.T) {
	t.Run("语句序列解析顺序", func(t *testing.T) {
		program := parseAST(t, "x := 1\ny := 2\nz := 3")
		if len(program.Statements) != 3 {
			t.Fatalf("期望 3 条语句, 得到 %d", len(program.Statements))
		}
		s0 := program.Statements[0].(*VarDeclStmt)
		s1 := program.Statements[1].(*VarDeclStmt)
		s2 := program.Statements[2].(*VarDeclStmt)
		if s0.Name != "x" || s1.Name != "y" || s2.Name != "z" {
			t.Errorf("期望 x,y,z")
		}
	})
}

func Test_ASTNode_EmptyProgram(t *testing.T) {
	t.Run("空程序解析", func(t *testing.T) {
		program := parseAST(t, "")
		if len(program.Statements) != 0 {
			t.Errorf("期望 0 条语句, 得到 %d", len(program.Statements))
		}
	})
}

func Test_ASTNode_CommentOnly(t *testing.T) {
	t.Run("只有注释的程序解析", func(t *testing.T) {
		program := parseAST(t, "// comment\n# also comment")
		if len(program.Statements) != 0 {
			t.Errorf("期望 0 条语句, 得到 %d", len(program.Statements))
		}
	})
}

func Test_ASTNode_ValidatePass(t *testing.T) {
	t.Run("解析后Validate通过", func(t *testing.T) {
		parser := NewParser()
		if err := parser.Validate("x := 1\nif x > 0 { 1 }"); err != nil {
			t.Errorf("Validate 失败: %v", err)
		}
	})
}

// ========== 复杂AST结构验证 ==========

func Test_ASTNode_NestedIf(t *testing.T) {
	t.Run("嵌套if的AST深度", func(t *testing.T) {
		source := "if true { if true { if true { 1 } } }"
		program := parseAST(t, source)
		topIf := stmtAs[*IfStmt](t, program, 0)
		l1 := topIf.Then.Statements[0].(*IfStmt)
		l2 := l1.Then.Statements[0].(*IfStmt)
		if len(l2.Then.Statements) != 1 {
			t.Errorf("期望第三层有 1 条语句, 得到 %d", len(l2.Then.Statements))
		}
	})
}

func Test_ASTNode_NestedLoop(t *testing.T) {
	t.Run("嵌套循环的AST深度", func(t *testing.T) {
		source := "for { for { for { break } } }"
		program := parseAST(t, source)
		l1 := stmtAs[*ForStmt](t, program, 0)
		l2 := l1.Body.Statements[0].(*ForStmt)
		l3 := l2.Body.Statements[0].(*ForStmt)
		_, ok := l3.Body.Statements[0].(*BreakStmt)
		if !ok {
			t.Errorf("期望 BreakStmt, 得到 %T", l3.Body.Statements[0])
		}
	})
}

func Test_ASTNode_FuncBodyMultipleStmts(t *testing.T) {
	t.Run("函数体内多条语句", func(t *testing.T) {
		source := "fn f() { x := 1\ny := 2\nz := 3\nreturn x + y + z }"
		program := parseAST(t, source)
		fn := stmtAs[*FuncDeclStmt](t, program, 0)
		if len(fn.Body.Statements) != 4 {
			t.Fatalf("期望 4 条语句, 得到 %d", len(fn.Body.Statements))
		}
		if _, ok := fn.Body.Statements[0].(*VarDeclStmt); !ok {
			t.Error("第1条应为 VarDeclStmt")
		}
		if _, ok := fn.Body.Statements[3].(*ReturnStmt); !ok {
			t.Error("第4条应为 ReturnStmt")
		}
	})
}

func Test_ASTNode_ArrayNestedExpr(t *testing.T) {
	t.Run("数组内嵌套表达式", func(t *testing.T) {
		program := parseAST(t, "[1 + 2, 3 * 4]")
		stmt := stmtAs[*ExprStmt](t, program, 0)
		arr := stmt.Expr.(*ArrayExpr)
		if len(arr.Elements) != 2 {
			t.Fatalf("期望 2 个元素, 得到 %d", len(arr.Elements))
		}
		if _, ok := arr.Elements[0].(*BinaryExpr); !ok {
			t.Errorf("期望第1个元素为 BinaryExpr")
		}
	})
}

func Test_ASTNode_MapNestedExpr(t *testing.T) {
	t.Run("Map内嵌套表达式", func(t *testing.T) {
		program := parseAST(t, `{"a": 1 + 2, "b": 3 * 4}`)
		stmt := stmtAs[*ExprStmt](t, program, 0)
		m := stmt.Expr.(*MapExpr)
		if len(m.Pairs) != 2 {
			t.Fatalf("期望 2 对, 得到 %d", len(m.Pairs))
		}
		if _, ok := m.Pairs[0].Value.(*BinaryExpr); !ok {
			t.Errorf("期望第1个值为 BinaryExpr")
		}
	})
}

// ========== 解析器Token消费验证 ==========

func Test_ASTNode_TokenConsumption(t *testing.T) {
	t.Run("完整消费所有Token", func(t *testing.T) {
		program := parseAST(t, "1 + 2\n3 + 4")
		if len(program.Statements) != 2 {
			t.Errorf("期望 2 条语句, 得到 %d", len(program.Statements))
		}
	})
}

func Test_ASTNode_LeftoverTokens(t *testing.T) {
	t.Run("多余闭合括号报错", func(t *testing.T) {
		_, errs := parseASTWithErrs(t, "1 + 2)")
		if len(errs) == 0 {
			t.Error("期望有错误")
		}
	})
	t.Run("多余闭合花括号报错", func(t *testing.T) {
		_, errs := parseASTWithErrs(t, "1 }")
		if len(errs) == 0 {
			t.Error("期望有错误")
		}
	})
}

func Test_ASTNode_EOFHandling(t *testing.T) {
	t.Run("EOF正确处理", func(t *testing.T) {
		sources := []string{"42", "x := 1", "if true { 1 }", "[1,2,3]"}
		for _, src := range sources {
			program := parseAST(t, src)
			if program == nil {
				t.Errorf("[%s] Program 为空", src)
			}
		}
	})
}

func Test_ASTNode_ExtraCloseBrace(t *testing.T) {
	t.Run("多余闭合符号处理", func(t *testing.T) {
		_, errs := parseASTWithErrs(t, "1 2 3 }")
		if len(errs) == 0 {
			t.Error("期望有多余闭合符号错误")
		}
	})
}

// ========== 类型注解解析验证 ==========

func Test_ASTNode_TypedAssign(t *testing.T) {
	t.Run("类型安全声明AST结构", func(t *testing.T) {
		program := parseAST(t, "x :=>int 10")
		stmt := stmtAs[*VarDeclStmt](t, program, 0)
		if stmt.Name != "x" {
			t.Errorf("期望 Name=x, 得到 %s", stmt.Name)
		}
		if stmt.TypeAnnot == nil {
			t.Fatal("期望 TypeAnnot 非空")
		}
		baseType := stmt.TypeAnnot.(*BaseTypeExpr)
		if baseType.Name != "int" {
			t.Errorf("期望类型名 int, 得到 %s", baseType.Name)
		}
	})
	t.Run("类型安全声明执行", func(t *testing.T) {
		runIntTest(t, "x :=>int 10\nx", 10)
	})
}

func Test_ASTNode_TypeAnnotation(t *testing.T) {
	t.Run("float类型注解", func(t *testing.T) {
		program := parseAST(t, "x :=>float 3.14")
		stmt := stmtAs[*VarDeclStmt](t, program, 0)
		baseType := stmt.TypeAnnot.(*BaseTypeExpr)
		if baseType.Name != "float" {
			t.Errorf("期望 float, 得到 %s", baseType.Name)
		}
	})
	t.Run("string类型注解", func(t *testing.T) {
		program := parseAST(t, "x :=>string \"hello\"")
		stmt := stmtAs[*VarDeclStmt](t, program, 0)
		baseType := stmt.TypeAnnot.(*BaseTypeExpr)
		if baseType.Name != "string" {
			t.Errorf("期望 string, 得到 %s", baseType.Name)
		}
	})
	t.Run("类型别名解析", func(t *testing.T) {
		program := parseAST(t, "x :=>i 10")
		stmt := stmtAs[*VarDeclStmt](t, program, 0)
		baseType := stmt.TypeAnnot.(*BaseTypeExpr)
		if baseType.Name != "int" {
			t.Errorf("期望 int (别名i), 得到 %s", baseType.Name)
		}
	})
}

func Test_ASTNode_TypeAnnotationWithValue(t *testing.T) {
	t.Run("类型注解加表达式值", func(t *testing.T) {
		program := parseAST(t, "x :=>int 10 + 20")
		stmt := stmtAs[*VarDeclStmt](t, program, 0)
		if stmt.TypeAnnot == nil {
			t.Fatal("期望 TypeAnnot 非空")
		}
		bin, ok := stmt.Value.(*BinaryExpr)
		if !ok {
			t.Fatalf("期望 Value 为 BinaryExpr, 得到 %T", stmt.Value)
		}
		if bin.Operator != "+" {
			t.Errorf("期望 +, 得到 %s", bin.Operator)
		}
	})
	t.Run("类型注解加表达式执行", func(t *testing.T) {
		runIntTest(t, "x :=>int 10 + 20\nx", 30)
	})
}

// ========== #fn指令解析验证 ==========

func Test_ASTNode_FnDirective(t *testing.T) {
	t.Run("#fn声明解析为DefDirectiveStmt", func(t *testing.T) {
		program := parseAST(t, "#fn foo(int)=>int\n10")
		stmt := stmtAs[*DefDirectiveStmt](t, program, 0)
		if stmt.Name != "foo" {
			t.Errorf("期望 Name=foo, 得到 %s", stmt.Name)
		}
	})
}

func Test_ASTNode_FnDirectiveParams(t *testing.T) {
	t.Run("#fn单参数列表", func(t *testing.T) {
		program := parseAST(t, "#fn foo(int)=>int\n10")
		stmt := stmtAs[*DefDirectiveStmt](t, program, 0)
		if len(stmt.Params) != 1 {
			t.Fatalf("期望 1 个参数, 得到 %d", len(stmt.Params))
		}
		baseType := stmt.Params[0].TypeAnnot.(*BaseTypeExpr)
		if baseType.Name != "int" {
			t.Errorf("期望参数类型 int, 得到 %s", baseType.Name)
		}
	})
	t.Run("#fn多参数列表", func(t *testing.T) {
		program := parseAST(t, "#fn add(int, int)=>int\n10")
		stmt := stmtAs[*DefDirectiveStmt](t, program, 0)
		if len(stmt.Params) != 2 {
			t.Fatalf("期望 2 个参数, 得到 %d", len(stmt.Params))
		}
	})
	t.Run("#fn无参数", func(t *testing.T) {
		program := parseAST(t, "#fn bar()=>int\n10")
		stmt := stmtAs[*DefDirectiveStmt](t, program, 0)
		if len(stmt.Params) != 0 {
			t.Errorf("期望 0 个参数, 得到 %d", len(stmt.Params))
		}
	})
}

func Test_ASTNode_FnDirectiveReturnType(t *testing.T) {
	t.Run("#fn返回类型解析", func(t *testing.T) {
		program := parseAST(t, "#fn foo(int)=>string\n10")
		stmt := stmtAs[*DefDirectiveStmt](t, program, 0)
		if stmt.Return == nil {
			t.Fatal("期望 Return 非空")
		}
		baseType := stmt.Return.(*BaseTypeExpr)
		if baseType.Name != "string" {
			t.Errorf("期望返回类型 string, 得到 %s", baseType.Name)
		}
	})
	t.Run("#fn返回bool类型", func(t *testing.T) {
		program := parseAST(t, "#fn check(int)=>bool\n10")
		stmt := stmtAs[*DefDirectiveStmt](t, program, 0)
		baseType := stmt.Return.(*BaseTypeExpr)
		if baseType.Name != "bool" {
			t.Errorf("期望返回类型 bool, 得到 %s", baseType.Name)
		}
	})
}

func Test_ASTNode_MultipleFnDirectives(t *testing.T) {
	t.Run("多个#fn声明", func(t *testing.T) {
		source := "#fn foo(int)=>int\n#fn bar(string)=>bool\n10"
		program := parseAST(t, source)
		if len(program.Statements) < 3 {
			t.Fatalf("期望至少 3 条语句, 得到 %d", len(program.Statements))
		}
		d0 := stmtAs[*DefDirectiveStmt](t, program, 0)
		d1 := stmtAs[*DefDirectiveStmt](t, program, 1)
		if d0.Name != "foo" {
			t.Errorf("期望第1个函数名 foo, 得到 %s", d0.Name)
		}
		if d1.Name != "bar" {
			t.Errorf("期望第2个函数名 bar, 得到 %s", d1.Name)
		}
	})
}

// ========== 错误恢复验证 ==========

func Test_ASTNode_ErrorPosition(t *testing.T) {
	t.Run("解析错误包含行号", func(t *testing.T) {
		_, errs := parseASTWithErrs(t, "1 + + 2")
		if len(errs) == 0 {
			t.Fatal("期望有错误")
		}
		if errs[0].Line <= 0 {
			t.Errorf("期望行号 > 0, 得到 %d", errs[0].Line)
		}
	})
	t.Run("解析错误包含列号", func(t *testing.T) {
		_, errs := parseASTWithErrs(t, "1 + + 2")
		if len(errs) == 0 {
			t.Fatal("期望有错误")
		}
		if errs[0].Column <= 0 {
			t.Errorf("期望列号 > 0, 得到 %d", errs[0].Column)
		}
	})
	t.Run("解析错误包含消息", func(t *testing.T) {
		_, errs := parseASTWithErrs(t, "1 + + 2")
		if len(errs) == 0 {
			t.Fatal("期望有错误")
		}
		if errs[0].Message == "" {
			t.Error("期望错误消息非空")
		}
	})
}

func Test_ASTNode_PartialParseAST(t *testing.T) {
	t.Run("部分解析的AST结构", func(t *testing.T) {
		program, errs := parseASTWithErrs(t, "1 + + 2")
		if len(errs) == 0 {
			t.Fatal("期望有错误")
		}
		if len(program.Statements) < 1 {
			t.Fatal("期望至少 1 条语句")
		}
	})
}

func Test_ASTNode_ConsecutiveErrors(t *testing.T) {
	t.Run("连续错误检测", func(t *testing.T) {
		_, errs := parseASTWithErrs(t, "1 + + 2 + + 3")
		if len(errs) < 2 {
			t.Errorf("期望至少 2 个错误, 得到 %d", len(errs))
		}
	})
}

// ========== Position信息验证 ==========

func Test_ASTNode_PositionLine(t *testing.T) {
	t.Run("AST节点保留行号", func(t *testing.T) {
		program := parseAST(t, "42")
		pos := program.Statements[0].Pos()
		if pos.Line != 1 {
			t.Errorf("期望行号 1, 得到 %d", pos.Line)
		}
	})
	t.Run("第二行节点的行号", func(t *testing.T) {
		program := parseAST(t, "1\n42")
		pos := program.Statements[1].Pos()
		if pos.Line != 2 {
			t.Errorf("期望行号 2, 得到 %d", pos.Line)
		}
	})
}

func Test_ASTNode_PositionColumn(t *testing.T) {
	t.Run("AST节点保留列号", func(t *testing.T) {
		program := parseAST(t, "42")
		pos := program.Statements[0].Pos()
		if pos.Column != 1 {
			t.Errorf("期望列号 1, 得到 %d", pos.Column)
		}
	})
	t.Run("缩进后列号正确", func(t *testing.T) {
		program := parseAST(t, "  42")
		pos := program.Statements[0].Pos()
		if pos.Column != 3 {
			t.Errorf("期望列号 3, 得到 %d", pos.Column)
		}
	})
}

func Test_ASTNode_MultiLinePosition(t *testing.T) {
	t.Run("多行节点的起始位置", func(t *testing.T) {
		source := "x := 1\nif x > 0 {\n  1\n}"
		program := parseAST(t, source)
		ifStmt := program.Statements[1].(*IfStmt)
		pos := ifStmt.Pos()
		if pos.Line != 2 {
			t.Errorf("期望行号 2, 得到 %d", pos.Line)
		}
	})
	t.Run("变量声明的位置", func(t *testing.T) {
		source := "\n\nx := 42"
		program := parseAST(t, source)
		pos := program.Statements[0].Pos()
		if pos.Line != 3 {
			t.Errorf("期望行号 3, 得到 %d", pos.Line)
		}
	})
}
