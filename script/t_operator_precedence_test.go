package script

import "testing"

// ========== 运算符优先级完整测试 ==========
// 覆盖所有运算符的优先级与结合性, 以引擎实际行为为准
// 引擎优先级(从高到低):
//   PREFIX(10): 一元 -, !, ^
//   PRODUCT(9): *, /, %
//   SUM(8): +, -
//   BITOP(7): |, ^, &, <<, >> (同优先级, 左结合)
//   LESSGREATER(6): <, >, <=, >=
//   EQUALS(5): ==, !=
//   AND(4): &&
//   OR(3): ||
//   ASSIGN(2): =, +=, -=, *=, /=

// ========== 1. 括号优先级 ==========

func Test_OperatorPrecedence_Parentheses(t *testing.T) {
	// 括号改变运算顺序
	t.Run("(1+2)*3=9", func(t *testing.T) {
		runIntTest(t, "(1 + 2) * 3", 9)
	})
	t.Run("(2+3)*(4+5)=45", func(t *testing.T) {
		runIntTest(t, "(2 + 3) * (4 + 5)", 45)
	})
	t.Run("2*(3+4)=14", func(t *testing.T) {
		runIntTest(t, "2 * (3 + 4)", 14)
	})
	t.Run("((1+2)+3)+4=10", func(t *testing.T) {
		runIntTest(t, "((1 + 2) + 3) + 4", 10)
	})
	t.Run("(10-2)*(10-2)=64", func(t *testing.T) {
		runIntTest(t, "(10 - 2) * (10 - 2)", 64)
	})
	t.Run("(1+2)*(3+4)-(5+6)=10", func(t *testing.T) {
		runIntTest(t, "((1 + 2) * (3 + 4)) - (5 + 6)", 10)
	})
}

// ========== 2. 一元运算符优先级 ==========

func Test_OperatorPrecedence_UnaryMinus(t *testing.T) {
	// 一元负号优先于乘除和加减
	t.Run("-1+2=1", func(t *testing.T) {
		runIntTest(t, "-1 + 2", 1)
	})
	t.Run("-(1+2)=-3", func(t *testing.T) {
		runIntTest(t, "-(1 + 2)", -3)
	})
	t.Run("-2*3=-6", func(t *testing.T) {
		runIntTest(t, "-2 * 3", -6)
	})
	t.Run("2*-3=-6", func(t *testing.T) {
		// 前缀运算优先于乘法
		runIntTest(t, "2 * -3", -6)
	})
	t.Run("-(-5)=5", func(t *testing.T) {
		runIntTest(t, "-(-5)", 5)
	})
}

func Test_OperatorPrecedence_UnaryNot(t *testing.T) {
	// 逻辑非优先于比较运算
	t.Run("!true==false", func(t *testing.T) {
		// PREFIX > EQUALS: (!true) == false = true
		runBoolTest(t, "!true == false", true)
	})
	t.Run("!false==true", func(t *testing.T) {
		runBoolTest(t, "!false == true", true)
	})
	t.Run("!true&&!false=false", func(t *testing.T) {
		// (!true) && (!false) = false && true = false
		runBoolTest(t, "!true && !false", false)
	})
	t.Run("!(1>2)=true", func(t *testing.T) {
		runBoolTest(t, "!(1 > 2)", true)
	})
}

func Test_OperatorPrecedence_UnaryBitNot(t *testing.T) {
	// 一元 ^ 为位取反
	t.Run("^5=-6", func(t *testing.T) {
		runIntTest(t, "^5", -6)
	})
	t.Run("^0=-1", func(t *testing.T) {
		runIntTest(t, "^0", -1)
	})
	t.Run("^5+1=-5", func(t *testing.T) {
		// PREFIX > SUM: (^5)+1 = -6+1 = -5
		runIntTest(t, "^5 + 1", -5)
	})
}

// ========== 3. 乘除模优先级 ==========

