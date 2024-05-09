package ch04

import (
	"net"
	"time"
)

// Not actually executable method. Just examples

func ZeroWindowErr() error {
	conn, _ := net.Dial("tcp", "127.0.0.1:8080")

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return err
		}
		handle(buf[:n]) // 긴 시간 동안 blocking됨
		// handle로 인해 긴 시간 동안 blocking되고, 수신 버퍼는 계속 채워지게 됨.
		// 결과적으로 수신 버퍼가 가득 차 더 이상 데이터를 받지 못하는 zero window 상태가 됨.
		// 이 경우 송신자에서 데이터 흐름을 throttle을 하여 해소 시도 가능
	}
}
func handle(buf []byte) {
	// do something very long
	time.Sleep(time.Second * 10)
}


func CloseWaitStuckErr() error {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			// defer c.Close() // 이부분을 잊지 말고 반드시 connection을 close하도록!

			buf := make([]byte, 1024)

			for {
				n, err := c.Read(buf)
				// 해당 err가 발생시 defer를 등록하지 않으면 connection close 없이 고루틴 종료
				// TCP 소켓은 CLOSE_WAIT 상태에 머물러 있게 됨
				if err != nil {
					return
				}

				handle(buf[:n])
			}
		}(conn)
	}
}
