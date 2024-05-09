package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

func TestPayloads(t *testing.T) {
    b1 := Binary("Clear is better than clever.")
    b2 := Binary("Don't panic.")
    s1 := String("Errors are value.") 
    payloads := []Payload{&b1, &s1, &b2}

    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    go func() {
        conn, err := listener.Accept()
        if err != nil {
            t.Error(err)
            return
        }
        defer conn.Close()

        for _, p := range payloads {
            // paylaod를 conn을 통해 write
            _, err = p.WriteTo(conn)
            if err != nil {
                t.Error(err)
                break
            }
        }
    }()

    conn, err := net.Dial(
        listener.Addr().Network(), listener.Addr().String(),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()

    for _, expected := range payloads {
        // conn을 통해 읽어서 Payload로 decode
        actual, err := decode(conn)
        if err != nil {
            t.Fatal(err)
        }

        if !reflect.DeepEqual(expected, actual) {
            t.Errorf("value mismatch: %v != %v", expected, actual)
            continue
        }

        t.Logf("[%T] %[1]q", actual)
    }
}

func TestMaxPayloadSize(t *testing.T) {
    var buf bytes.Buffer
    err := buf.WriteByte(BinaryType)
    if err != nil {
        t.Fatal(err)
    }

    // 강제로 payload 최대 명시 사이즈를 넘기도록 테스트
    // 1GB size임을 4 byte 데이터를 통해 write
    err = binary.Write(&buf, binary.BigEndian, uint32(1<<30))
    if err != nil {
        t.Fatal(err)
    }

    var b Binary
    _, err = b.ReadFrom(&buf)
    if err != ErrMaxPayloadSize {
        t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
    }
}