func Test_OperatorPrecedence_MulDiv(t *testing.T) {
	// 乘除模优先于加减
	t.Run("1+2*3=7", func(t *testing.T) {
		runIntTest(t, "1 + 2 * 3", 7)
	})
	t.Run("2*3+1=7", func(t *testing.T) {
		runIntTest(t, "2 * 3 + 1", 7)
	})
	t.Run("10-6/2=7", func(t *testing.T) {
		runIntTest(t, "10 - 6 / 2", 7)
	})
	t.Run("10/2+3=8", func(t *testing.T) {
		runIntTest(t, "10 / 2 + 3", 8)
	})
	t.Run("2+3*4-1=13", func(t *testing.T) {
		runIntTest(t, "2 + 3 * 4 - 1", 13)
	})
	t.Run("100/5/2=10", func(t *testing.T) {
		// 左结合: (100/5)/2=10
		runIntTest(t, "100 / 5 / 2", 10)
	})
	t.Run("2*3*4=24", func(t *testing.T) {
		runIntTest(t, "2 * 3 * 4", 24)
	})
}

func Test_OperatorPrecedence_Modulo(t *testing.T) {
	// 取模与乘除同级
	t.Run("10%3+1=2", func(t *testing.T) {
		runIntTest(t, "10 % 3 + 1", 2)
	})
	t.Run("1+10%3=2", func(t *testing.T) {
		runIntTest(t, "1 + 10 % 3", 2)
	})
	t.Run("10%(3+1)=2", func(t *testing.T) {
		runIntTest(t, "10 % (3 + 1)", 2)
	})
	t.Run("10%3*2=2", func(t *testing.T) {
		// 同级左结合: (10%3)*2 = 1*2 = 2
		runIntTest(t, "10 % 3 * 2", 2)
	})
}

// ========== 4. 加减优先级 ==========

func Test_OperatorPrecedence_AddSub(t *testing.T) {
	// 加减同级, 左结合
	t.Run("1+2+3=6", func(t *testing.T) {
		runIntTest(t, "1 + 2 + 3", 6)
	})
	t.Run("1+2-3=0", func(t *testing.T) {
		runIntTest(t, "1 + 2 - 3", 0)
	})
	t.Run("10-3-2=5", func(t *testing.T) {
		runIntTest(t, "10 - 3 - 2", 5)
	})
	t.Run("1-2+3=2", func(t *testing.T) {
		runIntTest(t, "1 - 2 + 3", 2)
	})
}

// ========== 5. 位运算优先级(全部同级, 左结合) ==========

func Test_OperatorPrecedence_BitwiseSameLevel(t *testing.T) {
	// 引擎中所有位运算符优先级相同, 左结合
	t.Run("1|2&3=3", func(t *testing.T) {
		// (1|2)&3 = 3&3 = 3
		runIntTest(t, "1 | 2 & 3", 3)
	})
	t.Run("1&2|3=3", func(t *testing.T) {
		// (1&2)|3 = 0|3 = 3
		runIntTest(t, "1 & 2 | 3", 3)
	})
	t.Run("1^2&3=3", func(t *testing.T) {
		// (1^2)&3 = 3&3 = 3
		runIntTest(t, "1 ^ 2 & 3", 3)
	})
	t.Run("1<<2&3=0", func(t *testing.T) {
		// (1<<2)&3 = 4&3 = 0
		runIntTest(t, "1 << 2 & 3", 0)
	})
	t.Run("1|2<<2=12", func(t *testing.T) {
		// (1|2)<<2 = 3<<2 = 12
		runIntTest(t, "1 | 2 << 2", 12)
	})
}

func Test_OperatorPrecedence_BitwiseAndArithmetic(t *testing.T) {
	// 算术(SUM/PRODUCT)优先于位运算(BITOP)
	t.Run("1<<2+3=32", func(t *testing.T) {
		// + 优先于 <<: 1<<(2+3) = 1<<5 = 32
		runIntTest(t, "1 << 2 + 3", 32)
	})
	t.Run("1+2<<3=24", func(t *testing.T) {
		// + 优先于 <<: (1+2)<<3 = 3<<3 = 24
		runIntTest(t, "1 + 2 << 3", 24)
	})
	t.Run("(1<<2)+3=7", func(t *testing.T) {
		// 括号强制先移位: 4+3=7
		runIntTest(t, "(1 << 2) + 3", 7)
	})
	t.Run("1&2==0=true", func(t *testing.T) {
		// BITOP > EQUALS: (1&2)==0 -> true
		runBoolTest(t, "1 & 2 == 0", true)
	})
}

func Test_OperatorPrecedence_BitwiseConstants(t *testing.T) {
	t.Run("0xFF&0x0F=15", func(t *testing.T) {
		runIntTest(t, "0xFF & 0x0F", 15)
	})
	t.Run("0xF0|0x0F=255", func(t *testing.T) {
		runIntTest(t, "0xF0 | 0x0F", 255)
	})
	t.Run("0xAA^0xFF=85", func(t *testing.T) {
		runIntTest(t, "0xAA ^ 0xFF", 85)
	})
}

