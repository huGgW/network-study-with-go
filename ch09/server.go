package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"network-study-with-go/ch09/handlers"
	"network-study-with-go/ch09/middleware"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var (
	addr = flag.String("listen", "127.0.0.1:8080", "listen address")
	// tls 연결을 위한 인증서
	cert = flag.String("cert", "", "certificate")
	// tls 연결을 위한 private key
	pkey  = flag.String("key", "", "private key")
	files = flag.String("files", "./files", "static file directory")
)

func main() {
	flag.Parse()

	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully shutdown")
}

func run(addr, files, cert, pkey string) error {
	// Handler 등록 //
	mux := http.NewServeMux()
	// 정적 파일을 서빙하기 위한 경로
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			middleware.RestrictPrefix(
				".", http.FileServer(http.Dir(files)),
			),
		),
	)
	// 기본 route
	mux.Handle("/",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// http.ResponseWriter가 http.Pusher 객체인지 확인
					if pusher, ok := w.(http.Pusher); ok {
						// push시 요청이 client에서 온 것으로 취급하기 떄문에
						// client 관점에서의 리소스 경로를 제공해야 한다.
						targets := []string{
							"/static/style.css",
							"/static/hiking.svg",
						}
						for _, target := range targets {
							// http.Pusher의 경우, 별도 요청 없이 client한테 리소스 push 가능
							if err := pusher.Push(target, nil); err != nil {
								log.Printf("%s push failed: %v", target, err)
							}
						}
					}
					// push 후에 응답을 처리 (index.html)
					// (응답을 먼저 처리한다면 client에서 push를 대기하지 않고 리소스 요청을 할 가능성 있음)
					http.ServeFile(w, r, filepath.Join(files, "index.html"))
				},
			),
		},
	)
	// absolute path '/2'
	// 기본 경로 먼저 방문 후 해당 경로를 방문한다면,
	// 이미 같은 정적 리소스를 사용하고 브라우저상에 cache되어 있을 것이므로
	// 별도의 요청 없이 cache된 자료를 사용할 것이다.
	mux.Handle("/2",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join(files, "index2.html"))
				},
			),
		},
	)

	// 서버 실행 //
	srv := &http.Server{
		Addr: addr,
		Handler: mux,
		IdleTimeout: time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	done := make(chan struct{})
	// interrupt 시 graceful하게 종료하기 위하여
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for {
			if <-c == os.Interrupt {
				// srv.Shutdown: graceful하게 종료
				// 서버의 리스너를 종료하여 수신 연결을 막고,
				// 모든 client의 연결이 끝날 때까지 blocking하여 response의 전송을 마무리하도록 함.
				if err := srv.Shutdown(context.Background()); err != nil {
					log.Printf("shutdown: %v", err)
				}
				close(done)
				return
			}
		}
	}()

	log.Printf("Serving files in %q over %s\n", files, srv.Addr)

	var err error
	if cert != "" && pkey != "" {
		log.Println("TLS enabled")
		err = srv.ListenAndServeTLS(cert, pkey)
	} else {
		err = srv.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		err = nil
	}
	
	<- done

	return err
}
