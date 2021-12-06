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
	host = flag.String("host", "localhost",
		"Certificate's comma-separated host names and IPs")
	certFn = flag.String("cert", "cert.pem", "certificate file name")
	keyFn  = flag.String("key", "key.pem", "private key file name")
)

func main() {
	/*
		❯ ./cert --host 22
		fmt.Println(*host, *certFn, *keyFn)
		localhost cert.pem key.pem
		fmt.Println(*host, *certFn, *keyFn)
		22 cert.pem key.pem

	*/
	flag.Parse()

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Fatal(err)
	}
	notBefore := time.Now()
	template := x509.Certificate{
		// Each certificate needs a serial number, which a certificate authority
		//typically assigns
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"wzk"},
		},
		NotBefore: notBefore,
		NotAfter:  notBefore.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		// Since you want to use this certificate for client
		//authentication, you must include the x509.ExtKeyUsageClientAuth value
		// If you omit this value, the server won’t be able to verify the certificate when
		//presented by the client.
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
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

	// 生成私钥
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	// 生成证书
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatal(err)
	}

	cert, err := os.Create(*certFn)
	if err != nil {
		log.Fatal(err)
	}

	err = pem.Encode(cert, &pem.Block{
		Type:    "CERTIFICATE",
		Headers: nil,
		Bytes:   der,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", *certFn)

	// generate key(私人密钥)
	key, err := os.OpenFile(*keyFn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}

	privKey, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}

	err = pem.Encode(key, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privKey})
	if err != nil {
		log.Fatal(err)
	}
	if err := key.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", *keyFn)
}