// ========== 6. 移位运算 ==========

func Test_OperatorPrecedence_Shift(t *testing.T) {
	t.Run("1<<4=16", func(t *testing.T) {
		runIntTest(t, "1 << 4", 16)
	})
	t.Run("256>>4=16", func(t *testing.T) {
		runIntTest(t, "256 >> 4", 16)
	})
	t.Run("1<<2>3=true", func(t *testing.T) {
		// BITOP > LESSGREATER: (1<<2)>3 -> 4>3 = true
		runBoolTest(t, "1 << 2 > 3", true)
	})
}

// ========== 7. 比较运算优先级 ==========

func Test_OperatorPrecedence_Comparison(t *testing.T) {
	// 比较优先于相等比较
	// 算术优先于比较
	t.Run("1+1==2=true", func(t *testing.T) {
		runBoolTest(t, "1 + 1 == 2", true)
	})
	t.Run("2*2>3=true", func(t *testing.T) {
		runBoolTest(t, "2 * 2 > 3", true)
	})
	t.Run("10%3==1=true", func(t *testing.T) {
		runBoolTest(t, "10 % 3 == 1", true)
	})
	t.Run("1<2==true=true", func(t *testing.T) {
		// LESSGREATER > EQUALS: (1<2)==true -> true
		runBoolTest(t, "1 < 2 == true", true)
	})
}

func Test_OperatorPrecedence_EqualityVsBitwise(t *testing.T) {
	// 位运算优先于相等比较
	t.Run("3&1==1=true", func(t *testing.T) {
		// (3&1)==1 -> 1==1 -> true
		runBoolTest(t, "3 & 1 == 1", true)
	})
	t.Run("5|2!=5=true", func(t *testing.T) {
		// (5|2)!=5 -> 7!=5 -> true
		runBoolTest(t, "5 | 2 != 5", true)
	})
}

// ========== 8. 逻辑运算优先级 ==========

func Test_OperatorPrecedence_LogicalAndOr(t *testing.T) {
	// && 优先于 ||
	t.Run("true&&false||true=true", func(t *testing.T) {
		// true && (false||true)? 不, AND > OR: (true&&false)||true = false||true = true
		runBoolTest(t, "true && false || true", true)
	})
	t.Run("true||false&&false=true", func(t *testing.T) {
		// AND > OR: true || (false&&false) = true||false = true
		runBoolTest(t, "true || false && false", true)
	})
}

func Test_OperatorPrecedence_LogicalVsComparison(t *testing.T) {
	// 比较优先于逻辑
	t.Run("1>0&&2>1=true", func(t *testing.T) {
		runBoolTest(t, "1 > 0 && 2 > 1", true)
	})
	t.Run("1>2||3>2=true", func(t *testing.T) {
		runBoolTest(t, "1 > 2 || 3 > 2", true)
	})
	t.Run("1>0&&2>1&&3>2=true", func(t *testing.T) {
		runBoolTest(t, "1 > 0 && 2 > 1 && 3 > 2", true)
	})
	t.Run("1>0||2>1&&false=true", func(t *testing.T) {
		// AND > OR: 1>0 || (2>1 && false) = true || false = true
		runBoolTest(t, "1 > 0 || 2 > 1 && false", true)
	})
}

// ========== 9. 结合性测试 ==========

func Test_OperatorPrecedence_LeftAssociativity(t *testing.T) {
	// 左结合: a-b-c = (a-b)-c
	t.Run("100-50-25=25", func(t *testing.T) {
		runIntTest(t, "100 - 50 - 25", 25)
	})
	t.Run("100/5/2=10", func(t *testing.T) {
		runIntTest(t, "100 / 5 / 2", 10)
	})
	t.Run("2-3-4=-5", func(t *testing.T) {
		runIntTest(t, "2 - 3 - 4", -5)
	})
	t.Run("1+2-3+4-5=-1", func(t *testing.T) {
		runIntTest(t, "1 + 2 - 3 + 4 - 5", -1)
	})
}

