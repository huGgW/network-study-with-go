//go:build exclude
// For Example Only

package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
)

// Dependency Injection on a single handler with Closure (Anonymous Function)
func diSingle() {
	dbHandler := func(db *sql.DB) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				err := db.Ping()
				// 데이터베이스 관련 작업 수행
			},
		)
	}

	http.Handle("/three", dbHandler(db))
}

// 여러 개의 핸들러를 구조체 메서드를 이용하여 의존성 주입
type Handlers struct {
	db  *sql.DB
	log *log.Logger
}

func (h *Handlers) Handler1() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := h.db.Ping()
			if err != nil {
				h.log.Printf("db ping: %v", err)
			}

			// 데이터베이스 관련 작업 수행
		},
	)
}

func (h *Handlers) Handler2() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// ...
		},
	)
}

func diMultiple() {
	h := &Handlers{
		db:  db,
		log: log.New(os.Stderr, "handlers: ", log.Lshortfile),
	}
	http.Handle("/one", h.Handler1())
	http.Handle("/two", h.Handler2())
	// ...
}
