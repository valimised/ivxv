package internal

import (
	"bytes"
	pem2 "encoding/pem"
	"fmt"
)

const (
	X509Header           = "PUBLIC KEY"
	PKCS8Header          = "PRIVATE KEY"
	EncryptedPKCS8Header = "ENCRYPTED PRIVATE KEY"
)

// Der2Pem unmarshalls DER to PEM.
//
// header must be valid PEM header.
func Der2Pem(der []byte, header string) ([]byte, error) {
	var b bytes.Buffer
	err := pem2.Encode(&b, &pem2.Block{
		Type:  header,
		Bytes: der,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert DER to PEM: %v", err)
	}

	return b.Bytes(), err
}

// Pem2Der unmarshalls PEM to DER. If pem is DER bytes, nothing is done.
// Also ensures that PEM header from pem matches provided header.
func Pem2Der(pem []byte, header string) ([]byte, error) {
	// If block == nil, then input is DER
	block, der := pem2.Decode(pem)
	if block != nil {
		if block.Type != header {
			return nil, fmt.Errorf("expected PEM header %s, but got %s", header, block.Type)
		}

		// No trailing bytes allowed
		if len(der) != 0 {
			return nil, fmt.Errorf("trailing bytes left after PEM decoding")
		}

		der = block.Bytes
	}

	return der, nil
}