func Test_OperatorPrecedence_MixedChain(t *testing.T) {
	// 同级混合运算的左结合行为
	t.Run("2*3/4*5=7", func(t *testing.T) {
		// ((2*3)/4)*5 = (6/4)*5 = 1*5 = 5
		runIntTest(t, "2 * 3 / 4 * 5", 5)
	})
	t.Run("12/3*2=8", func(t *testing.T) {
		// (12/3)*2 = 4*2 = 8
		runIntTest(t, "12 / 3 * 2", 8)
	})
}

// ========== 10. 混合复杂表达式 ==========

func Test_OperatorPrecedence_MixedArithmetic(t *testing.T) {
	t.Run("1+2*3-4/2=5", func(t *testing.T) {
		// 1+6-2=5
		runIntTest(t, "1 + 2 * 3 - 4 / 2", 5)
	})
	t.Run("2*(3+4)-1=13", func(t *testing.T) {
		runIntTest(t, "2 * (3 + 4) - 1", 13)
	})
	t.Run("((1+2)*3)-4=5", func(t *testing.T) {
		runIntTest(t, "((1 + 2) * 3) - 4", 5)
	})
	t.Run("2*3+4*5+6*7=68", func(t *testing.T) {
		// 6+20+42=68
		runIntTest(t, "2 * 3 + 4 * 5 + 6 * 7", 68)
	})
	t.Run("((2+3)*(4+5))/(1+2)=15", func(t *testing.T) {
		// (5*9)/3 = 45/3 = 15
		runIntTest(t, "((2 + 3) * (4 + 5)) / (1 + 2)", 15)
	})
	t.Run("100/10/2+3*4=17", func(t *testing.T) {
		// (100/10)/2=5, 3*4=12, 5+12=17
		runIntTest(t, "100 / 10 / 2 + 3 * 4", 17)
	})
}

func Test_OperatorPrecedence_BitwiseComplex(t *testing.T) {
	// 位运算与算术混合
	t.Run("1+2|4=7", func(t *testing.T) {
		// SUM > BITOP: (1+2)|4 = 3|4 = 7
		runIntTest(t, "1 + 2 | 4", 7)
	})
	t.Run("8-4&3=0", func(t *testing.T) {
		// SUM > BITOP: (8-4)&3 = 4&3 = 0
		runIntTest(t, "8 - 4 & 3", 0)
	})
	t.Run("2*3^1=7", func(t *testing.T) {
		// PRODUCT > BITOP: (2*3)^1 = 6^1 = 7
		runIntTest(t, "2 * 3 ^ 1", 7)
	})
}

func Test_OperatorPrecedence_LogicalComplex(t *testing.T) {
	// 逻辑运算与比较混合
	t.Run("(1>0)&&(2<3)=true", func(t *testing.T) {
		runBoolTest(t, "(1 > 0) && (2 < 3)", true)
	})
	t.Run("1+1==2&&2*2==4=true", func(t *testing.T) {
		// (1+1==2) && (2*2==4) = true && true
		runBoolTest(t, "1 + 1 == 2 && 2 * 2 == 4", true)
	})
	t.Run("1>2||2>1&&3>2=true", func(t *testing.T) {
		// AND > OR: 1>2 || (2>1 && 3>2) = false || true = true
		runBoolTest(t, "1 > 2 || 2 > 1 && 3 > 2", true)
	})
}

// ========== 11. 浮点数运算优先级 ==========

func Test_OperatorPrecedence_FloatArithmetic(t *testing.T) {
	t.Run("1.5+2.5*2.0=6.5", func(t *testing.T) {
		runFloatTest(t, "1.5 + 2.5 * 2.0", 6.5)
	})
	t.Run("(1.5+2.5)*2.0=8.0", func(t *testing.T) {
		runFloatTest(t, "(1.5 + 2.5) * 2.0", 8.0)
	})
	t.Run("10.0/3.0", func(t *testing.T) {
		runFloatTest(t, "10.0 / 3.0", 3.3333333333333335)
	})
	t.Run("2.0*3.0+4.0=10.0", func(t *testing.T) {
		runFloatTest(t, "2.0 * 3.0 + 4.0", 10.0)
	})
}

// ========== 12. 整数除法 ==========

func Test_OperatorPrecedence_IntDivision(t *testing.T) {
	t.Run("7/2=3", func(t *testing.T) {
		runIntTest(t, "7 / 2", 3)
	})
	t.Run("-7/2=-3", func(t *testing.T) {
		runIntTest(t, "-7 / 2", -3)
	})
	t.Run("7/-2=-3", func(t *testing.T) {
		runIntTest(t, "7 / -2", -3)
	})
	t.Run("0/5=0", func(t *testing.T) {
		runIntTest(t, "0 / 5", 0)
	})
}

