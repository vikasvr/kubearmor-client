package install

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"time"
)

func GeneratePki(namespace string, serviceName string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ca, cakey, err := GenerateCA()
	if err != nil {
		return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), err
	}
	csr, csrkey, err := GenerateCSR(namespace, serviceName)
	if err != nil {
		return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), err
	}
	crt, err := SignCSR(ca, cakey, csr, csrkey)
	if err != nil {
		return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &cakey.PublicKey, cakey)
	if err != nil {
		return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), err
	}
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), err
	}
	crtPEM := new(bytes.Buffer)
	err = pem.Encode(crtPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt,
	})
	if err != nil {
		return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), err
	}
	crtKeyPEM := new(bytes.Buffer)
	err = pem.Encode(crtKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(csrkey),
	})
	if err != nil {
		return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{}), err
	}
	return caPEM, crtPEM, crtKeyPEM, nil
}

func GenerateCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(123),
		Subject: pkix.Name{
			Organization: []string{"kubearmor"},
			Country:      []string{"US"},
			Province:     []string{""},
			CommonName:   "kubearmor-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(3, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return &x509.Certificate{}, &rsa.PrivateKey{}, errors.New("cannot generate ca private key")
	}

	return ca, caPrivKey, nil
}

func GenerateCSR(namespace string, serviceName string) (*x509.Certificate, *rsa.PrivateKey, error) {
	csr := &x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			Organization: []string{"kubearmor"},
			Country:      []string{"US"},
			Province:     []string{""},
			CommonName:   "kubearmor-webhook",
		},
		DNSNames: []string{
			serviceName + "." + namespace + ".svc",
			serviceName + "." + namespace + ".svc.cluster.local",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(3, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
	}
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return &x509.Certificate{}, &rsa.PrivateKey{}, errors.New("cannot generate csr private key")
	}
	return csr, certPrivKey, nil
}

func SignCSR(caCrt *x509.Certificate, caKey *rsa.PrivateKey, csrCrt *x509.Certificate, csrKey *rsa.PrivateKey) ([]byte, error) {
	certBytes, err := x509.CreateCertificate(rand.Reader, csrCrt, caCrt, &csrKey.PublicKey, caKey)
	if err != nil {
		return []byte{}, errors.New("cannot sign the csr")
	}
	return certBytes, nil
}
