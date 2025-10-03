package elgamal

import (
	"bytes"
	"encoding/asn1"
	"fmt"
	"testing"
	"tivi.io/core/crypto/elgamal/nizkp"
	"tivi.io/core/test/fixtures"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/util"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

func TestPrivateKeyElementASN1MarshalAndUnmarshal(t *testing.T) {
	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			var keyI crypto.DecryptionKey
			parameters := NewParameters(g)

			// private key element is a random scalar
			rnd, err := group.RandomScalar(parameters.Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			// Create private key
			keyI, err = NewPrivateKey(parameters, rnd)
			if err != nil {
				t.Fatal(err)
			}

			// ASN.1 marshal private key element
			der, err := keyI.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			// ASN.1 unmarshal private key element
			rnd2, err := ASN1UnmarshalPrivateElement(g, der)
			if err != nil {
				t.Fatal(err)
			}

			if rnd.Equal(rnd2) != nil {
				t.Fatal("unequal private key elements")
			}
		})
	}
}

func TestPublicKeyElementASN1MarshalAndUnmarshal(t *testing.T) {
	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			var keyI crypto.DecryptionKey
			parameters := NewParameters(g)

			// private key element is a random scalar
			rnd, err := group.RandomScalar(parameters.Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			// Create private key
			keyI, err = NewPrivateKey(parameters, rnd)
			if err != nil {
				t.Fatal(err)
			}

			var pkeyI crypto.EncryptionKey

			// Derive public key element from private key
			pkeyI = keyI.EncryptionKey()

			// ASN.1 marshal public key element
			der, err := pkeyI.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			// ASN.1 unmarshal public key element
			pub, err := ASN1UnmarshalPublicElement(g, der)
			if err != nil {
				t.Fatal(err)
			}

			pub2 := pkeyI.(*PublicKey).Public()

			if pub.Equal(pub2) != nil {
				t.Fatal("unequal public key elements")
			}
		})
	}
}

func TestDecryptionProofASN1MarshalAndUnmarshal(t *testing.T) {
	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			plaintext := []byte("Hello World!")

			challenge, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			response, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			// ElGamal ciphertext a
			A, err := group.RandomElement(g)
			if err != nil {
				t.Fatal(err)
			}

			// ElGamal ciphertext b
			B, err := group.RandomElement(g)
			if err != nil {
				t.Fatal(err)
			}

			// Plaintext as group element
			E, err := g.Encode(plaintext)
			if err != nil {
				t.Fatal(err)
			}

			ciphertext := NewCiphertext(A, B)

			dec := NewDecryption(g, E)

			transcript := nizkp.NewProof(challenge, response)

			proof := NewDecryptionProof(ciphertext, dec, transcript)

			der, err := proof.MarshalASN1()
			if err != nil {
				t.Fatal(err)
			}

			proof2, err := ASN1UnmarshalDecryptionProof(g, der)
			if err != nil {
				t.Fatal(err)
			}

			if proof.ciphertext.A().Equal(proof2.ciphertext.A()) != nil {
				t.Fatal("ElGamal proof ciphertext a elements differ")
			}

			if proof.ciphertext.B().Equal(proof2.ciphertext.B()) != nil {
				t.Fatal("ElGamal proof ciphertext b elements differ")
			}

			if proof.decryption.Value().Equal(proof2.decryption.Value()) != nil {
				t.Fatal("ElGamal decryptions differ")
			}

			if proof.Proof().Challenge().Equal(proof2.Proof().Challenge()) != nil {
				t.Fatal("ElGamal transcripts differ")
			}

			if proof.Proof().Response().Equal(proof2.Proof().Response()) != nil {
				t.Fatal("ElGamal proofs differ")
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("Imagine plaintext =]")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			// private key element
			x, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			parameters := NewParameters(g)
			key, err := NewPrivateKey(parameters, x)
			if err != nil {
				t.Fatal(err)
			}

			// Encryption randomness
			rnd, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			ct, err := key.publicKey.Encrypt(rnd, plaintext)
			if err != nil {
				t.Fatal(err)
			}

			// Decrypt
			dec, err := key.Decrypt(ct, true)
			if err != nil {
				t.Fatal(err)
			}

			plaintext2, err := dec.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(plaintext, plaintext2) {
				t.Fatal("decrypted and initial plaintexts differ")
			}
		})
	}
}

