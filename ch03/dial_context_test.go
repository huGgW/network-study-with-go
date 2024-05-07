package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

/*
Context: Goroutine간에 데이터를 전달, 실행중인 작업을 취소할 수 있는 기능 제공

주요 컨셉
- 취소 신호: 고루틴이 작업을 중단하도록 알리는 신호
- 타임아웃: 특정 시간이 지나면 작업이 자동으로 취소 (기간 설정)
- 데드라인: 명시된 시간까지 작업이 완료되어야 하며, 그렇지 않으면 취소됨 (시각 설정)
- 값 전달: context를 통해 고루틴 사이에 값(메타데이터) 전달 가능

핵심 함수와 타입
Context 인터페이스: 모든 컨텍스트 타입의 기본이 되는 인터페이스입니다.
context.Background(): 모든 컨텍스트의 루트로, 주로 메인 함수나 테스트에서 시작점으로 사용됩니다.
context.TODO(): 아직 어떤 컨텍스트를 사용해야 할지 결정되지 않았을 때 사용합니다.
context.WithCancel(parent Context) (ctx Context, cancel CancelFunc):
    부모 컨텍스트를 기반으로 새로운 컨텍스트를 생성하고, 취소 함수도 반환합니다.
    이 함수를 호출하면 관련된 모든 컨텍스트가 취소됩니다.
context.WithDeadline(parent Context, d time.Time) (Context, CancelFunc):
    특정 시간을 데드라인으로 설정합니다.
context.WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc):
    특정 시간만큼 지나면 컨텍스트가 취소됩니다.
*/

// Context를 사용하여 timeout 발생 테스트
func TestDialContext(t *testing.T) {
    dl := time.Now().Add(5 * time.Second) // 현재로부터 5초 뒤 deadline 설정
    ctx, cancel := context.WithDeadline(context.Background(), dl) // deadline이 설정된 context 생성
    defer cancel() // 함수 종료 시 context cancel이 바로 되도록 등록하여 garbage collect되도록

    var d net.Dialer // DialContext is method of Dialer struct
    d.Control = func(_, _ string, _ syscall.RawConn) error {
        time.Sleep(5*time.Second + time.Millisecond) // context deadline을 넘도록 대기 설정
        return nil
    }

    conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
    if err == nil {
        conn.Close()
        t.Fatal("connection did not time out")
    }

    nErr, ok := err.(net.Error)
    if !ok {
        t.Error(err)
    } else {
        if !nErr.Timeout() {
            t.Errorf("error is not a timeout: %v", err)
        }
    }

    if ctx.Err() != context.DeadlineExceeded {
        t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
    }
}
