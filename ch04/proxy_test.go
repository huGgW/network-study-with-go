package ch04

import (
	"io"
	"net"
	"sync"
	"testing"
)

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup

	// ping -> pong 반환, 그 외에 받은대로 반환하는 echo 서버
	server, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()

				for {
					buf := make([]byte, 1024)
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}

						return
					}

					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = c.Write([]byte("pong"))
					default:
						_, err = c.Write(buf[:n])
					}

					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}

						return
					}
				}
			}(conn)
		}
	}()

	// Server <--> Client를 연결해주는 proxy server
	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			// Proxy <- Client 연결 수락
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}

			go func(from net.Conn) {
				defer from.Close()

				// Proxy -> Server로 연결 시도
				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}
				defer to.Close()

				// Server <--> Client proxy
				err = proxy(from, to)
				if err != nil && err != io.EOF {
					t.Error(err)
				}
			}(conn)
		}
	}()

	
	// Client를 생성하여 실제 테스트
	conn, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	msgs := []struct{ Message, Reply string }{
		{"ping", "pong"},
		{"ping", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	for i, m := range msgs {
		_, err := conn.Write([]byte(m.Message))
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		actual := string(buf[:n])
		t.Logf("%q -> proxy -> %q", m.Message, actual)
		if actual != m.Reply {
			t.Errorf(
				"%d: expected reply: %q; actual: %q",
				i, m.Reply, actual,
			)
		}
	}

	// 정리
	conn.Close()
	proxyServer.Close()
	server.Close()
	wg.Wait()
}
