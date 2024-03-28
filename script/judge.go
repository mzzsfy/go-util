package script

type judge interface {
    expression
}

type ifJudge struct {
    offset_   int
    condition func() bool
    trueBody  func() any
    falseBody func() any
}

func (f ifJudge) offset() int {
    return f.offset_
}

func (f ifJudge) compute() any {
    if f.condition() {
        return f.trueBody()
    }
    return f.falseBody()
}

type ternaryJudge struct {
    ifJudge
}
