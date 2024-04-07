package helper

import (
    "testing"
)

func Test_TryWithStack_NoPanic(t *testing.T) {
    TryWithStack(func() {}, func(e any, stack []Stack) {
        t.Errorf("Unexpected error: %v", e)
    })
}

func Test_TryWithStack_PanicWithoutError(t *testing.T) {
    TryWithStack(func() { panic(nil) }, func(err any, stack []Stack) {
        if err != nil {
            t.Errorf("Expected nil error, got: %v", err)
        }
    })
}

func Test_TryWithStack_PanicWithError(t *testing.T) {
    expectedError := "test error"
    TryWithStack(func() { panic(expectedError) }, func(err any, stack []Stack) {
        if err != expectedError {
            t.Errorf("Expected error: %v, got: %v", expectedError, err)
        }
    })
}

func Test_TryWithStack_PanicWithStack(t *testing.T) {
    s := "Test_TryWithStack_PanicWithStack"
    t.Run("Test_TryWithStack_PanicWithStack", func(t *testing.T) {
        TryWithStack(func() {
            panic("test error")
        }, func(err any, stack []Stack) {
            //t.Log("\n", FormatStack(stack))
            name := SimpleFunctionName(stack[0].PC)
            i := len(s) + 5
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[1].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
        })
    })
    t.Run("Test_TryWithStack_PanicWithStack1", func(t *testing.T) {
        TryWithStack(func() {
            func() {
                panic("test error")
            }()
        }, func(err any, stack []Stack) {
            //t.Log("\n", FormatStack(stack))
            name := SimpleFunctionName(stack[0].PC)
            i := len(s) + 5
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[1].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[2].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
        })
    })
    t.Run("Test_TryWithStack_PanicWithStack2", func(t *testing.T) {
        TryWithStack(func() {
            func() {
                func() {
                    panic("test error")
                }()
            }()
        }, func(err any, stack []Stack) {
            //t.Log("\n", FormatStack(stack))
            name := SimpleFunctionName(stack[0].PC)
            i := len(s) + 5
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[1].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[2].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
            name = SimpleFunctionName(stack[3].PC)
            if name[:i] != s+".func" {
                t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
            }
        })
    })
}
