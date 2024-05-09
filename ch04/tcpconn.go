package ch04

import (
	"errors"
	"net"
	"time"
)

// Not actually executable method. Just examples

// net.TCPConn의 경우 일부 메서드가 운영체제에 따라 사용할 수 없거나 hard limit이 존재할 수 있음.
// 따라서 반드시 필요한 경우에만 해당 객체를 이용하여야 한다.
func ExampleTCPConn() error {
	// Simple case: type assertion
	ls, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return err
	}
	netConn, err := ls.Accept()
	if err != nil {
		return err
	}
	assertTcpConn, ok := netConn.(*net.TCPConn)
	if !ok {
		return errors.New("assertion fail")
	}
	assertTcpConn.Close()
	ls.Close()


	// TCP Listen
	localhost, err := net.ResolveTCPAddr("tcp", "127.0.0.1:")
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", localhost)
	if err != nil {
		return err
	}

	// TCP Accept
	serverTcpConn, err := listener.AcceptTCP()
	if err != nil {
		return err
	}

	serverTcpConn.Close()
	listener.Close()


	// TCP Dial
	addr, err := net.ResolveTCPAddr("tcp", "www.google.com:http")
	if err != nil {
		return err
	}

	dialTcpConn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	// KEEP ALIVE
	// 수신자로부터 메세지가 정상적으로 도달하였음을 확인하도록 요청하여 네트워크 연결의 무결성을 확인하기 위한 메세지
	// keepalive 메세지가 특정 수 이상 승인되지 않으면 os는 네트워크 연결 종료
	// heartbeet로 데드라인을 앞당기면 선제적으로 네트워크 상에서 발생할 수 있는 문제 탐지 가능

	err = dialTcpConn.SetKeepAlive(true) // keepalive를 사용할 것인지 설정
	err = dialTcpConn.SetKeepAlivePeriod(time.Minute) // keepalive 메세지 보내는 주기 설정


	// Linger
	// 데이터를 net.Conn 객체에 썼으나 아직 전송되지 못했거나 ACK를 받지 못한 상태에서 네트워크 연결이 끊긴 경우,
	// 기본적으로 OS는 background에서 데이터 전송 마무리
	// 이러한 기본동작을 수정하기 위해 SetLinger method 사용

	// If sec \< 0 (the default), the operating system finishes sending the data in the background.
	// If sec == 0, the operating system discards any unsent or unacknowledged data.
	//    RST 패킷을 보내어 즉시 연결 중단, 일반적 종료 절차를 무시.
	// If sec > 0, the data is sent in the background as with sec \< 0. On some operating systems including Linux, this may cause Close to block until all data has been sent or discarded. On some operating systems after sec seconds have elapsed any remaining unsent data may be discarded.
	dialTcpConn.SetLinger(-1)


	// 네트워크 읽기/쓰기 버퍼 크기 변경
	err = dialTcpConn.SetReadBuffer(212992)
	err = dialTcpConn.SetWriteBuffer(212992)



	dialTcpConn.Close()
	return nil
}
