package ch04

import (
	"io"
	"log"
	"net"
	"os"
	"testing"
)

// log.Logger를 embedding하여 네트워크 트래픽 로깅
type Monitor struct {
	*log.Logger
}

// Implements io.Writer interface
func (m *Monitor) Write(p []byte) (int, error) {
	// return len(p), m.Output(2, string(p))

	// TeeReader, MultiWriter의 경우 writer에서 오류가 나면 전체 reader/writer도 오류 발생
	// 따라서 로깅에 의해 주요 로직이 영향받지 않도록 하려면 로깅 에러는 내부에서 처리하는 방식이 적절
	err := m.Output(2, string(p))
	if err != nil {
		log.Println(err)
	}
	return len(p), nil
}


func TestMonitor(_ *testing.T) {
	monitor := &Monitor{
		Logger: log.New(os.Stdout, "monitor: ", 0),
	}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}

	done := make(chan struct{})
	
	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		b := make([]byte, 1024)
		// io.TeeReader: r로부터 데이터 읽음 -> w를 이용해 write -> buffer에 읽음
		// 해당 과정 (특히 write이 먼저 발생하는 과정)에서 blocking이 발생함을 유의
		// write 과정 중 에러가 발생해도 TeeReader를 통해 에러 값이 반환됨을 유의
		r := io.TeeReader(conn, monitor)
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

		// io.MultiWriter: write시 주어진 writer들에게 순차적으로 write함.
		// 중간에 에러가 발생하면 즉시 중단되고 반환됨을 유의
		w := io.MultiWriter(conn, monitor)

		_, err = w.Write(b[:n])
		if err != nil && err != io.EOF {
			monitor.Println(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}

	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}

	conn.Close()
	<-done
}