// ========== 13. 综合优先级链 ==========

func Test_OperatorPrecedence_FullChain(t *testing.T) {
	// 从最低到最高优先级的综合验证
	t.Run("arithmetic_full", func(t *testing.T) {
		// 1 + 2*3 - 4/2 + 5%3 = 1+6-2+2 = 7
		runIntTest(t, "1 + 2 * 3 - 4 / 2 + 5 % 3", 7)
	})
	t.Run("bitwise_full", func(t *testing.T) {
		// 1 | 2 ^ 3 & 4 = (1|2)^(3&4)... no, left-to-right
		// ((1|2)^3)&4 = (3^3)&4 = 0&4 = 0
		runIntTest(t, "1 | 2 ^ 3 & 4", 0)
	})
	t.Run("shift_arithmetic_bit", func(t *testing.T) {
		// 1 + 2 << 3 | 4
		// SUM > BITOP: (1+2) << (3|4)? No, BITOP all same level, left-to-right
		// (1+2) << 3 | 4: ((3<<3)|4) = 24|4 = 28
		runIntTest(t, "1 + 2 << 3 | 4", 28)
	})
	t.Run("logical_full", func(t *testing.T) {
		// 1+1==2 && 3>2 || 5<1
		// AND > OR: ((1+1==2)&&(3>2)) || (5<1)
		// = (true && true) || false = true
		runBoolTest(t, "1 + 1 == 2 && 3 > 2 || 5 < 1", true)
	})
}

// ========== 14. 括号嵌套深度 ==========

func Test_OperatorPrecedence_DeepNesting(t *testing.T) {
	t.Run("deep_3_levels", func(t *testing.T) {
		// (((1+2)))
		runIntTest(t, "(((1 + 2)))", 3)
	})
	t.Run("deep_mixed", func(t *testing.T) {
		// ((1+(2*3))-(4/2)) = (1+6-2) = 5
		runIntTest(t, "((1 + (2 * 3)) - (4 / 2))", 5)
	})
	t.Run("paren_override_prec", func(t *testing.T) {
		// 括号覆盖默认优先级: (2+3)*(4+5) vs 2+3*4+5
		runIntTest(t, "(2 + 3) * (4 + 5)", 45)
		runIntTest(t, "2 + 3 * 4 + 5", 19)
	})
}

// ========== 15. 前缀运算与中缀混合 ==========

func Test_OperatorPrecedence_PrefixInfix(t *testing.T) {
	t.Run("neg_in_sum", func(t *testing.T) {
		// -2 + 3 * 4 = -2 + 12 = 10
		runIntTest(t, "-2 + 3 * 4", 10)
	})
	t.Run("neg_in_product", func(t *testing.T) {
		// 3 * -4 + 2 = -12 + 2 = -10
		runIntTest(t, "3 * -4 + 2", -10)
	})
	t.Run("not_in_logical", func(t *testing.T) {
		// !false && true = true && true = true
		runBoolTest(t, "!false && true", true)
	})
	t.Run("bitnot_in_sum", func(t *testing.T) {
		// ^0 + 1 = -1 + 1 = 0
		runIntTest(t, "^0 + 1", 0)
	})
	t.Run("double_neg", func(t *testing.T) {
		// - -5 = 5, then 5 * 2 = 10
		runIntTest(t, "- -5 * 2", 10)
	})
}

// ========== 16. 比较运算符之间的关系 ==========

func Test_OperatorPrecedence_LessGreaterVsEquals(t *testing.T) {
	// LESSGREATER 与 EQUALS 的优先级关系
	t.Run("lt_eq_true", func(t *testing.T) {
		// (1<2)==true
		runBoolTest(t, "1 < 2 == true", true)
	})
	t.Run("gt_neq", func(t *testing.T) {
		// (3>2)!=false -> true
		runBoolTest(t, "3 > 2 != false", true)
	})
	t.Run("le_eq", func(t *testing.T) {
		// (2<=2)==true
		runBoolTest(t, "2 <= 2 == true", true)
	})
	t.Run("ge_eq", func(t *testing.T) {
		// (3>=4)==false
		runBoolTest(t, "3 >= 4 == false", true)
	})
}
