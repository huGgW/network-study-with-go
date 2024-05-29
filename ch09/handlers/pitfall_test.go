package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWriteHeader(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
        // Write Body First -> 200을 가정, OK로 response가 반환되었을 가능성 존재
        // 즉, Write 시 status code 설정 안되어있으면 암묵적으로 w.WriteHeader 호출
	    _, _ = w.Write([]byte("Bad request")) 
	    w.WriteHeader(http.StatusBadRequest) // 반영되지 않음
	}
    r := httptest.NewRequest(http.MethodGet, "http://test", nil)
    w := httptest.NewRecorder()
    handler(w, r)
    t.Logf("Response status: %q", w.Result().Status)

    handler = func(w http.ResponseWriter, r *http.Request) {
        // Status Code를 먼저 write
        w.WriteHeader(http.StatusBadRequest)
        // Body를 Write할 때 앞서 설정했던 status code 제대로 반영됨
        _, _ = w.Write([]byte("Bad request"))
    }
    r = httptest.NewRequest(http.MethodGet, "http://test", nil)
    w = httptest.NewRecorder()
    handler(w, r)
    t.Logf("Response status: %q", w.Result().Status)

    // http.Error method 통해 에러 반환 간단하게 가능
}
