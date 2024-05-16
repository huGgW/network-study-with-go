package ch05

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

// UDP의 Client에서 net.Conn 인터페이스 사용시
// 따로 송신자, 수신자의 주소를 매번 확인할 필요 없이 깔끔하게 코드 작성 가능.
// 허나 TCP의 기능은 없기 때문에 
// Write 메서드에서 목적지에서 패킷을 수신하지 못해도 에러를 반환하지 않는등 
// 여전히 애플리케이션상에서 처리해야 됨.
func TestDialUDP(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }
    defer cancel()


    // 기존 `net.Dial`을 이용하여 스트림 지향적으로 udp 연결을 수립
    // Client에서만 이를 사용 가능, Server는 반드시 net.PacketConn을 사용해야 함.
    client, err := net.Dial("udp", serverAddr.String())
    if err != nil {
        t.Fatal(err)
    }
    defer func() { client.Close() }()


    interloper, err := net.ListenPacket("udp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    interrupt := []byte("pardon me")
    n, err := interloper.WriteTo(interrupt, client.LocalAddr())
    if err != nil {
        t.Fatal(err)
    }
    _ = interloper.Close()

    if l := len(interrupt); l != n {
        t.Fatalf("wrote %d bytes of %d", n, l)
    }


    ping := []byte("ping")
    // net.Conn은 주소를 제공할 필요 없이 Dial에서 연결한 주소로 메시지 전송
    _, err = client.Write(ping)
    if err != nil {
        t.Fatal(err)
    }

    buf := make([]byte, 1024)
    // net.Conn은 항상 Dial에서 연결한 주소에서 메시지를 읽음
    n, err = client.Read(buf)
    if err != nil {
        t.Fatal(err)
    }

    if !bytes.Equal(ping, buf[:n]) {
        t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
    }

    err = client.SetDeadline(time.Now().Add(time.Second))
    if err != nil {
        t.Fatal(err)
    }

    // 패킷을 또 읽으려고 해도 현재 interloper에서 온 패킷만 존재하고
    // Dial을 통해 연결한 echo server에서 온 패킷은 존재하지 않으므로
    // Read가 일어나지 않음!!
    _, err = client.Read(buf)
    if err == nil {
        t.Fatal("unexpected packet")
    }
}
