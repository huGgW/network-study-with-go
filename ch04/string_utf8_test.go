package ch04

import (
	"bytes"
	"testing"
)

func TestTypeAscii(t *testing.T) {
    var wst String = "Hello, World!!"

    var buffer bytes.Buffer
    wn, err := wst.WriteTo(&buffer)
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("Write bytes: %d\n", wn)

    var rst String
    rn, err := rst.ReadFrom(&buffer)
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("Read bytes: %d\n", rn)

    if rst.String() != wst.String() {
        t.Fatalf("Expected: %s, \nActual:   %s\n", wst.String(), rst.String())
    }
}

func TestTypeUTF(t *testing.T) {
    var wst String = "안녕하세요, 이것은 한글이에요."

    var buffer bytes.Buffer
    wn, err := wst.WriteTo(&buffer)
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("Write bytes: %d\n", wn)

    var rst String
    rn, err := rst.ReadFrom(&buffer)
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("Read bytes: %d\n", rn)

    if rst.String() != wst.String() {
        t.Fatalf("Expected: %s, \nActual:   %s\n", wst.String(), rst.String())
    }
}
