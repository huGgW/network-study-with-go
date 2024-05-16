package ch05

import (
	"context"
	"fmt"
	"net"
)

// 송신자가 받은 udp 패킷을 그대로 echoing해주는 서버
func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	s, err := net.ListenPacket("udp", addr) // UDP 연결 생성, (net.PacketConn, error) 반환
	if err != nil {
		return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
	}

	go func() {
		go func() {
			<-ctx.Done() // Done 신호를 받을 때까지 blocking
			_ = s.Close() // Done 신호 받은 후 해당 line 통해 server close
		}()

		buf := make([]byte, 1024)

		// 매 연결마다 새로운 연결 객체를 생성할 필요 없음.
		for {
		    // UDP는 HandShake 과정이 없기에 Accept 과정이 없음.
		    // 입력받는 모든 메시지를 읽고, 세션 수립 등이 없어 
		    //     패킷 안의 주소에 의존하여 노드를 구분할 수 있음.
			n, clientAddr, err := s.ReadFrom(buf) // server <- client
			if err != nil {
				return
			}

			// UDP는 세션이 없기 때문에 매개변수를 통해 보낼 노드 주소를 특정해야됨
			_, err = s.WriteTo(buf[:n], clientAddr) // server -> client
			if err != nil {
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}
