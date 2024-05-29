package handlers

import (
	"html/template"
	"io"
	"net/http"
)

var t = template.Must(
	template.New("hello").Parse("Hello, {{.}}!"),
)

func DefaultHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
		    // 서버의 request body는 닫을 때 암묵적으로 소비되지 않음.
		    // 따라서 반드시 handling이 끝난 후 소비해야 한다!
		    // TCP 연결은 선택적으로 닫는다.
			defer func(r io.ReadCloser) {
				_, _ = io.Copy(io.Discard, r) // request body 소비
				_ = r.Close()
			}(r.Body)

			var b []byte

			switch r.Method {
			case http.MethodGet:
				b = []byte("friend")

			case http.MethodPost:
				var err error
				b, err = io.ReadAll(r.Body)
				if err != nil {
					http.Error(
						w, "Internal server error",
						http.StatusInternalServerError,
					)
					return
				}

			default:
			    // RFC 표준 준수 X
			    // (Allow 헤더에 해당 핸들러의 가능한 메서드들 명시하지 않았음.)
				http.Error(
					w, "Method not allowed",
					http.StatusMethodNotAllowed,
				)
				return
			}

			_ = t.Execute(w, string(b))
		},

	)
}
