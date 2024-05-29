package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddleware(t *testing.T) {
    // 대기 시간동안 타이머 작동
    // 타이머가 끝나기 전까지 http.Handler가 반환하지 않으면
    // http.Handler를 block하고 503 Service Unavailable을 반환
    handler := http.TimeoutHandler (
        http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusNoContent)
            time.Sleep(time.Minute)
        }),
        time.Second,
        "Timed out while reading response",
    )

    r := httptest.NewRequest(http.MethodGet, "http://test/", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, r)

    resp := w.Result()
    if resp.StatusCode != http.StatusServiceUnavailable {
        t.Fatalf("unexpected status code: %q", resp.Status)
    }

    b, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Fatal(err)
    }
    _ = resp.Body.Close()

    if actual := string(b); actual != "Timed out while reading response" {
        t.Logf("unexpected body: %q", actual)
    }
}
