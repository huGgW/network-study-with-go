package ch08

import (
	"net/http"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {
    // Head 요청
    resp, err := http.Head("https://www.time.gov")
    if err != nil {
        t.Fatal(err)
    }
    resp.Body.Close() // close body since we don't use it

    now := time.Now().Round(time.Nanosecond)
    date := resp.Header.Get("Date") // header에서 Date 키의 값 가져오기
    if date == "" { // 값이 존재하지 않는 경우
        t.Fatal("no Date header received from time.gov")
    }

    dt, err := time.Parse(time.RFC1123, date)
    if err != nil {
        t.Fatal(err)
    }

    t.Logf("time.gov: %s (skew %s)", dt, now.Sub(dt))
}
