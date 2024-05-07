package ch03

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

// 하트비트: 네트워크 연결의 데드라인을 지속해서 뒤로 설정하기 위한 의도로
//     응답을 받기 이해 원격지로 봬는 메시지
// 네트워크 연결이 계속 지속되어야 하여 애플리케이션 계층에서 긴 유휴 시간을 가져야 할 경우 사용

// Goroutine에서 동작하도록 context를 매개변수로 받음
func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration
	select {
	case <-ctx.Done():
		return
	case interval = <-reset: // reset 채널에서 초기 간격 받아옴
	default:
	}

	if interval <= 0 {
		interval = defaultPingInterval
	}

	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		// select: 여러 개의 channel로부터 준비된 channel 하나를 선택하여 동작
		// default가 없는 경우 blocking으로 동작
		select {
		case <-ctx.Done(): // context가 취소 -> 함수 종료
			return
		case newInterval := <-reset: // reset 시그널 -> 새로운 interval로 timer를 리셋
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C: // timer 만료 -> ping을 보냄
			if _, err := w.Write([]byte("ping")); err != nil {
				// 여기서 연속으로 발생하는 타임아웃을 추적하고 처리.
				return
			}
		}

		_ = timer.Reset(interval)
	}
}
