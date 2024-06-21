package ch11

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

func TestClientTLS(t *testing.T) {
	// httptest.NewTLSServer: 테스트 위한 새로운 인증서 생성, tls 세부 환경구성 자동 처리.
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 인증서 없는 경우 (http 통신), https로 리다이렉트
			if r.TLS == nil {
				u := "https://" + r.Host + r.RequestURI
				http.Redirect(w, r, u, http.StatusMovedPermanently)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf(
			"expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode,
		)
	}

	// Define new transport
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{tls.CurveP256}, // p256은 시간차 공격에 저항 있음, 최소 TLS 1.2 이상 사용
			MinVersion:       tls.VersionTLS12,
		},
	}

	// Use created transport in http2
	// client를 새로 생성한 transport로 override할 것이므로 http/2를 사용하기 위해서는 반드시 필요.
	err = http2.ConfigureTransport(tp)
	if err != nil {
		t.Fatal(err)
	}

	// Override transport configure to created one for client
	client2 := &http.Client{Transport: tp}

	_, err = client2.Get(ts.URL)
	if err == nil || !strings.Contains(err.Error(), "certificate signed by unknown authority") {
		t.Fatalf("expected unknown authority error; actual: %q", err)
	}

	// 인증서 검증 건너뛰기
	tp.TLSClientConfig.InsecureSkipVerify = true

	resp, err = client2.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
}

func TestClientTLSGoogle(t *testing.T) {
	conn, err := tls.DialWithDialer(
		&net.Dialer { Timeout: 30 * time.Second },
		"tcp",
		"www.google.com:443",
		&tls.Config {
			CurvePreferences: []tls.CurveID{ tls.CurveP256 },
			MinVersion: tls.VersionTLS12,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	state := conn.ConnectionState()
	t.Logf("TLS 1.%d", state.Version-tls.VersionTLS10)
	t.Log(tls.CipherSuiteName(state.CipherSuite))
	t.Log(state.VerifiedChains[0][0].Issuer.Organization[0])

	_ = conn.Close()
}
