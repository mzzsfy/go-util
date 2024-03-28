package script

import "testing"

func Test1(t *testing.T) {
    t.Run("1+1", func(t *testing.T) {
        if e, err := Parse("1+1"); err != nil {
            t.Fatal(err)
        } else {
            r := e.Execute(nil)
            if r.IsErr() {
                t.Fatal(r.Any())
            }
            if r.Int() != 2 {
                t.Fatal(r.Any())
            }
        }
    })
}
