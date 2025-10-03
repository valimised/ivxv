package distributed

import (
	"bytes"
	"testing"
	"tivi.io/core/crypto"
	"tivi.io/core/crypto/util"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

func TestDistributed(t *testing.T) {
	threshold, parties := uint64(3), uint64(5)
	testMessage := []byte("test-message")
	testExtra := []byte("test-extra")

	for _, g := range group.All() {
		t.Run(string(g.Name()), func(t *testing.T) {
			// User should provide:
			// - threshold
			// - parties
			// - group name

			// User generates private key shares and EncryptionKey key
			privKeyShares, pkey, err := NewPrivateKeyShares(g, parties, threshold)
			if err != nil {
				t.Fatal(err)
			}
			// Encrypt msg
			rnd, err := group.RandomScalar(pkey.Parameters().Group().Order())
			ciphertext, err := pkey.Encrypt(rnd, testMessage)
			if err != nil {
				t.Fatal(err)
			}

			// DecryptAll ciphertext by parts
			decshares := make([]crypto.Decryption, parties)

			// DecryptAll ciphertext by parts and with proof for each part
			decshares2 := make([]crypto.Decryption, parties)

			decshares3 := make([][]byte, parties)
			decshares4 := make([][]byte, parties)
			proofs := make([][]byte, parties)

			for i, share := range privKeyShares {
				// Simulate simple decrypted
				decshares[i], err = share.Decrypt(ciphertext, true)
				if err != nil {
					t.Fatal(err)
				}

				// Simulate decrypted with proofs
				rnd, err := group.RandomScalar(pkey.Parameters().Group().Order())
				decshares2[i], proofs[i], err = share.ProvableDecrypt(rnd, ciphertext, testExtra, true)
				if err != nil {
					t.Fatal(err)
				}

				decshares4[i], err = decshares[i].Marshal()
				if err != nil {
					t.Fatal(err)
				}

				decshares3[i], err = decshares2[i].Marshal()
				if err != nil {
					t.Fatal(err)
				}
			}

			decryptionShareCombiner := NewDecryptionShareCombiner(g, threshold)

			combinedDecrypted, err := decryptionShareCombiner.DecryptionSharesCombine(ciphertext, decshares4...)
			if err != nil {
				t.Fatal(err)
			}

			combinedDecrypted2, err := decryptionShareCombiner.DecryptionSharesCombine(ciphertext, decshares3...)
			if err != nil {
				t.Fatal(err)
			}

			// Plaintext after DecryptAll
			decmsg, err := combinedDecrypted.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			// Plaintext after ProvableDecryptAll
			decmsg2, err := combinedDecrypted2.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			// Check that plaintexts are equal to initial msg
			if !bytes.Equal(decmsg, testMessage) {
				t.Fatalf("%x %x\n", decmsg, testMessage)
			}
			if !bytes.Equal(decmsg2, testMessage) {
				t.Fatalf("%x %x\n", decmsg2, testMessage)
			}

			// VerifyAll proofs from ProvableDecryptAll by public key shares.
			// With extra bytes
			for i, decshare := range proofs {
				err = privKeyShares[i].pkey.Verify(decshare, testExtra)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Without extra bytes, should fail, because proof challenge uses
			// testExtra in a hash creation, and since we used testExtra in
			// ProvableDecryptAll, we should also use testExtra in verify.
			//
			// However, if you use nil for extra in ProvableDecryptAll, then
			// VerifyAll extra should also be nil
			for i, decshare := range proofs {
				err = privKeyShares[i].pkey.Verify(decshare, nil)
				if err == nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestShareRegeneration(t *testing.T) {
	threshold, parties := uint64(3), uint64(5)
	for _, g := range group.All() {
		t.Run(string(g.Name()), func(t *testing.T) {
			// User generates private key shares and EncryptionKey key
			privKeyShares, _, err := NewPrivateKeyShares(g, parties, threshold)
			if err != nil {
				t.Fatal(err)
			}

			privSH := make([]crypto.EncryptionPrivateKeyShare, len(privKeyShares))
			for i, pp := range privKeyShares {
				privSH[i] = pp
			}

			index := parties - 1

			// Suppose we only know an Index of a share, but not share itself
			shareIndex := privKeyShares[index].params.Index()
			//
			//// Simulate storing private key shares in I/O
			//privKeySharesBytes := make([][]byte, len(privKeyShares))
			//for i, privKeyShare := range privKeyShares {
			//	privKeySharesBytes[i], err = privKeyShare.Marshal()
			//	if err != nil {
			//		t.Fatal(err)
			//	}
			//}

			// Then we wish to restore that private key share
			// Given only Index of private key share of interest and
			// #threshold amount of private key shares
			regenerator := NewRegenerator(g, threshold)
			var regeneratorI crypto.EncryptionPrivateKeyShareRegenerator = regenerator
			restoredSinglePrivKeyShare, err := regeneratorI.RegeneratePrivateKeyShare(shareIndex, privSH)
			if err != nil {
				t.Fatal(err)
			}

			// ASN1 marshal just to compare by bytes
			restoredSinglePrivKeyShareBytes, err := restoredSinglePrivKeyShare.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			// ASN1 marshal just to compare by bytes
			privKeySharesRestoredBytes, err := privKeyShares[index].Marshal()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(restoredSinglePrivKeyShareBytes, privKeySharesRestoredBytes) {
				t.Fatal(err)
			}
		})
	}
}

func TestEncryptDecryptManyData(t *testing.T) {
	data1 := []byte("Hello World!")
	data2 := []byte("I'm Alice")
	data3 := []byte("Good weather")
	data4 := []byte("Happy coding :)")
	dataSet := [][]byte{data1, data2, data3, data4}

	for _, g := range group.All() {
		t.Run(string(g.Name()), func(t *testing.T) {
			var parties uint64 = 5
			var threshold uint64 = 3

			// Create new private key
			var pubKeyI crypto.EncryptionKey
			privKeyShares, pubKeyI, err := NewPrivateKeyShares(g, parties, threshold)
			if err != nil {
				t.Fatal(err)
			}

			// All private key shares should have exactly same algorithm
			privKeyShare0Algo := privKeyShares[0].Parameters().Algorithm()

			// Encrypt all ciphertexts, so at output we get ciphertext={ct1,ct2}
			rnd, err := group.RandomScalar(pubKeyI.Parameters().Group().Order())
			ciphertext, err := util.EncryptAll(rnd, pubKeyI, privKeyShare0Algo, dataSet...)
			if err != nil {
				t.Fatal(err)
			}

			decryptionSharesPerCiphertext := make([][]crypto.Decryption, threshold)
			decryptionSharesPerCiphertextBytes := make([][][]byte, threshold)
			cts := make([][]byte, len(dataSet))
			// Each private key share should decrypt a ciphertext, i.e.
			// since ciphertext={c1,c2}, so we got decrypted={dec1,dec2}
			for i, privKeyShare := range privKeyShares[:threshold] {
				// DecryptAll ciphertext -> {dec1,dec2}
				cts, decryptionSharesPerCiphertext[i], err = util.DecryptAll(privKeyShare, privKeyShare.Parameters().Algorithm(), ciphertext, true)
				if err != nil {
					t.Fatal(err)
				}

				decryptionSharesBytes := make([][]byte, len(dataSet))

				// Range over all decrypted ({dec1,dec2}) and ASN1 marshal each
				for j, decryptionShare := range decryptionSharesPerCiphertext[i] {
					decryptionSharesBytes[j], err = decryptionShare.Marshal()
					if err != nil {
						t.Fatal(err)
					}
				}

				decryptionSharesPerCiphertextBytes[i] = make([][]byte, len(dataSet))
				// Store ASN1 marshalled decrypted ciphertext pair {dec1B,dec2B}
				decryptionSharesPerCiphertextBytes[i] = decryptionSharesBytes
			}

			// Time to combine decrypted shares
			//
			// We combine it as follows, suppose we have ciphertexts={ct1,ct2}
			// and 3 private key shares then combining is done as follows:
			// plain1={dec1,dec11,dec111}, plain2={dec2,dec22,dec222}
			decryptionShareCombiner := NewDecryptionShareCombiner(g, threshold)

			for i := range dataSet {
				decryptionSharesBytesPerSpecificCiphertext := make([][]byte, threshold)
				for j := range decryptionSharesPerCiphertextBytes {
					decryptionSharesBytesPerSpecificCiphertext[j] = append(decryptionSharesBytesPerSpecificCiphertext[j], decryptionSharesPerCiphertextBytes[j][i]...)
				}

				// Combine decryption shares for ciphertext ct[i]
				combinedDecrypted, err := decryptionShareCombiner.DecryptionSharesCombine(cts[i], decryptionSharesBytesPerSpecificCiphertext...)
				if err != nil {
					t.Fatal(err)
				}

				plaintext, err := combinedDecrypted.Plaintext()
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(plaintext, dataSet[i]) {
					t.Fatal("decrypted not equal to initial data")
				}
			}
		})
	}
}

func TestEncryptProvableDecryptManyData(t *testing.T) {
	extra := []byte("Reji")
	data1 := []byte("Hello World!")
	data2 := []byte("I'm Alice")
	dataSet := [][]byte{data1, data2}

	for _, g := range group.All() {
		t.Run(string(g.Name()), func(t *testing.T) {
			var parties uint64 = 3
			var threshold uint64 = 2

			// Create new private key
			var pubKeyI crypto.EncryptionKey
			privKeyShares, pubKeyI, err := NewPrivateKeyShares(g, parties, threshold)
			if err != nil {
				t.Fatal(err)
			}

			// All private key shares should have exactly same algorithm
			privKeyShare0Algo := privKeyShares[0].Parameters().Algorithm()

			// Encrypt all ciphertexts, so at output we get ciphertext={ct1,ct2}
			rnd, err := group.RandomScalar(pubKeyI.Parameters().Group().Order())
			ciphertext, err := util.EncryptAll(rnd, pubKeyI, privKeyShare0Algo, dataSet...)
			if err != nil {
				t.Fatal(err)
			}

			decryptionSharesProofsPerCiphertext := make([][][]byte, threshold)
			decryptionSharesPerCiphertext := make([][]crypto.Decryption, threshold)
			decryptionSharesPerCiphertextBytes := make([][][]byte, threshold)
			cts := make([][]byte, len(dataSet))

			// Each private key share should decrypt a ciphertext, i.e.
			// since ciphertext={c1,c2}, so we got decrypted={dec1,dec2}
			for i, privKeyShare := range privKeyShares[:threshold] {
				// DecryptAll ciphertext -> {dec1,dec2}
				rnd, err := group.RandomScalar(pubKeyI.Parameters().Group().Order())
				cts, decryptionSharesPerCiphertext[i], decryptionSharesProofsPerCiphertext[i], err = util.ProvableDecryptAll(rnd, privKeyShare, privKeyShare.Parameters().Algorithm(), ciphertext, extra, true)
				if err != nil {
					t.Fatal(err)
				}

				decryptionSharesBytes := make([][]byte, len(dataSet))

				// Range over all decrypted ({dec1,dec2}) and ASN1 marshal each
				for j, decryptionShare := range decryptionSharesPerCiphertext[i] {
					decryptionSharesBytes[j], err = decryptionShare.Marshal()
					if err != nil {
						t.Fatal(err)
					}
				}

				decryptionSharesPerCiphertextBytes[i] = make([][]byte, len(dataSet))
				// Store ASN1 marshalled decrypted ciphertext pair {dec1B,dec2B}
				decryptionSharesPerCiphertextBytes[i] = decryptionSharesBytes
			}

			// Time to combine decrypted shares
			//
			// We combine it as follows, suppose we have ciphertexts={ct1,ct2}
			// and 3 private key shares then combining is done as follows:
			// plain1={dec1,dec11,dec111}, plain2={dec2,dec22,dec222}
			decryptionShareCombiner := NewDecryptionShareCombiner(g, threshold)

			for i := range dataSet {
				decryptionSharesBytesPerSpecificCiphertext := make([][]byte, threshold)
				for j := range decryptionSharesPerCiphertextBytes {
					err = privKeyShares[j].EncryptionKey().Verify(decryptionSharesProofsPerCiphertext[j][i], extra)
					if err != nil {
						t.Fatal(err)
					}
					decryptionSharesBytesPerSpecificCiphertext[j] = append(decryptionSharesBytesPerSpecificCiphertext[j], decryptionSharesPerCiphertextBytes[j][i]...)
				}

				// Combine decryption shares for ciphertext ct[i]
				combinedDecrypted, err := decryptionShareCombiner.DecryptionSharesCombine(cts[i], decryptionSharesBytesPerSpecificCiphertext...)
				if err != nil {
					t.Fatal(err)
				}

				plaintext, err := combinedDecrypted.Plaintext()
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(plaintext, dataSet[i]) {
					t.Fatal("decrypted not equal to initial data")
				}
			}
		})
	}
}
