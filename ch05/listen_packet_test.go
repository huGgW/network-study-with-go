package ch05

import (
	"bytes"
	"context"
	"net"
	"testing"
)

// TCP는 연결 객체와 리스너를 구분하는 반면,
// UDP는 세션 기능이 없어 둘이 구분되어 있지 않음.
// 즉, 패킷을 수신한 후 애플리케이션 상에서 처리해야 될 작업이 많음
//	(송신자 주소 검증 등)
func TestListenPacketUDP(t *testing.T) {
    // echo server
	ctx, cancel := context.WithCancel(context.Background())
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()


    // client
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// defer를 익명함수로 wrap한 이유는
	// client가 변경될 경우를 tracking하기 위해서다.
	// 익명함수를 감싸지 않은 경우,
	// 해당 line을 실행 중일 때의 client의 주소의 method를 종료 시 호출하도록 한다.
	// 반면, 익명함수로 감싼 경우,
	// 함수 종료 시, 익명 함수 안의 내용을 실행하게 되므로
	// 종료 시점에서의 client 주소의 method를 호출하게 된다.
	defer func() { _ = client.Close() }()


	// Interupt하기 위한 서버
	// echo server가 응답하기 전 client를 가로체기 위한 서버
	interlooper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

    // "pardon me"를 먼저 client로 보내 interupt
	interrupt := []byte("pardon me")
	n, err := interlooper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	_ = interlooper.Close()

	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}


    // client -> server로 "ping"을 보냄
	ping := []byte("ping")
	_, err = client.WriteTo(ping, serverAddr)
	if err != nil {
		t.Fatal(err)
	}


    // client에서 패킷을 읽음.
    // 이 때, interupt에서 먼저 보낸 "pardon me" 패킷이 읽혀야 됨.
	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(interrupt, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	if addr.String() != interlooper.LocalAddr().String() {
		t.Errorf(
			"expected message from %q; actual sender is %q",
			interlooper.LocalAddr(), addr,
		)
	}


    // 다시 client에서 패킷을 읽을 때, 이제서야 "ping"이 echo server로부터 옴.
	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}

	if addr.String() != serverAddr.String() {
		t.Errorf(
			"expected message from %q; actual sender is %q",
			serverAddr, addr,
		)
	}
}
