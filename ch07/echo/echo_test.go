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

func TestEchoServerUnix(t *testing.T) {
	// 임시 directory 생성
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { // 종료 시 임시 directory와 하위 파일들 제거
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join( // 해당 process에 대한 socket 주소 (파일명) 생성
		dir,
		fmt.Sprintf("%d.sock", os.Getpid()),
	)
	rAddr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { cancel() }()

	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	// 생성한 socket streaming echo server로 연결
	conn, err := net.Dial("unix", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { conn.Close() }()

	msg := []byte("ping")
	for range 3 {
		_, err = conn.Write(msg) // "ping" 메세지 3번 쓰기
		if err != nil {
			t.Fatal(err)
		}
	}

	// echo로부터 읽어오기
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	// 3번 "ping"이 연결된 것이 결과로 읽혔어야 함.
	expected := bytes.Repeat(msg, 3)
	if !bytes.Equal(expected, buf[:n]) {
		t.Fatalf(
			"expected reply %q; actual reply %q",
			expected, buf[:n],
		)
	}
}
