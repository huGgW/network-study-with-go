package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllowList(t *testing.T) {
	allowList := []string{
		"sage.svg",
		"text/*",
		"*/*/bye.*",
	}

	testCases := []struct {
		path string
		code int
	}{
		{"http://test/static/sage.svg", http.StatusOK},
		{"http://test/static/.secret", http.StatusNotFound},
		{"http://test/static/.dir/secret", http.StatusNotFound},
		{"http://test/static/text/hello.txt", http.StatusOK},
		{"http://test/static/text/ttext/bye.txt", http.StatusOK},
	}

	handler := http.StripPrefix(
		"/static/",
		AllowListMiddleware(
			allowList,
			http.FileServer(http.Dir("../files")),
		),
	)

	for i, c := range testCases {
		w := httptest.NewRecorder()
		r := httptest.NewRequest( http.MethodGet, c.path, nil)
		handler.ServeHTTP(w, r)

		if actual := w.Result().StatusCode; actual != c.code {
            t.Errorf("%d: expected %d; actual %d", i, c.code, actual)
		}
	}
}
