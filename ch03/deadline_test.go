package ch03

import (
	"net"
	"testing"
	"time"
)

func TestDeadline(t *testing.T) {
	sync := make(chan struct{})

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

		defer func() {
			conn.Close()
			// early return으로 인해 sync channel이 blocking을 유발하지 않도록 close
			close(sync)
		}()

		// deadline: 패킷 주고받음 없이 네트워크 연결이 유휴 상태로 지속할 수 있는 시간
		// 기본적으로 Go에서 deadline은 무한이다.
		// SetReadDeadline, SetWriteDeadline을 이용하여 read, write에 대한 deadline을,
		// SetDeadline을 통하여 동시에 deadline을 설정 가능.
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1)
		// deadline을 지나면 timeout 에러 반환
		_, err = conn.Read(buf) // 원격 node로부터 데이터를 받을 때 까지 blocking
		nErr, ok := err.(net.Error)
		if !ok || !nErr.Timeout() {
			t.Errorf("expected timeout error; actual: %v", err)
		}

		sync <- struct{}{}

        // deadline을 뒤로 연장
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

    // 첫 read가 데드라인을 넘어갈 때까지 대기하여 데드라인 발생
	<-sync
	// 이후 write을 하여 두번쨰 read는 정상적으로 데이터를 받음
	_, err = conn.Write([]byte("1"))
}
