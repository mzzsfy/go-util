package helper

import "testing"

func TestDfa(t *testing.T) {
    for i := 1; i < 10; i++ {
        dfa := NewDfa(MakeNewDfsNode[bool](i))
        values := []string{"hello", "world", "word", "work", "wordd", "worddd", "你好", "你好!", "世界", "世界123123", "世1界", "▲■※■▲△▲※", "︿■○●▲※", "👌😂■😶‍🌫️🎞️", "👌△😂😶‍🌫️🎞️"}
        for _, value := range values {
            dfa.Add([]byte(value), true)
        }
        for _, value := range values {
            test := dfa.Test([]byte(value))
            if test == nil {
                t.Errorf("Test failed,the resulting value is nil, value: %s", value)
            } else if !test.Value() {
                t.Errorf("Test failed")
            }
        }
        {
            test := dfa.Test([]byte("wor"))
            if test != nil {
                t.Errorf("Test failed")
            }
            test = dfa.Test([]byte("worddl"))
            if test != nil {
                t.Errorf("Test failed")
            }
            test = dfa.Test([]byte("👌○😂😶‍🌫️🎞️"))
            if test != nil {
                t.Errorf("Test failed")
            }
            test = dfa.Test([]byte("👌"))
            if test != nil {
                t.Errorf("Test failed")
            }
        }
        for _, value := range values {
            test := dfa.Test([]byte(value))
            if test == nil {
                t.Errorf("Test failed,the resulting value is nil, value: %s", value)
            } else if !test.Value() {
                t.Errorf("Test failed")
            }
        }
    }
}

func BenchmarkDfa_Add(b *testing.B) {
    for j := 1; j < 5; j++ {
        j := j
        b.Run("Add"+NumberToString(j), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                dfa := NewDfa(MakeNewDfsNode[struct{}](j))
                values := []string{"hello", "world", "word", "work", "wordd", "worddd"}
                for _, value := range values {
                    dfa.AddSimple(append([]byte(value), NumberToString(i)...))
                }
            }
        })
    }
}

func BenchmarkDfa_Test(b *testing.B) {
    for j := 1; j < 5; j++ {
        j := j
        b.Run("Test"+NumberToString(j), func(t *testing.B) {
            dfa := NewDfa(MakeNewDfsNode[struct{}](j))
            values := []string{"hello", "world", "word", "work", "wordd", "worddd"}
            for _, value := range values {
                dfa.AddSimple([]byte(value))
            }
            for i := 0; i < 100; i++ {
                for _, value := range values {
                    dfa.AddSimple(append([]byte(value), NumberToString(i)...))
                }
            }
            for i := 0; i < t.N; i++ {
                for _, value := range values {
                    test := dfa.Test(StringToBytes(value))
                    if test == nil {
                        t.Errorf("Test failed,the resulting value is nil, value: %s", value)
                    }
                }
            }
        })
    }
}

func BenchmarkDfa_Test1(b *testing.B) {
    for j := 1; j < 5; j++ {
        j := j
        b.Run("Test1_"+NumberToString(j), func(t *testing.B) {
            dfa := NewDfa(MakeNewDfsNode[struct{}](j))
            values := []string{"hello", "world", "word", "work", "wordd", "worddd"}
            for _, value := range values {
                dfa.AddSimple([]byte(value))
            }
            for i := 0; i < 100; i++ {
                for _, value := range values {
                    dfa.AddSimple(append([]byte(value), NumberToString(i)...))
                }
            }
            t.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    for _, value := range values {
                        test := dfa.Test(StringToBytes(value))
                        if test == nil {
                            t.Errorf("Test failed,the resulting value is nil, value: %s", value)
                        }
                    }
                }
            })
        })
    }
}
