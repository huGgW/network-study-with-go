package ch05

import (
	"bytes"
	"context"
	"net"
	"testing"
)

// Echo 서버 테스트
// UDP는 신뢰성이 없어 
// 패킷이 손실되는 경우, (버퍼가 가득 차거나 램이 부족하면)
// 패킷 순서가 뒤바뀌는 경우 (multi thread를 이용한 전송)
// 등으로 인해 테스트는 실패할 수 있음
func TestEchoServerUDP(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    serverAddr, err := echoServerUDP(ctx, "127.0.0.1:") // echo udp server 실행
    if err != nil {
        t.Fatal(err)
    }
    defer cancel() // 테스트 종료 시 context Cancel을 통해 해당 자원 회수

    client, err := net.ListenPacket("udp", "127.0.0.1:") // client udp 생성
    if err != nil {
        t.Fatal(err)
    }
    defer client.Close()

    msg := []byte("ping")
    _, err = client.WriteTo(msg, serverAddr)
    if err != nil {
        t.Fatal(err)
    }

    buf := make([]byte, 1024)
    n, addr, err := client.ReadFrom(buf)
    if err != nil {
        t.Fatal(err)
    }

    if addr.String() != serverAddr.String() {
        t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
    }

    if !bytes.Equal(msg, buf[:n]) {
        t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
    }
}
