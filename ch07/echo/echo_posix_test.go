//go:build darwin || linux

// 특정 platform에서만 build하도록 build constraint 설정

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

func TestEchoServerUnixDatagram(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixgram")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	// server socket 파일 생성
	ctx, cancel := context.WithCancel(context.Background())
	sSocket := filepath.Join(
		dir, fmt.Sprintf("s%d.sock", os.Getpid()),
	)
	// echo server 생성
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	err = os.Chmod(sSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	// Client Socket 파일 생성
	cSocket := filepath.Join(
		dir, fmt.Sprintf("c%d.sock", os.Getpid()),
	)
	// Client 연결 생성
	client, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.Close()
	}()

	err = os.Chmod(cSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	// "ping" 3번 쓰기 (보내기)
	msg := []byte("ping")
	for range 3 {
		_, err = client.WriteTo(msg, serverAddr)
		if err != nil {
			t.Fatal(err)
		}
	}

	// "ping" 3번 읽기 (받기)
	buf := make([]byte, 1024)
	for range 3 {
		n, addr, err := client.ReadFrom(buf) // packet 지향은 각 메시지를 구분하여 저장한다!
		if err != nil {
			t.Fatal(err)
		}

		if addr.String() != serverAddr.String() {
			t.Fatalf("received reply from %q instead of %q",
				addr, serverAddr)
		}
		if !bytes.Equal(msg, buf[:n]) {
			t.Fatalf("expected reply %q; actual reply %q",
				msg, buf[:n])
		}
	}
}
