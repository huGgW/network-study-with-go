package ch04

import (
	"io"
	"net"
)

func proxyConn(source, destination string) error {
	// go 1.11 이후 io.Copy, io.CopyN에서 net.Conn 대신 net.TCPConn 사용시
	// user space를 경유하지 않아 더 빠른 데이터 전송 가능

	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	defer connSource.Close()

	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	defer connDestination.Close()

	// io.Copy(dst io.Writer, src io.Reader)
	// reader로부터 데이터를 읽어와 writer로 데이터를 써주는 함수

	// source -> destination
	go func() { // 노드 중 하나라도 연결이 끊기면 io.Copy는 종료되어 memory leak 없음
		io.Copy(connDestination, connSource)
	}()

	// destination -> source
	_, err = io.Copy(connSource, connDestination)

	return err
}

// io.Reader, io.Writer interface를 매개변수로 받아 다양한 종류의 io에 적용 가능
func proxy(from io.Reader, to io.Writer) error {
	// from, to가 writer, reader 인터페이스도 구현하였는지 확인 (역방향 copy를 위해)
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)

	if fromIsWriter && toIsReader {
		go func() { io.Copy(fromWriter, toReader) }()
	}

	_, err := io.Copy(to, from)

	return err
}