func TestEncryptProvableDecryptVerify(t *testing.T) {
	plaintext := []byte("0000.101")
	salt := []byte("Unique ID for a proof of a given decryption, that allows you to differ among other decryptions")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			// ElGamal parameters
			parameters := NewParameters(g)

			// private key element is a random scalar
			x, err := group.RandomScalar(parameters.Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			key, err := NewPrivateKey(parameters, x)
			if err != nil {
				t.Fatal(err)
			}

			// Randomness used in encryption
			rnd, err := group.RandomScalar(key.Parameters().Group().Order())

			// Encrypt
			ct, err := key.publicKey.Encrypt(rnd, plaintext)
			if err != nil {
				t.Fatal(err)
			}

			// Randomness used in ElGamal proof creation
			rnd2, err := group.RandomScalar(key.Parameters().Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			decrypted, err := key.Decrypt(ct, true)
			if err != nil {
				t.Fatal(err)
			}

			decrypted2, err := key.Public().DecryptWithRandomness(rnd, ct, true)
			if err != nil {
				t.Fatal(err)
			}

			// Decrypt with proofs returned
			proof, err := key.Prove(rnd2, ct, decrypted, salt)
			if err != nil {
				t.Fatal(err)
			}

			proof2, err := key.Prove(rnd2, ct, decrypted2, salt)
			if err != nil {
				t.Fatal(err)
			}

			// decrypted plaintext
			plaintext2, err := decrypted.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			plaintext3, err := decrypted2.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			// Should be the same
			if !bytes.Equal(plaintext, plaintext2) {
				t.Fatal("decrypted and initial plaintexts differ")
			}

			// Should be the same
			if !bytes.Equal(plaintext, plaintext3) {
				t.Fatal("decrypted and initial plaintexts differ")
			}

			// Should fail, because ProvableDecryptAll used salt and therefore
			// VerifyAll should use as well
			if err = key.publicKey.Verify(proof, nil); err == nil {
				t.Fatal(err)
			}

			// VerifyAll with salt should succeed, since ProvableDecrypted uses
			// salt
			if err = key.publicKey.Verify(proof, salt); err != nil {
				t.Fatal(err)
			}

			if err = key.publicKey.Verify(proof2, salt); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestE2E is a complete process of ElGamal encryption scheme:
//  1. ElGamal key parameters generation
//  2. ElGamal private key element generation
//  3. ElGamal private key generation behind the crypto.DecryptionKey interface
//  4. Random scalar generation that will be used in encryption
//  5. Plaintext encryption, result is a byte slice that can be stored in I/O
//  6. Ciphertext decryption, decrypted value is crypto.Decryption interface
//  7. Right after the decryption we extract an initial plaintext from it
//  8. We may or may not immediately generate proofs of a correct decryption,
//     in case we don't do that right away, we store decrypted values in I/O
//  9. Admin chooses predefined and agreed with auditor proof options that TIVI
//     Core provides, and composes a chain to be verified later by auditor
//  10. Admin also creates a random scalar to be used in proofs generation
//  11. Admin generates a proof chain
//  12. Admin stores proof chain in I/O along with necessary data that is somehow
//     linked to the proof chain generated. NB! It is a responsibility of admin
//     to store proof chain and its related data correctly
//  13. Auditor imports proof chain and its related data from I/O
//  14. Auditor composes the same proof chain options as admin did
//     (remember that these options are already been discussed and agreed on)
//  15. Auditor verifies proof chain
func TestE2E(t *testing.T) {
	plaintext := []byte("0000.101")
	salt := []byte("Unique ID for a proof of a given decryption, that allows you to differ among other decryptions")

	for _, g := range group.All() {
		t.Run(fmt.Sprintf("Testing e2e encryption scheme for %v group", g.Name()), func(t *testing.T) {
			t.Logf("Group %s is chosen for ElGamal key parameters", g.Name())

			params := NewParameters(g)
			t.Log("ElGamal key parameters generation")

			priv, err := group.RandomScalar(params.Group().Order())
			if err != nil {
				t.Fatal(err)
			}
			t.Log("ElGamal private key element generation")

			var key crypto.DecryptionKey
			key, err = NewPrivateKey(params, priv)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("ElGamal private key generation behind the crypto.DecryptionKey interface")

			s, err := group.RandomScalar(key.Parameters().Group().Order())
			t.Log("Random scalar generation that will be used in encryption")

			ct, err := key.EncryptionKey().Encrypt(s, plaintext)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("Plaintext encryption, result is a byte slice that can be stored in I/O")

			decrypted, err := key.Decrypt(ct, true)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("Ciphertext decryption, decrypted value is crypto.Decryption interface")

			plaintext2, err := decrypted.Plaintext()
			if err != nil {
				t.Fatal(err)
			}
			t.Log("Right after the decryption we extract an initial plaintext from it")

			// Compare that initial plaintext and decrypted one match
			if !bytes.Equal(plaintext, plaintext2) {
				t.Fatal("initial and decrypted plaintexts differ")
			}

			decryptedDer, err := decrypted.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(`We may or may not immediately generate proofs of a correct decryption,
                         in case we don't do that right away, we store decrypted values in I/O`)

			opts := []crypto.ProofOpts{
				crypto.NewKeyPairProofOpts(),                    // key pair proof
				crypto.NewCiphertextProofOpts(ct, decryptedDer), // ciphertext proof
			}
			t.Log(`Admin chooses available proof options that TIVI Core provides, and
			             composes a chain to be verified by auditor`)

			s, err = group.RandomScalar(key.Parameters().Group().Order())
			if err != nil {
				t.Fatal(err)
			}
			t.Log("Admin also creates a random scalar to be used in proofs generation")

			proof, err := key.ProofProve(s, salt, opts)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("Admin generates a proof chain")

			proofFixture := fixtures.CiphertextProof{
				Ciphertext: ct,
				Decryption: decryptedDer,
				Proof:      proof,
			}
			proofDer, err := asn1.Marshal(proofFixture)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(`Admin stores proof chain in I/O along with necessary data that is somehow
			             linked to the proof chain generated. NB! It is a responsibility of admin
			             to store proof chain and its related data correctly`)

			proofFixture = fixtures.CiphertextProof{}
			_, err = asn1.Unmarshal(proofDer, &proofFixture)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("Auditor imports proof chain and its related data from I/O")

			opts = []crypto.ProofOpts{
				crypto.NewKeyPairProofOpts(),
				crypto.NewCiphertextProofOpts(proofFixture.Ciphertext, proofFixture.Decryption),
			}
			t.Log(`Auditor composes the same proof chain options as admin did
                         (remember that these options are already been discussed and agreed on)`)

			if err = key.EncryptionKey().ProofVerify(proof, salt, opts); err != nil {
				t.Fatal(err)
			}
			t.Log("Auditor verifies proof chain")

			// This steps should fail, since salt hasn't been used during proof chain generation
			if err = key.EncryptionKey().ProofVerify(proof, nil, opts); err == nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEncryptDecryptManyData(t *testing.T) {
	plaintext1 := []byte("Hello World!")
	plaintext2 := []byte("I'm Alice")
	plaintext3 := []byte("Good weather")
	plaintext4 := []byte("Happy coding :)")
	plaintexts := [][]byte{plaintext1, plaintext2, plaintext3, plaintext4}

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			// ElGamal parameters
			parameters := NewParameters(g)

			// ElGamal private key element
			rnd, err := group.RandomScalar(parameters.Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			// ElGamal private key
			key, err := NewPrivateKey(parameters, rnd)
			if err != nil {
				t.Fatal(err)
			}

			rnd, err = group.RandomScalar(key.Parameters().Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			ciphertext, err := util.EncryptAll(rnd, key.EncryptionKey(), nil, plaintexts...)
			if err != nil {
				t.Fatal(err)
			}

			_, decs, err := util.DecryptAll(key, nil, ciphertext, true)
			if err != nil {
				t.Fatal(err)
			}

			for i, dec := range decs {
				plaintext, err := dec.Plaintext()
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(plaintext, plaintexts[i]) {
					t.Fatal("decrypted plaintext is not equal to initial plaintext")
				}
			}
		})
	}
}

func TestEncryptProvableDecryptManyData(t *testing.T) {
	salt := []byte(":p")
	plaintext1 := []byte("Hello World!")
	plaintext2 := []byte("I'm Alice")
	plaintexts := [][]byte{plaintext1, plaintext2}

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			parameter := NewParameters(g)
			rnd, err := group.RandomScalar(parameter.Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			var key crypto.DecryptionKey

			key, err = NewPrivateKey(parameter, rnd)
			if err != nil {
				t.Fatal(err)
			}

			rnd, err = group.RandomScalar(key.Parameters().Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			ciphertext, err := util.EncryptAll(rnd, key.EncryptionKey(), nil, plaintexts...)
			if err != nil {
				t.Fatal(err)
			}

			rnd2, err := group.RandomScalar(key.Parameters().Group().Order())
			if err != nil {
				t.Fatal(err)
			}

			_, decs, proofs, err := util.ProvableDecryptAll(rnd2, key, nil, ciphertext, salt, true)
			if err != nil {
				t.Fatal(err)
			}

			for i, dec := range decs {
				plaintext, err := dec.Plaintext()
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(plaintext, plaintexts[i]) {
					t.Fatal("decrypted ElGamal plaintext doesn't match initial one")
				}

				err = key.EncryptionKey().Verify(proofs[i], salt)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
