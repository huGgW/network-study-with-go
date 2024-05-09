package ch04

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	count = flag.Int("c", 3, "number of pings: <= 0 means forever")
	interval = flag.Duration("i", time.Second, "interval between pings")
	timeout = flag.Duration("W", 5 * time.Second, "time to wait for a reply")
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] host:port\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

// ICMP가 필터링된 상황에서 TCP의 Handshake 요청을 사용하여 원격 시스템의 상태 확인
// 원격 시스템의 포트를 소진, TCP는 ICMP에 비해 오버헤드가 크므로 유의 필요
func Ping() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Print("host:port is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	target := flag.Arg(0)
	fmt.Println("PING", target)

	if *count <= 0 {
		fmt.Println("CTRL+C to stop.")
	}

	msg := 0

	for (*count <= 0) || (msg < *count) {
		msg++
		fmt.Print(msg, " ")

		start := time.Now()
		// TCP 연결 시도, 호스트 응답 없음을 대비하여 timeout 설정
		c, err := net.DialTimeout("tcp", target, *timeout)
		dur := time.Since(start) // Handshake가 끝나는데 걸리는 시간 측정

		if err != nil {
			fmt.Printf("fail in %s: %v\n", dur, err)
			// timeout일 경우 재시도, 아니면 프로그램 종료
			// TCP 재시작 후 상태를 모니터링 하는 경우 유용
			if nErr, ok := err.(net.Error); !ok || !nErr.Timeout() {
				os.Exit(1)
			}
		} else {
			_ = c.Close()
			fmt.Println(dur)
		}

		time.Sleep(*interval)
	}
}