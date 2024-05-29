package middleware

import (
	"log"
	"net/http"
	"time"
)

// 보통은 이 예시와 다르게 하나의 middleware는 하나의 기능만을 하도록
// 작은 미들웨어를 여러개 생성.
func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            // 경우에 따라 middleware에서 바로 client에 응답
            if r.Method == http.MethodTrace {
                http.Error(w, "Method not allowed",
                    http.StatusMethodNotAllowed)
            }
            // 헤더 설정
            w.Header().Set("X-Content-Type-Options", "nosniff")

            start := time.Now()
            // 대부분 middleware는 주어진 handler를 그대로 호출.
            // 주어진 handler에서 client에 응답
            next.ServeHTTP(w, r)

            // metric 수집, 로그 기록
            log.Printf("Next handler duration %v", time.Now().Sub(start))
        },
    )
}

