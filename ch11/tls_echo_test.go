package ch11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEchoServerTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverAddress := "localhost:34443"
	maxIdle := time.Second
	server := NewTLSServer(ctx, serverAddress, maxIdle, nil)
	done := make(chan struct{})

	go func() {
		err := server.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil && !strings.Contains(
			err.Error(),
			"use of closed network conection",
		) {
			t.Error(err)
			return
		}
		done <- struct{}{}
	}()
	server.Ready() // 수신 준비 완료까지 blocking

	// Read certificate file
	cert, err := os.ReadFile("cert.pem")
	if err != nil {
		t.Fatal(err)
	}

	// Create new certification pool
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); ok { // append read certificate to pool
		t.Fatal("failed to append certificate to pool")
	}

	tlsConfig := &tls.Config {
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion: tls.VersionTLS12,
		RootCAs: certPool, // 반드시 certPool 내의 인증서를 사용하거나 이로 서명한 연결만 인증
	}

	// tlsConfig를 이용하여 tls 연결 수립
	conn, err := tls.Dial("tcp", serverAddress, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	hello := []byte("hello")
	_, err = conn.Write(hello)
	if err != nil {
		t.Fatal(err)
	}

	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	if actual := b[:n]; !bytes.Equal(hello, actual) {
		t.Fatalf("expected %q; actual %q", hello, actual)
	}

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	cancel()
	<-done
}
