package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"io"
	"log"
	"os"
)

func main() {
	privateKeyBits := flag.Int("key", 2048, "bits for RSA private key")
	keyOut := flag.String("key-out", "-", "File to store RSA private key to; - for stdout")

	commonName := flag.String("cn", "StormForger", "Common Name to be used for CSR")
	csrFormat := flag.String("csr-format", "der-base64", "Output Format: pem, der-base64, der")
	csrOut := flag.String("csr-out", "-", "File to store CSR to; - for stdout")

	flag.Parse()

	if *commonName == "" {
		log.Fatal("Common Name (CN) may not be empty!")
	}

	if *csrFormat != "pem" && *csrFormat != "der-base64" && *csrFormat != "der" {
		log.Fatal("Output format must be pem or der-base64!")
	}

	keyBytes, err := rsa.GenerateKey(rand.Reader, *privateKeyBits)
	if err != nil {
		log.Fatal(err)
	}

	encode("RSA PRIVATE KEY", "pem", x509.MarshalPKCS1PrivateKey(keyBytes), buildWriter(*keyOut))

	subject := pkix.Name{CommonName: *commonName}
	rawSubj := subject.ToRDNSequence()

	asn1Subj, err := asn1.Marshal(rawSubj)
	if err != nil {
		log.Fatal(err)
	}
	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	if err != nil {
		log.Fatal(err)
	}

	encode("CERTIFICATE REQUEST", *csrFormat, csrBytes, buildWriter(*csrOut))
}

func buildWriter(target string) io.Writer {
	var out io.Writer
	if target == "-" {
		out = os.Stdout
	} else {
		var err error
		out, err = os.Create(target)
		if err != nil {
			log.Fatal(err)
		}
	}

	return out
}

func encode(t, format string, data []byte, out io.Writer) {
	switch format {
	case "pem":
		pem.Encode(out, &pem.Block{Type: t, Bytes: data})
	case "der":
		out.Write(data)
	case "der-base64":
		encoder := base64.NewEncoder(base64.StdEncoding, out)
		defer encoder.Close()
		encoder.Write(data)
	}
}
