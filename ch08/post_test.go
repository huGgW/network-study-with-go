package ch08

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type User struct {
	First string
	Last  string
}

// JSON 테스트

// Post 요청 처리하는 함수 반환
func handlePostUser(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Server의 request body의 경우 close 전 반드시 명시적으로 소비해야됨.
		defer func(r io.ReadCloser) {
			_, _ = io.Copy(io.Discard, r)
			_ = r.Close()
		}(r.Body)

		// Post가 아닌 경우 405 코드 반환
		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		var u User
		// request body를 json 디코딩 시도
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			t.Error(err)
			http.Error(w, "Decode Failed", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func TestPostUser(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(handlePostUser(t)),
	)
	defer ts.Close()

	// 잘못된 타입 요청에 대하여 처리 테스트
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	// 405 코드가 왔는지 체크
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusMethodNotAllowed, resp.StatusCode)
	}

	// json 으로 인코딩
	buf := new(bytes.Buffer)
	u := User{First: "Adam", Last: "Woodbeck"}
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	// Post 정상적으로 동작하는지 체크
	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusAccepted { // Accepted되었는지 체크
		t.Fatalf("expected status %d; actual status %d",
			http.StatusAccepted, resp.StatusCode)
	}

	_ = resp.Body.Close()
}

// multipart
func TestMultipartPost(t *testing.T) {
	// Multipart Request Body 생성

	reqBody := new(bytes.Buffer) // buffer 생성

	// buffer를 wrap하는 multipart writer 생성
	// 초기화시 random boundary 생성
	w := multipart.NewWriter(reqBody)

	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"description": "Form values with attached files",
	} {
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
	} {
		// multipart 섹션 writer를 생성
		filePart, err := w.CreateFormFile(
			fmt.Sprintf("file%d", i+1),
			filepath.Base(file),
		)
		if err != nil {
			t.Fatal(err)
		}

		// File 열기
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		// file을 filePart로 copy하기
		_, err = io.Copy(filePart, f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	err := w.Close() // multipart writer를 닫아야 request body가 올바르게 boundary 추가
	if err != nil {
		t.Fatal(err)
	}


	// 요청 서버로 전송

	// 60초 타임아웃 설정
	ctx, cancel := context.WithTimeout(
		context.Background(), 60*time.Second,
	)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://httpbin.org/post",
		reqBody,
	)
	// Content Type 헤더 설정을 통해 multipart 데이터 전송을 알림
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d; actual status %d",
	http.StatusOK, resp.StatusCode)
	}

	t.Logf("\n%s", b)
}
