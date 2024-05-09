package ch04

import (
	"bufio"
	"net"
	"reflect"
	"strings"
	"testing"
)

const payload = "The bigger the interface, the weaker the abstration."

func TestScanner(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != (nil) {
			t.Error(err)
			return
		}
		defer conn.Close()

		// Payload를 네트워크를 통해 전송
		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial(listener.Addr().Network(), listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Scanner  생성
	scanner := bufio.NewScanner(conn)
	// Scanner는 기본적으로 개행문자'\n' 을 기준으로 데이터를 분할 (bufio.ScanLines)
	// 여기서는 공백을 기준으로 데이터를 분할하는 bufio.ScanWords를 사용하도록 설정
	scanner.Split(bufio.ScanWords)

	var words []string

	// Scanner는  네트워크 연결에서 읽을 데이터가 있는 한 계속해서 읽은 후 True를 반환
	// 실패 혹은 완료 시 False를 반환, scanner.Err 메서드를 통해 에러 반환 가능
	// Scan 시 내부에서는 구분자를 찾을 떄 까지 여러번의 Read 메서드 호출
	for scanner.Scan() {
		// Scanner.Text -> 네트워크 연결로부터 읽은 데이터 청크를 문자열로 변환.
		// 이 때 구분자는 제외됨.
		words = append(words, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}

	t.Logf("Scanned words: %#v", words)

	expected := strings.Split(payload, " ")
	if !reflect.DeepEqual(words, expected) {
		t.Fatal("inaccurate scanned word list")
	}
}
