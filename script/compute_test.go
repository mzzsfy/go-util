package script

import (
    "fmt"
    "github.com/mzzsfy/go-util/seq"
    "sync/atomic"
    "testing"
)

func TestParseNumberExpr(t *testing.T) {
    exprs := []string{
        "1.1+3.2",
        ".0000001+3",
        ".1111112+1999.000000001",
        "3+3",
        "3+3-3",
        "3.0+3*3",
        "(3+3)*3",
        "(3+3)*3+3",
        "(3+3)*(3+3)",
        "((3+3)+(.3+3))*3",
        "3*((3+3)+(3+3))",
        "3-((3+3)+(3+3))",
        "((3+3)+(3+3))-3",
        "((3+3)+(3+3)-3)+3",
        "3+((3+3)+(3+3)-3)",
        "3+((3+3)+(3+3)-3)*3",
        "3+((3+3)+(3+3)-3)*3",
        "3+3+3+(3-(3+3)+(3+3)-3)*3",
        "3*3+3+(3-(3+3)+(3+3)-3)*3",
        "3*3+3*(3-(3+3)+(3+3)-3)*3",
        "3*3*3+((3+3)+(3+3)-3)*3",
        "3*3+3*((3+3)+(3+3)-3)*3+3*3",
        "3*3+3*((3+3)+(3+3)-3)*(3+3)*3",
    }
    idGen := int32(0)
    exprs = append(exprs, fmt.Sprintf(
        "-%d+%d+((-(%d*%d))+(%d*((%d+%d)*%d))*(%d*%d+%d%%%d*(%d*%d*%d-%d+%d-%d%%%d-%d-%d)+%d+(%d-%d)*%d)-%d)+%d*%d*%d+%d+%d",
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
    ))
    exprs = append(exprs, fmt.Sprintf(
        "-%d+%d+(-(%d*(.%d+%d)*%d)*(%d/%d+%d%%%d*(%d*%d*%d-%d+%d-%d%%%d-%d-%d)+%d+(%d-%d)*%d)-%d)+%d*%d*%d+%d+%d",
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
    ))
    exprs = append(exprs, "1+2+a+b")
    scope := &globalScope{v: map[string]any{"a": 1, "b": 2}}
    for _, exprStr := range exprs {
        t.Log(exprStr)
        expr := parseNumberExpr([]rune(exprStr), 0)
        s1 := fmt.Sprint(seq.From(func(t func(any)) {
            node := expr.expr
            for {
                if node.next == nil {
                    break
                }
                t(node)
                node = *node.next
            }
        }).ToSlice())
        t.Log(expr.compute(scope), s1)
        expr.optimization()
        i := 0
        n := expr.expr
        for {
            if n.next == nil {
                break
            }
            n = *n.next
            i++
        }
        t.Log(expr.compute(scope), i)
    }
}

func BenchmarkParseNumberExpr(b *testing.B) {
    expr := parseNumberExpr([]rune("3+3"), 0)
    b.Run("noOptimization", func(b *testing.B) {
        scope := globalScope{}
        for i := 0; i < b.N; i++ {
            expr.compute(scope)
        }
    })
    b.Run("optimization", func(b *testing.B) {
        scope := globalScope{}
        expr.optimization()
        for i := 0; i < b.N; i++ {
            expr.compute(scope)
        }
    })
    expr1 := parseNumberExpr([]rune("3+3+a"), 0)
    b.Run("noOptimization1", func(b *testing.B) {
        scope := globalScope{v: map[string]any{
            "a": 1,
        }}
        for i := 0; i < b.N; i++ {
            expr1.compute(scope)
        }
    })
    b.Run("optimization1", func(b *testing.B) {
        scope := globalScope{v: map[string]any{
            "a": 1,
        }}
        expr.optimization()
        for i := 0; i < b.N; i++ {
            expr1.compute(scope)
        }
    })
}
