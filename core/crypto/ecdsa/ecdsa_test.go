package ecdsa

import (
	"os"
	"strings"
	"testing"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/x509"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

func storePKCS8PrivateKey(key crypto.KeyInfo) ([]byte, error) {
	return x509.PEMEncodePKCS8PrivateKey(key)
}

func storeX509PublicKey(key crypto.KeyInfo) ([]byte, error) {
	return x509.X509MarshalPEM(key)
}

func sign(signer crypto.Signer, data []byte) ([]byte, error) {
	return signer.Sign(data)
}

func verify(verifier crypto.SignatureVerifier, signature, data []byte) error {
	return verifier.Verify(signature, data)
}

func TestSignAndVerify(t *testing.T) {
	data := []byte("Hello World!")
	for _, g := range group.All() {
		// Skip non-NIST groups
		if !strings.Contains(string(g.Name()), "NIST") {
			continue
		}

		t.Run(string(g.Name()), func(t *testing.T) {
			// Generate new ECDSA private key with only Group given, btw,
			// we derive public key from a private
			priv, err := New(g)
			if err != nil {
				t.Fatal(err)
			}
			// Cast to interface
			var privI crypto.SigningPrivateKey = priv
			var pubI = priv.PublicKey()

			// Since generated keys is likely to be marshalled after creation (and
			// later reused, on unmarshal), we simulate storing in I/O
			privBytes, err := storePKCS8PrivateKey(privI)
			if err != nil {
				t.Fatal(err)
			}
			os.WriteFile(string(g.Name()), privBytes, 0777)

			pubBytes, err := storeX509PublicKey(pubI)
			if err != nil {
				t.Fatal(err)
			}

			// Simulate reading private key from I/O to concrete implementation
			// first
			_, privIRestored, err := x509.DecodePKCS8PrivateKey[crypto.SigningPrivateKey](privBytes)
			if err != nil {
				t.Fatal(err)
			}

			// SignAll custom data
			signed, err := sign(privIRestored, data)
			if err != nil {
				t.Fatal(err)
			}

			// Simulate reading public key from I/O to concrete implementation
			// first
			_, pubIRestored, err := x509.X509Unmarshal[crypto.SigningPublicKey](pubBytes)
			if err != nil {
				t.Fatal(err)
			}

			// VerifyAll signature
			err = verify(pubIRestored, signed, data)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSignAndVerifyManyData(t *testing.T) {
	data1 := []byte("Hello World!")
	data2 := []byte("I'm Alice")
	data3 := []byte("Good weather")
	data4 := []byte("Happy coding :)")
	dataSet := [][]byte{data1, data2, data3, data4}

	for _, g := range group.All() {
		// Skip non-NIST groups
		if !strings.Contains(string(g.Name()), "NIST") {
			continue
		}
		t.Run(string(g.Name()), func(t *testing.T) {
			// Generate new ECDSA signing key
			var signingKeyI crypto.SigningPrivateKey
			signingKeyI, err := New(g)
			if err != nil {
				t.Fatal(err)
			}

			// SignAll many data
			signature, err := crypto.SignAll(signingKeyI, dataSet...)
			if err != nil {
				t.Fatal(err)
			}

			// VerifyAll many data
			err = crypto.VerifyAll(signingKeyI.PublicKey(), signature, dataSet...)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
