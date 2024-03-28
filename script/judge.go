package script

type judge interface {
    expression
}

type ifJudge struct {
    offset    int
    condition func(Scope) bool
    trueBody  func(Scope) any
    falseBody func(Scope) any
}

func (f ifJudge) compute(scope Scope) any {
    if f.condition(scope) {
        return f.trueBody(scope)
    }
    return f.falseBody(scope)
}

type ternaryJudge struct {
    ifJudge
}
