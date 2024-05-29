package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// request body를 drain하고 close함을 보장하기 위한 middleware
func drainAndClose(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	})
}

func TestSimpleMux(t *testing.T) {
	serveMux := http.NewServeMux()
	// path와 handler 등록
	// trailing slash가 없는 경우 절대 경로로서 정확히 일치해야 함.
	// trailing slash가 있는 경우 해당 prefix만 일치하면 그쪽으로 라우팅.
	//     slash가 마지막에 붙지 않아 있는 요청도 해당 절대 경로가 등록되어 있지 않으면
	//     slash를 붙여서 해당 경로를 알려줌 (301 Moved Permanently)
	serveMux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)
	serveMux.HandleFunc(
		"/hello",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "Hello friend.")
		},
	)
	serveMux.HandleFunc(
		"/hello/there/",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "Why, hello there.")
		},
	)
	mux := drainAndClose(serveMux)

	testCases := []struct {
		path     string
		response string
		code     int
	}{
		{"http://test/", "", http.StatusNoContent},
		{"http://test/hello", "Hello friend.", http.StatusOK},
		{"http://test/hello/there/", "Why, hello there.", http.StatusOK},
		{"http://test/hello/there",
			"<a href=\"/hello/there/\">Moved Permanently</a>.\n\n",
			http.StatusMovedPermanently,
		},
		{"http://test/hello/there/you", "Why, hello there.", http.StatusOK},
		{"http://test/hello/and/goodbye", "", http.StatusNoContent},
		{"http://test/something/else/entirely", "", http.StatusNoContent},
		{"http://test/hello/you", "", http.StatusNoContent},
	}

	for i, c := range testCases {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		resp := w.Result()

		if actual := resp.StatusCode; c.code != actual {
			t.Errorf("%d: expected code %d; actual %d", i, c.code, actual)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()

		if actual := string(b); c.response != actual {
			t.Errorf("%d: expected response %q; actual %q", i,
				c.response, actual)
		}
	}
}
