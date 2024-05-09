package ch04

import (
	"io"
	"net"
	"crypto/rand"
	"testing"
)

// 고정된 버퍼에 데이터 읽기
func TestReadIntoBuffer(t *testing.T) {
    payload := make([]byte, 1<<24) // 16mb 고정된 버퍼 생성
    _, err := rand.Read(payload) // create random payload
    if err != nil {
        t.Fatal(err)
    }

    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    go func() {
        conn, err := listener.Accept()
        if err != nil {
            t.Log(err)
            return
        }
        defer conn.Close()

        _, err = conn.Write(payload) // payload를 네트워크 연결로 씀 (전송)
        if err != nil {
            t.Error(err)
        }
    }()

    conn, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()

    buf := make([]byte, 1<<19) // 512kb 고정된 버퍼
    totalReadByte := 0
    // 512kb가 최대 buffer 크기이므로 여러 번에 나누어 읽어야 함.
    for {
        n, err := conn.Read(buf) // buffer 크기만큼 읽고, 읽은 byte와 err를 반환.
        if err != nil {
            if err != io.EOF {
                t.Error(err)
            }
            break
        }

        t.Logf("read %d bytes", n) // buf[:n]은 conn 객체에서 읽은 데이터
        totalReadByte += n
    }

    if totalReadByte != 1<<24 {
        t.Errorf("Total Read Byte: %d not equal to Send Byte %d", totalReadByte, 1<<24)
    }
}
