package ch03

import (
	"net"
	"syscall"
	"testing"
	"time"
)

// net.DialTimeout 함수는 net.Dialer interface에 대한 제어권을 제공하지 않아
// test를 위한 별도 구현체를 구현한 DialTimeout 생성.
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
    d := net.Dialer {
        // Error를 반환하기 위한 Control 함수를 재정의.
        Control: func(_, addr string, _ syscall.RawConn) error {
            // DNS Timeout 에러 mocking
            return &net.DNSError {
                Err: "connection timed out",
                Name: addr,
                Server: "17.0.0.1",
                IsTimeout: true,
                IsTemporary: true,
            }
        },
        Timeout: timeout,
    }
    return d.Dial(network, address)
}

// routing 불가한 주소를 통해 timeout 발생 테스트
func TestDialTimeout(t *testing.T) {
    // DialTimeout 함수는 추가적으로 Timeout duration 매개변수를 통해 timeout 설정
    // 여러 개의 주소가 주어졌을 시, 모든 연결 시도가 실패, 타임아웃 시에만 net.DialTimeout 에러 반환
    //      첫 번쨰 이루어지는 연결만 유지되고 나머지 취소
    c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
    if err == nil {
        c.Close()
        t.Fatal("connection did not timeout")
    }

    // error -> net.Error type으로 assertion
    nErr, ok := err.(net.Error)
    if !ok {
        t.Fatal(err)
    }

    if !nErr.Timeout() {
        t.Fatal("error is not a timeout")
    }
}
