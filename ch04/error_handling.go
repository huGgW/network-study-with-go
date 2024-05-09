package ch04

import (
	"errors"
	"log"
	"net"
	"time"
)

func sendHelloWorldRetry() error {
	var (
		err error
		n int
		i = 7
	)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:", time.Second * 10)
	if err != nil {
		return err
	}
	defer conn.Close()

	// For loop을 통해 일시적 에러에 대한 retry를 시도
	for ; i > 0; i-- {
		n, err = conn.Write([]byte("hello world"))
		if err != nil {
			// error가 net.Error로 assertion되고 Timeout일 경우 잠시 대기 후 재시도
			// net.Error.Temporary()는 deprecated됨
			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				log.Println("timeout error (timeout):", nErr)
				time.Sleep(10 * time.Second)
				continue
			}
			return err
		}
		break
	}

	if i == 0 {
		return errors.New("temporary write failure threshold exceeded")
	}

	log.Printf("wrote %d bytes to %s\n", n, conn.RemoteAddr())
	return nil
}

