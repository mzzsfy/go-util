package helper

import "testing"

func Test_StringToBytes(t *testing.T) {
    bs := StringToBytes("test")
    if string(bs) != "test" {
        t.Errorf("expected %v, got %v", "test", string(bs))
    }
}

func Test_BytesToString(t *testing.T) {
    bytes := []byte("test")
    s := BytesToString(bytes)
    if s != "test" {
        t.Errorf("expected %v, got %v", "test", s)
    }
}
