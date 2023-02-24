package conf

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

type Conf struct {
	Server    Server
	Elections []EHS
}

type Server struct {
	Address      string
	TLS          TLS
	BatchMaxSize int
	OpenApiPath  string
	Xroad        XRoad
}

type XRoad struct {
	Certificate string
}

type TLS struct {
	Certificate string
	Key         string
	Insecure    bool
}

type EHS struct {
	Address    string
	Name       string
	ServerName string
	RootCA     string
	ClientCert string
	ClientKey  string
}

func (t TLS) insecure() (cert, key []byte, err error) {
	signer, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate signer: %w", err)
	}
	tmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		SerialNumber:          big.NewInt(1),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
	}
	selfsigned, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, signer.Public(), signer)
	if err != nil {
		return nil, nil, fmt.Errorf("create self-signed certificate: %w", err)
	}
	encsigner, err := x509.MarshalPKCS8PrivateKey(signer)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}
	certpem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: selfsigned,
	})
	keypem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: encsigner,
	})
	return certpem, keypem, nil
}

func (t TLS) X509Certificate() (*tls.Certificate, error) {
	var certPEMblock, keyPEMblock []byte
	var err error
	if !t.Insecure {
		certPEMblock, err = os.ReadFile(t.Certificate)
		if err != nil {
			return nil, fmt.Errorf("get certificate: %w", err)
		}
		keyPEMblock, err = os.ReadFile(t.Key)
		if err != nil {
			return nil, fmt.Errorf("get key: %w", err)
		}
	} else {
		certPEMblock, keyPEMblock, err = t.insecure()
		if err != nil {
			return nil, fmt.Errorf("generate insecure certificate: %w", err)
		}
	}
	crt, err := tls.X509KeyPair(certPEMblock, keyPEMblock)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %w", err)
	}
	return &crt, nil
}

func LoadJSON(path string) (conf Conf, err error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, fmt.Errorf("reading configuration failed: %w", err)
	}
	if err = json.Unmarshal(bytes, &conf); err != nil {
		return conf, fmt.Errorf("configuration unmarshal failed: %w", err)
	}
	return
}
