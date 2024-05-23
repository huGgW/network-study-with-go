package echo

import (
	"context"
	"net"
	"os"
)

// 스트림 기반의 네트워크 타입을 네트워크 문자열로 전달받아
// 여러 스트리밍 네트워크에 적용 가능
//
// 네트워크 타입 tcp인 경우, 주소는 IP주소와 포트의 조합.
// 네트워크 타입이 unix, unixpacket의 경우 주소는 존재하지 않는 파일 경로.
func streamingEchoServer(
	ctx context.Context, network string, addr string,
) (net.Addr, error) {
    // net.Listen 혹은 net.ListenUnix를 사용하는 경우, close 시 소켓 파일 제거해줌.
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			<-ctx.Done() // context를 취소하면 서버가 종료되도록
			_ = s.Close()
		}()

		for {
			conn, err := s.Accept() // 연결 수립
			if err != nil {
				return
			}

			go func() {
				defer func() { conn.Close() }()

				for {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf) // Read
					if err != nil {
						return
					}

					_, err = conn.Write(buf[:n]) // Write
					if err != nil {
						return
					}
				}
			}()
		}
	}()

	return s.Addr(), nil
}

// 데이터그램 기반 네트워크 타입을 이용한 echo server
func datagramEchoServer(
	ctx context.Context, network string, addr string,
) (net.Addr, error) {
    // net.ListenPacket은 close시 소켓 파일을 따로 제거하지 않음.
	s, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			<-ctx.Done()
			s.Close()
			if network == "unixgram" {
			    // unix domain socket인 경우 수동으로 해당 파일을 제거해주어야 함.
				os.Remove(addr)
			}
		}()

		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}

			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}
