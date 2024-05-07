package ch03

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    done := make(chan struct{})
    go func() {
        defer func() { done <- struct{}{} }()

        for {
            // Accept the connection
            // 리스너가 수신 연결 감지하고 TCP Handshake가 완료될때까지 Block됨.
            conn, err := listener.Accept()
            // error on connection such as failing TCP handshake or closed listener
            if err != nil {
                t.Log(err)
                return
            }

            // conn은 net.Conn interface로, 이 경우는 net.TCPConn 객체의 pointer

            // Handler goroutine
            // Asynchronously handling buisness logic using connection
            // This prevents buisness logic blocks the next connection
            go func(c net.Conn) {
                defer func() {
                    // Send the FIN packet to server before exit the function
                    c.Close()
                    done <- struct{}{}
                }()

                // Buisness Logic using TCP Connection
                buf := make([]byte, 1024)
                for {
                    // 4 way handshake 과정에 의해 FIN 패킷을 받으면 err로 io.EOF 반환
                    n, err := c.Read(buf)
                    if err != nil {
                        if err != io.EOF {
                            t.Error(err)
                        }
                        return // FIN으로 인해 io.EOF 발생 시 handler 종료
                    }
                    t.Logf("received: %q", buf[:n])
                }
            }(conn)
        }
    }()

    // net.Dial: network 종류와 ip addr + port number를 받음
    // 주소로는 ipv4, ipv6, http 등 가능, 여러 개의 주소 가능.
    // 주어진 곳으로 연결을 시도, 연결 후 net.Conn 객체와 error interface 반환
    conn, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }

    // graceful termination 시작.
    conn.Close()
    <-done
    listener.Close()
    <-done
}
