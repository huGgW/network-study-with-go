package echo

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnixPacket(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixpacket")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	rAddr, err := streamingEchoServer(ctx, "unixpacket", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	// unixpacket 타입 네트워크는 세션 지향형이므로 net.Dial을 이용하여 연결 초기화
	conn, err := net.Dial("unixpacket", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { conn.Close() }()

    // "ping" 메시지 3번 쓰기
	msg := []byte("ping")
	for range 3 {
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	// unixpacket은 데이터그램 기반 네트워크 타입과 유사하게 한 번 읽기에 하나의 메시지 반환
	buf := make([]byte, 1024)
	for range 3 {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(msg, buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q",
				msg, buf[:n])
		}
	}

    // "ping" 메시지 3번 더 쓰기
	for range 3 {
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf = make([]byte, 2) // 각 응답의 첫 2바이트만 읽도록 
	for range 3 {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		// msg의 앞의 2바이트만 실제로 읽히고 나머지는 매번 버려짐을 확인
		if !bytes.Equal(msg[:2], buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q",
				msg[:2], buf[:n])
		}
	}
}
