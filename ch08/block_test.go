package ch08

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	select {}
}

func TestBlockIndefinitely(t *testing.T) {
	// httptest.NewServer를 통해 간단히 테스트 서버 생성
	ts := httptest.NewServer(
		http.HandlerFunc(blockIndefinitely),
	)

	// Go의 Http Client와 http.Get, http.Head, http.Post 등 helper method는
	// 기본적으로 timeout되지 않는다.
	_, _ = http.Get(ts.URL)
	t.Fatal("client did not indefinitely block")
}

func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(blockIndefinitely),
	)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5 * time.Second,
	)
	defer cancel()

    // request를 context와 함께 생성
    // context는 초기화와 함께 타이머가 동작함을 고려하여 시간 적절히 설정
    // (response body 읽는 시간까지 고려)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, ts.URL, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}

	_ = resp.Body.Close()
}

// deadline 없이 context를 생성 후 타이머를 이용해 구현 가능
/*
ctx, cancel := context.WithCancel(context.Background())
timer := time.AfterFunc(5 * time.Second, cancel)
// HTTP 요청 생성, 헤더 읽음
...
// response body 읽기 전 5초 추가
timer.Reset(5 * time.Second)
...
*/
