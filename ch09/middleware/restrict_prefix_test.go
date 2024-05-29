package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestrictPrefix(t *testing.T) {
    // handler 구조
    // 1. StripPrefix middleware => /static/ 접두사 제거
    // 2. RestrictPrefix middleware => '.' prefix로 존재하는 경우 client에게 에러 반환
    // 3. FileServer handler => ../files/ 경로를 root로 하는 정적 파일 서빙
	handler := http.StripPrefix("/static/",
		RestrictPrefix(
		    ".", http.FileServer(http.Dir("../files/")),
		),
	)

    testCases := []struct {
        path string
        code int
    }{
        {"http://test/static/sage.svg", http.StatusOK},
        {"http://test/static/.secret", http.StatusNotFound},
        {"http://test/static/.dir/secret", http.StatusNotFound},
    }

    for i, c := range testCases {
        r := httptest.NewRequest(http.MethodGet, c.path, nil)
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, r)

        actual := w.Result().StatusCode
        if c.code != actual {
            t.Errorf("%d: expected %d; actual %d", i, c.code, actual)
        }
    }
}
