package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

var (
	host   = flag.String("host", "localhost", "Certificate's comma-seperated host names and IPs")
	certFn = flag.String("cert", "cert.pem", "certificate file name")
	keyFn  = flag.String("key", "key.pem", "private key file name")
)

func main() {
	flag.Parse()

	// create random 128bit
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Fatal(err)
	}

	notBefore := time.Now()
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Adam Woodbeck"},
		},
		NotBefore: notBefore,
		NotAfter:  notBefore.Add(10 * 365 * 24 * time.Hour),
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth, // client 인증에 사용할 것이므로 이를 반드시 포함
		},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	for _, h := range strings.Split(*host, ",") {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// P-256 타원 곡선 이용하여 새료운 ECDSA private key 생성
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// 암호화를 위해 무작위 값을 추출하기 위한 entropy source, 새로운 인증서 생성 위한 탬플릿,
	// 상위 인증서, public key, private key를 받고,
	// DER(Distinguished Encoding Rules) 포맷으로 encoding된 인증서가 포함된 byte slice 반환.
	der, err := x509.CreateCertificate(rand.Reader, &template,
		&template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatal(err)
	}

	cert, err := os.Create(*certFn)
	if err != nil {
		log.Fatal(err)
	}

	// DER 포맷의 byte slice를 pem.Block 객체를 생성하여 PEM 포맷으로 인코딩.
	err = pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err != nil {
		log.Fatal(err)
	}

	if err := cert.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", *certFn)

	key, err := os.OpenFile(*keyFn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}

	privKey, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}

	err = pem.Encode(key, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := key.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", *keyFn)
}
