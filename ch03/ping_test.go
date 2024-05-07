package ch03

import (
	"context"
	"io"
	"net"
	"time"
	"testing"
)

func TestPingerAdvanceDeadline(t *testing.T) {
	done := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	begin := time.Now()
	go func() {
		defer func() { close(done) }()

		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			conn.Close()
		}()

		// Timer를 1초로 설정, Pinger를 시작.
		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer)

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])

			// 응답을 받은 경우, Pinger 초기화, Deadline 연장
			resetTimer <- 0
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

    buf := make([]byte, 1024)
    // 4번 Ping 받은 후 Pong으로 응답 -> 4초
    for i := 0; i < 4; i++ {
        n, err := conn.Read(buf)
        if err != nil {
            t.Fatal(err)
        }
        t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])

        _, err = conn.Write([]byte("PONG!!!")) // Pong을 받은 서버는 timer를 리셋!
        if err != nil {
            t.Fatal(err)
        }
    }

	// 4번의 Ping을 받기만 함 -> 4초 
    for i := 0; i < 4; i++ {
        n, err := conn.Read(buf)
        if err != nil {
            if err != io.EOF {
                t.Fatal(err)
            }
            break
        }
        t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
    }

    // Deadline인 5초 째에 서버가 닫히며 done에 신호를 보냄 -> 1초
    <-done
    end := time.Since(begin).Truncate(time.Second)
    t.Logf("[%s] done", end)
    if end != 9 * time.Second { // 총 9초 동안 동작해야 함.
        t.Fatalf("expected EOF at 9 seconds; actual: %s", end)
    }
}
