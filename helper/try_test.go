package helper

import (
    "testing"
)

func Test_TryWithStack_NoPanic(t *testing.T) {
    TryWithStack(func() {}, func(err Err) {
        t.Errorf("Unexpected error: %v", err)
    })
}

func Test_TryWithStack_PanicWithoutError(t *testing.T) {
    TryWithStack(func() { panic(nil) }, func(err Err) {
        if err.Error != nil {
            t.Errorf("Expected nil error, got: %v", err.Error)
        }
    })
}

func Test_TryWithStack_PanicWithError(t *testing.T) {
    expectedError := "test error"
    TryWithStack(func() { panic(expectedError) }, func(err Err) {
        if err.Error != expectedError {
            t.Errorf("Expected error: %v, got: %v", expectedError, err.Error)
        }
    })
}

func Test_TryWithStack_PanicWithStack(t *testing.T) {
    TryWithStack(func() {
        panic("test error")
    }, func(err Err) {
        t.Log("\n", FormatStack(err.Stack))
        s := "Test_TryWithStack_PanicWithStack"
        name := SimpleFunctionName(err.Stack[0].PC)
        i := len(s) + 5
        if name[:i] != s+".func" {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
        name = SimpleFunctionName(err.Stack[1].PC)
        if name != s {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
    })
}

func Test_TryWithStack_PanicWithStack1(t *testing.T) {
    TryWithStack(func() {
        func() {
            panic("test error")
        }()
    }, func(err Err) {
        t.Log("\n", FormatStack(err.Stack))
        s := "Test_TryWithStack_PanicWithStack1"
        name := SimpleFunctionName(err.Stack[0].PC)
        if name[:len(s)+5] != s+".func" {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
        name = SimpleFunctionName(err.Stack[1].PC)
        if name[:len(s)+5] != s+".func" {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
        name = SimpleFunctionName(err.Stack[2].PC)
        if name != s {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
    })
}

func Test_TryWithStack_PanicWithStack2(t *testing.T) {
    TryWithStack(func() {
        func() {
            func() {
                panic("test error")
            }()
        }()
    }, func(err Err) {
        t.Log("\n", FormatStack(err.Stack))
        s := "Test_TryWithStack_PanicWithStack2"
        name := SimpleFunctionName(err.Stack[0].PC)
        if name[:len(s)+5] != s+".func" {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
        name = SimpleFunctionName(err.Stack[1].PC)
        if name[:len(s)+5] != s+".func" {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
        name = SimpleFunctionName(err.Stack[2].PC)
        if name[:len(s)+5] != s+".func" {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
        name = SimpleFunctionName(err.Stack[3].PC)
        if name != s {
            t.Errorf("Expected first function name to be Test_TryWithStack_PanicWithStackLog, got: %v", name)
        }
    })
}
