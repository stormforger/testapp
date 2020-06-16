package server

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/fullsailor/pkcs7"
	"github.com/sirupsen/logrus"
)

type tlsInspect struct {
	ServerName string `json:"server_name,omitempty"`
	Status     string `json:"status,omitempty"`
	Subject    string `json:"subject,omitempty"`
}

type x509Handlers struct {
	CACertPEMData        []byte
	CACertPKCS7DERBase64 []byte
	CACert               *x509.Certificate
	CAPrivateKey         interface{}
}

func RegisterX509ESTHandlers(router *mux.Router, serverCertificateFile, serverPrivateKeyFile string) error {
	caCertPEMData, err := ioutil.ReadFile(serverCertificateFile)
	if err != nil {
		return err
	}
	caPrivateKeyPEMData, err := ioutil.ReadFile(serverPrivateKeyFile)
	if err != nil {
		return err
	}

	x, err := buildX509Handlers(caCertPEMData, caPrivateKeyPEMData)
	if err != nil {
		return err
	}
	router.HandleFunc("/.well-known/est/cacerts", x.estCACertsHandler)
	router.HandleFunc("/.well-known/est/simpleenroll", x.estEnrollHandler).Methods("POST")
	router.HandleFunc("/.well-known/est/simplereenroll", x.estEnrollHandler).Methods("POST")
	router.HandleFunc("/x509/inspect", clientCertInspectHandler)

	return nil
}

func clientCertInspectHandler(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		http.Error(w, "No TLS connection", http.StatusBadRequest)
		return
	}
	certs := r.TLS.PeerCertificates

	tlsInspection := &tlsInspect{
		ServerName: r.TLS.ServerName,
	}

	if len(certs) == 0 {
		tlsInspection.Status = "no_cert"

		logrus.Warnf("x509inspect: No cert request from %v\n", r.RemoteAddr)
	} else {
		cert := certs[0]
		tlsInspection.Subject = cert.Subject.String()
		tlsInspection.Status = "client_cert"

		logrus.Infof("x509inspect: Hello %s from %v\n", cert.Subject, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	err := e.Encode(tlsInspection)
	if err != nil {
		http.Error(w, "Cannot marshal TLS information.", http.StatusInternalServerError)
		logrus.Errorf("json marshal: %v", err)
	}
}

func (x *x509Handlers) estEnrollHandler(w http.ResponseWriter, r *http.Request) {
	base64Decoder := base64.NewDecoder(base64.StdEncoding, r.Body)

	b, err := ioutil.ReadAll(base64Decoder)
	if err != nil {
		http.Error(w, "Could not decode base64 body", http.StatusBadRequest)
		return
	}

	clientCSR, err := x509.ParseCertificateRequest(b)
	if err != nil {
		http.Error(w, "Could not decode CSR", http.StatusBadRequest)
		return
	}
	if err = clientCSR.CheckSignature(); err != nil {
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	clientCRTTemplate := x509.Certificate{
		Signature:          clientCSR.Signature,
		SignatureAlgorithm: clientCSR.SignatureAlgorithm,

		PublicKeyAlgorithm: clientCSR.PublicKeyAlgorithm,
		PublicKey:          clientCSR.PublicKey,

		SerialNumber: big.NewInt(2),
		Issuer:       x.CACert.Subject,
		Subject:      clientCSR.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// create client certificate from template and CA public key
	clientCRTRaw, err := x509.CreateCertificate(rand.Reader, &clientCRTTemplate, x.CACert, clientCSR.PublicKey, x.CAPrivateKey)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/pkcs7-mime; smime-type=certs-only")
	w.Header().Set("Content-Transfer-Encoding", "base64")

	fmt.Fprint(w, base64.StdEncoding.EncodeToString(clientCRTRaw))
}

func (x *x509Handlers) estCACertsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pkcs7-mime; smime-type=certs-only")
	w.Header().Set("Content-Transfer-Encoding", "base64")

	w.Write(x.CACertPKCS7DERBase64)
}

func pemToPKCS7DERBase64(input []byte) ([]byte, error) {
	pemData := make([]byte, len(input))
	copy(pemData, input)
	data := []byte{}

	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}
		data = append(data, (*block).Bytes...)
	}

	// Build a PKCS#7 degenerate "certs only" structure from
	// that ASN.1 certificates data.
	data, err := pkcs7.DegenerateCertificate(data)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	base64Encoder := base64.NewEncoder(base64.StdEncoding, &b)

	base64Encoder.Write(data)
	base64Encoder.Close()

	return b.Bytes(), nil
}

func buildX509Handlers(caCertPEMData, caPrivateKeyFile []byte) (x509Handlers, error) {
	pemBlock, _ := pem.Decode(caCertPEMData)
	if pemBlock == nil {
		return x509Handlers{}, fmt.Errorf("pem.Decode failed")
	}
	caCRT, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return x509Handlers{}, err
	}

	caPKCS7, err := pemToPKCS7DERBase64(caCertPEMData)
	if err != nil {
		return x509Handlers{}, err
	}

	pemBlock, _ = pem.Decode(caPrivateKeyFile)
	if pemBlock == nil {
		return x509Handlers{}, fmt.Errorf("pem.Decode failed")
	}

	caPrivateKey, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return x509Handlers{}, fmt.Errorf("private key: %v", err)
	}

	return x509Handlers{
		CACertPEMData:        caCertPEMData,
		CACertPKCS7DERBase64: caPKCS7,
		CACert:               caCRT,
		CAPrivateKey:         caPrivateKey,
	}, nil
}
