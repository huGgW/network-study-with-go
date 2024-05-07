package ch03

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanout(t *testing.T) {
	// ctx.Err() 에는 context.Canceled, context.DeadlineExceeded, nil 3개의 값이 가능
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// listener는 하나의 연결을 수락 후, 연결을 종료
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	// dialer를 새로 생성하는 함수
	dial := func(
		ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup,
	) {
		defer wg.Done()

		var d net.Dialer
		// 주어진 context에서 address로 연결 시도
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			t.Logf("dial ctx err: %v\n", err)
			return
		} else {
			t.Logf("dial ctx err is nil\n")
		}
		c.Close()

		// select: 여러 개의 channel로부터 준비된 channel 하나를 선택하여 동작
		// default가 없는 경우 blocking으로 동작

		select {
		case <-ctx.Done(): // cancel된 경우 취소
		case response <- id: // cancel되지 않은 경우 response에 id 전송
		}
	}

	// WaitGroup: 여러 고루틴이 모두 완료될때까지 대기할 수 있도록 도와주는 동기화 매커니즘
	// Add(int) -> 카운터를 주어진 수만큼 증가
	// Done() -> 카운터를 1 감소
	// Wait() -> 카운터가 0이 될 때까지 대기 (blocking)

	res := make(chan int)
	var wg sync.WaitGroup

	// 여러 dial 생성
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	// 첫 response를 받자마자 context 취소, 다른 dial들 모두 취소하도록.
	response := <-res
	time.Sleep(1 * time.Second)
	cancel()
	wg.Wait()
	close(res)

	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual: %s", ctx.Err())
	}

	t.Logf("dialer %d retrieved the resource", response)
}
