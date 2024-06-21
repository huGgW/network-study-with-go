package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/huGgW/network-study-with-go/ch12/housework/v1"
	"google.golang.org/grpc"
)

var addr, certFn, keyFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
	flag.StringVar(&certFn, "cert", "cert.pem", "certificate file")
	flag.StringVar(&keyFn, "key", "key.pem", "private key file")
}

func main() {
	flag.Parse()

	server := grpc.NewServer() // 새로운 grpc 서버 인스턴스 생성
	rosie := new(Rosie)
	housework.RegisterRobotMaidServer(server, rosie) // server로 roise 등록

	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening for TLS connections on %s ...", addr)
	log.Fatal(
		server.Serve(
			tls.NewListener(
				listener,
				&tls.Config{
					Certificates:             []tls.Certificate{cert},
					CurvePreferences:         []tls.CurveID{tls.CurveP256},
					MinVersion:               tls.VersionTLS12,
					PreferServerCipherSuites: true,
				},
			),
		),
	)
}
