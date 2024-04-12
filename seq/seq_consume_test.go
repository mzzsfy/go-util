package seq

import "testing"

func Test_Match(t *testing.T) {
    {
        x := 0
        if !FromIntSeq().Skip(1).Take(3).Cache().AnyMatch(func(i int) bool {
            x++
            return i == 2
        }) {
            t.Error("AnyMatch fail")
        }
        if x != 2 {
            t.Error("AnyMatch fail,number of executions not 2", x)
        }
    }
    {
        x := 0
        if !FromIntSeq().Skip(1).Take(3).NonMatch(func(i int) bool {
            x++
            return i == 0
        }) {
            t.Error("NonMatch fail")
        }
        if x != 3 {
            t.Error("NonMatch fail,number of executions not 3", x)
        }
    }
    {
        x := 0
        if FromIntSeq().Skip(1).Take(3).NonMatch(func(i int) bool {
            x++
            return i == 2
        }) {
            t.Error("NonMatch fail")
        }
        if x != 2 {
            t.Error("NonMatch fail,number of executions not 2", x)
        }
    }
}
