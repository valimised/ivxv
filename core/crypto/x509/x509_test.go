package x509

import (
	"bytes"
	"sync"
	"testing"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	"tivi.io/core/crypto/elgamal/distributed"
	x509d "tivi.io/core/crypto/elgamal/distributed/x509"
	"tivi.io/core/crypto/elgamal/homomorphic"
	x509e "tivi.io/core/crypto/elgamal/x509"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

var (
	x509DefaultUnmarshaller = NewUnmarshaller[crypto.EncryptionKey](x509e.NewUnmarshaller())
	x509DefaultMarshaller   = NewMarshaller(x509e.NewMarshaller())

	pkcs8DefaultUnmarshaller = NewUnmarshallerPKCS8[crypto.DecryptionKey](x509e.NewUnmarshallerPKCS8())
	pkcs8DefaultMarshaller   = NewMarshallerPKCS8(x509e.NewMarshallerPKCS8())

	x509DistributedUnmarshaller = NewUnmarshaller[crypto.EncryptionKey](x509d.NewUnmarshaller())
	x509DistributedMarshaller   = NewMarshaller(x509e.NewMarshaller())

	pkcs8DistributedUnmarshaller = NewUnmarshallerPKCS8[crypto.DecryptionKey](x509d.NewUnmarshallerPKCS8())
	pkcs8DistributedMarshaller   = NewMarshallerPKCS8(x509e.NewMarshallerPKCS8())
)

func storePEMPKCS8PrivateKey(key crypto.KeyInfo) ([]byte, error) {
	return pkcs8DefaultMarshaller.Marshal(key)
}
func storePEMPKCS8PrivateKeyShare(key crypto.KeyInfo) ([]byte, error) {
	return pkcs8DistributedMarshaller.Marshal(key)
}
func storeDERPKCS8PrivateKey(key crypto.KeyInfo) ([]byte, error) {
	return pkcs8DefaultMarshaller.MarshalDER(key)
}
func storeDERPKCS8PrivateKeyShare(key crypto.KeyInfo) ([]byte, error) {
	return pkcs8DistributedMarshaller.MarshalDER(key)
}
func storePEMX509PublicKey(key crypto.KeyInfo) ([]byte, error) {
	return x509DefaultMarshaller.Marshal(key)
}

func storePEMX509PublicKeyShare(key crypto.KeyInfo) ([]byte, error) {
	return x509DistributedMarshaller.Marshal(key)
}

func storeDERX509PublicKey(key crypto.KeyInfo) ([]byte, error) {
	return x509DefaultMarshaller.MarshalDER(key)
}

func storeDERX509PublicKeyShare(key crypto.KeyInfo) ([]byte, error) {
	return x509DistributedMarshaller.MarshalDER(key)
}

// encrypt a plaintext.
func encrypt(encrypter crypto.Encrypter, plaintext []byte, rand *group.Scalar) ([]byte, error) {
	return encrypter.Encrypt(rand, plaintext)
}

// decrypt a ciphertext.
func decrypt(decrypter crypto.Decrypter, ciphertext []byte) (crypto.Decryption, error) {
	return decrypter.Decrypt(ciphertext)
}

func sign(signer crypto.Signer, data []byte) ([]byte, error) {
	rnd, err := group.RandomScalar(signer.Parameters().Group().Order())
	if err != nil {
		return nil, err
	}
	return signer.Sign(rnd, data)
}

func verifySignature(verifier crypto.SignatureVerifier, signature, data []byte) error {
	rnd, err := group.RandomScalar(verifier.Parameters().Group().Order())
	if err != nil {
		return err
	}
	return verifier.Verify(rnd, signature, data)
}

// provableDecrypt a ciphertext with a proof of a correct decryption provided.
func provableDecrypt(rand *group.Scalar, provableDecrypter crypto.ProvableDecrypter, ciphertext, extra []byte) (crypto.Decryption, []byte, error) {
	return provableDecrypter.ProvableDecrypt(rand, ciphertext, extra)
}

// verify decryption proof.
func verify(verifier crypto.EncryptionVerifier, proof, extra []byte) error {
	return verifier.Verify(proof, extra)
}

// encryptHomomorphic ciphertext (marks) homomorphically.
func encryptHomomorphic(encrypter crypto.HomomorphicEncrypter, pubkey crypto.EncryptionKey, marks []bool) ([]byte, error) {
	return encrypter.Encrypt(pubkey, marks)
}

// aggregateHomomorphic sums up all choices into a single choice.
func aggregateHomomorphic(aggregator crypto.HomomorphicAggregator, pk crypto.EncryptionKey, batchSize int, choices ...[]byte) ([]byte, error) {
	// Aggregate choices in parallel
	wg := new(sync.WaitGroup)
	lock := new(sync.Mutex)
	var aggregated []byte

	// Use homomorphic property of each choice to aggregate them together
	for i := 0; i < len(choices); i += batchSize {
		end := i + batchSize
		if end > len(choices) {
			end = len(choices)
		}
		batch := choices[i:end]

		// Each goroutine call will be recorded. Using wg.Wait() will be awaited
		wg.Add(1)

		// Use goroutines to process each batch in parallel
		go func() {
			// Aggregate batches (use as large batch as you can), when we aggregate
			// raw batch, we also wish to verify that each choice is correct
			aggregatedBatch, err := aggregator.Aggregate(pk, true, batch...)
			if err != nil {
				// Log error and return
				panic(err) // Don't ever panic in production!
			}

			// Aggregation of a final aggregated result should be done in a
			// synchronized manner, i.e. we, indeed, can produce independent
			// batches, but summing them all up into a single variable is a
			// synchronized process.
			go func() {
				lock.Lock()
				defer lock.Unlock()

				// Aggregate aggregated batches, withVerify is false, because
				// range proof verification therefore fails (we have already
				// aggregated choices here)
				aggregated = append(aggregated, aggregatedBatch...)
				aggregated, err = aggregator.Aggregate(pk, false, aggregated)
				if err != nil {
					// Log error and return
					panic(err) // Don't ever panic in production!
				}

				// Mark goroutine as finished (indicates that batch has been
				// aggregated and verified)
				wg.Done()
			}()
		}()
	}

	// Block and wait for all unfinished goroutines (code didn't reach
	// wg.Done() statement)
	wg.Wait()

	return aggregated, nil
}

// tallyHomomorphic decrypts aggregated choice and presents voting results.
func tallyHomomorphic(tallier crypto.HomomorphicTallier, privkey []crypto.DecryptionKey, pubkey []crypto.EncryptionKey, aggregated, extra []byte, maxCount int) ([]crypto.HomomorphicDecryption, error) {
	return tallier.Tally(privkey, pubkey, aggregated, maxCount, extra)
}

// verifyHomomorphic verifies decrypted choice proof.
func verifyHomomorphic(verifier crypto.HomomorphicVerifier, pubkey crypto.EncryptionKey, proof, extra []byte) error {
	return verifier.Verify(pubkey, proof, extra)
}

// TestEncryptionPrivateKey simulates ElGamal private key encryption from a
// perspective of a caller.
func TestEncryptionPrivateKey(t *testing.T) {
	for _, registeredGroup := range group.All() {
		t.Run(string(registeredGroup.Name()), func(t *testing.T) {
			// To generate encryption private key, user should provide:
			//	a) Group name
			// With provided group name user can obtain a registered Group
			// implementation
			name := string(registeredGroup.Name())
			g, err := group.Get(name)
			if err != nil {
				t.Fatal(err)
			}

			// Now user decides, which encryption scheme to use:
			//	a) ElGamal
			//	b) Paillier
			//	c) RSA
			// Since TIVI Core only implements ElGamal - it will be our choice
			r, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			params := elgamal.NewParameters(g)
			privKey, err := elgamal.NewPrivateKey(params, r)
			if err != nil {
				t.Fatal(err)
			}

			// ElGamal keys have been generated, but before we pass it
			// around, we should cast them to the interface
			var privKeyI crypto.DecryptionKey = privKey
			pubKeyI := privKeyI.EncryptionKey()

			// Simulate storing keys in some other place in I/O, this is
			// exactly what we do, when keys are generated
			privKeyBytes, err := storePEMPKCS8PrivateKey(privKeyI)
			if err != nil {
				t.Fatal(err)
			}

			pubKeyBytes, err := storePEMX509PublicKey(pubKeyI)
			if err != nil {
				t.Fatal(err)
			}

			// Now it is time to encrypt some data, for that, we need
			// public key, which is currently X509 bytes.
			//
			// Note, that X509 unmarshalling doesn't care about public key
			// implementation, even scheme... It is just X509 bytes unmarshalling
			_, pubKeyRestoredI, err := x509DefaultUnmarshaller.Unmarshal(pubKeyBytes) //(pubKeyBytes)
			if err != nil {
				t.Fatal(err)
			}

			// Now, let's encrypt some data. Note that we pass around public
			// key interface, not the implementation
			plaintext := []byte("Hello World!")
			rand, err := group.RandomScalar(pubKeyRestoredI.Parameters().Group().Order())
			ciphertext, err := encrypt(pubKeyRestoredI, plaintext, rand)
			if err != nil {
				t.Fatal(err)
			}

			// We have our plaintext encrypted for now, next step is to decrypt
			// our ciphertext to prove that decryption matches the initial
			// plaintext
			//
			// But before we do it, obviously we need a private key, let's get it
			// from PKCS8, note again, that private key is an interface
			_, privKeyRestoredI, err := pkcs8DefaultUnmarshaller.Unmarshal(privKeyBytes)
			if err != nil {
				t.Fatal(err)
			}

			// Note, again, we pass interface only. Let's decrypt a ciphertext
			decrypted, err := decrypt(privKeyRestoredI, ciphertext)
			if err != nil {
				t.Fatal(err)
			}

			// Let's check that initial plaintext matches decrypted one
			plaintext2, err := decrypted.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			// Compare plaintexts
			if !bytes.Equal(plaintext, plaintext2) {
				t.Fatal("Initial plaintext != decrypted plaintext")
			}

			// Let's do the decryption, but now with proofs (to verify them
			// later using public key, i.e. auditing)
			//
			// We will use extra bytes, in order to add kind of identifier to the
			// decrypted proof (extra bytes are used in HASH creation for a proof),
			// we also call it "to add a salt into a hash".
			//
			// Note again, only private key interface is passed in
			extra := []byte("Just some extra bytes")
			rand, err = group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}
			decrypted2, proof, err := provableDecrypt(rand, privKeyRestoredI, ciphertext, extra)
			if err != nil {
				t.Fatal(err)
			}

			// Let's do quick check that plaintexts match
			plaintext3, err := decrypted2.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			// Compare plaintexts again
			if !bytes.Equal(plaintext, plaintext3) {
				t.Fatal("Initial plaintext != decrypted with proof")
			}

			// Note that in ProvableDecryptAll we used extra bytes - that means
			// that for the Verify, same extra bytes should be used as well
			err = verify(pubKeyRestoredI, proof, extra)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// / TestDistributedEncryptionPrivateKey simulates distributed ElGamal private key
// encryption from a perspective of a caller.
func TestDistributedEncryptionPrivateKey(t *testing.T) {
	for _, registeredGroup := range group.All() {
		t.Run(string(registeredGroup.Name()), func(t *testing.T) {
			// To generate distributed encryption private key, user should provide:
			//	a) Group name
			//	b) threshold amount
			//	c) Parties amount
			var parties uint64 = 5
			var threshold uint64 = 3

			// With provided group name user can obtain a registered Group
			// implementation
			name := string(registeredGroup.Name())
			g, err := group.Get(name)
			if err != nil {
				t.Fatal(err)
			}

			// Now user decides, which distributed encryption scheme to use:
			//	a) ElGamal
			//	b) Paillier
			//	c) RSA
			// Since TIVI Core only implements distributed ElGamal - it will be our choice
			privKeyShares, pubKey, err := distributed.NewPrivateKeyShares(g, parties, threshold)
			if err != nil {
				t.Fatal(err)
			}

			// ElGamal key shares has been generated, but before we pass it
			// around, we should cast it to the interface
			privKeySharesI := make([]crypto.DecryptionKey, len(privKeyShares))
			for i, privKeyShare := range privKeyShares {
				privKeySharesI[i] = privKeyShare
			}

			// Simulate storing key shares in some other place in I/O, this is
			// exactly what we do, when key shares are generated
			privKeySharesBytes := make([][]byte, len(privKeySharesI))
			pubKeySharesBytes := make([][]byte, len(privKeySharesI))
			var privKeyShareI crypto.DecryptionKey
			var pubKeyShareI crypto.EncryptionKey

			for i, privKeyShare := range privKeyShares {
				// Only interface is passed in
				privKeyShareI = privKeyShare
				privKeySharesBytes[i], err = storePEMPKCS8PrivateKeyShare(privKeyShareI)
				if err != nil {
					t.Fatal(err)
				}

				// Only interface is passed in
				pubKeyShareI = privKeyShare.EncryptionKey()
				pubKeySharesBytes[i], err = storeDERX509PublicKeyShare(pubKeyShareI)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Interface
			var pubKeyI crypto.EncryptionKey = pubKey
			pubKeyBytes, err := storePEMX509PublicKey(pubKeyI)
			if err != nil {
				t.Fatal(err)
			}

			// Now it is time to encrypt some data, for that, we need distributed
			// public key, which is currently X509 bytes.
			//
			// Note, that X509 unmarshalling doesn't care about distributed public
			// key implementation, even scheme... It is just X509 bytes unmarshalling
			_, pubKeyRestoredI, err := x509DefaultUnmarshaller.Unmarshal(pubKeyBytes)
			if err != nil {
				t.Fatal(err)
			}

			// Now, let's encrypt some data
			plaintext := []byte("Hello World!")
			rand, err := group.RandomScalar(pubKeyRestoredI.Parameters().Group().Order())
			ciphertext, err := encrypt(pubKeyRestoredI, plaintext, rand)
			if err != nil {
				t.Fatal(err)
			}

			// We have our plaintext encrypted for now, next step is to decrypt
			// our ciphertext to prove that decryption matches the initial
			// plaintext
			//
			// But before we do it, obviously we need private key shares,
			// let's get them from PKCS8
			privKeySharesRestoredI := make([]crypto.DecryptionKey, threshold)
			for i, privKeyShareBytes := range privKeySharesBytes[:threshold] {
				_, privKeySharesRestoredI[i], err = pkcs8DistributedUnmarshaller.Unmarshal(privKeyShareBytes)
				if err != nil {
					t.Fatal(err)
				}
			}

			// NB! Please note, that during decryption/verification period
			// we only use exactly threshold amount of key shares (we can use
			// more shares than a threshold, but not less)

			// Distributed encryption will decrypt ciphertext by parts (part == share),
			// each part is interface as well
			decryptedI := make([]crypto.Decryption, len(privKeySharesRestoredI))

			// decryption parts of a ciphertext, before combined, should be
			// ASN1 marshalled, in order to allow many implementations (regular
			// or distributed) to format decrypted parts as they want
			decryptedBytes := make([][]byte, len(privKeySharesRestoredI))
			for i, privKeyShareRestoredI := range privKeySharesRestoredI {
				decryptedI[i], err = decrypt(privKeyShareRestoredI, ciphertext)
				if err != nil {
					t.Fatal(err)
				}

				decryptedBytes[i], err = decryptedI[i].Marshal()
				if err != nil {
					t.Fatal(err)
				}
			}

			// Distributed ElGamal concrete type to create a decryption shares
			// combiner
			decryptedCombiner := distributed.NewDecryptionShareCombiner(g, threshold)
			if err != nil {
				t.Fatal(err)
			}
			// Cast to interface
			var decryptedCombinerI crypto.DecryptionSharesCombiner = decryptedCombiner

			// decryption parts are combined to present a meaningful decrypted value.
			// This value is of interface type as well
			decryptedCombinedI, err := decryptedCombinerI.DecryptionSharesCombine(ciphertext, decryptedBytes...)
			if err != nil {
				t.Fatal(err)
			}

			// Let's check that initial plaintext matches decrypted one
			plaintext2, err := decryptedCombinedI.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			// Compare plaintexts
			if !bytes.Equal(plaintext, plaintext2) {
				t.Fatal("Initial plaintext != decrypted plaintext")
			}

			// Let's do the decryption, but now with proofs (to verify them later,
			// i.e. auditing)
			//
			// We will use extra bytes, in order to add kind of identifier to the
			// decrypted proof (extra bytes are used in HASH creation for a proof)
			extra := []byte("Just some extra bytes")
			proofsBytes := make([][]byte, len(privKeySharesRestoredI))
			decrypted2I := make([]crypto.Decryption, len(privKeySharesRestoredI))
			decrypted2Bytes := make([][]byte, len(privKeySharesRestoredI))
			for i, privKeyShareRestoredI := range privKeySharesRestoredI {
				rr, err := group.RandomScalar(g.Order())
				decrypted2I[i], proofsBytes[i], err = provableDecrypt(rr, privKeyShareRestoredI, ciphertext, extra)
				if err != nil {
					t.Fatal(err)
				}

				decrypted2Bytes[i], err = decrypted2I[i].Marshal()
				if err != nil {
					t.Fatal(err)
				}
			}

			decrypted2CombinedI, err := decryptedCombinerI.DecryptionSharesCombine(ciphertext, decrypted2Bytes...)
			if err != nil {
				t.Fatal(err)
			}

			// Let's do quick check that plaintexts matches
			plaintext3, err := decrypted2CombinedI.Plaintext()
			if err != nil {
				t.Fatal(err)
			}

			// Compare plaintexts again
			if !bytes.Equal(plaintext, plaintext3) {
				t.Fatal("Initial plaintext != decrypted with proof")
			}

			// Now, since we don't want to verify proofs straight away, and
			// instead we store them in I/O for later auditing (that is how we
			// actually do in production, since auditors process info after
			// the decryption is done)

			// Simulating proof storage in I/O (here we don't do anything, since
			// proof is already in form of []byte, so we could just os.Write())

			// At this point, all voting results are known and auditor wish
			// to process proof, and for that reason he needs a public key shares,
			// to verify each decrypted part's proof
			pubKeySharesRestoredI := make([]crypto.EncryptionKey, threshold)
			for i, pubKeyShareBytes := range pubKeySharesBytes[:threshold] {
				_, pubKeySharesRestoredI[i], err = x509DistributedUnmarshaller.Unmarshal(pubKeyShareBytes)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Note that during ProvableDecryptAll we used extra bytes - that means
			// that for the VerifyAll, same extra bytes should be used as well
			for i, proofBytes := range proofsBytes {
				err = verify(pubKeySharesRestoredI[i], proofBytes, extra)
				if err != nil {
					t.Fatal(err)
				}
			}

			// That's it! All main e2e workflow is done!
		})
	}
}

// TestHomomorphicEncryptionPrivateKey simulates homomorphic ElGamal private key
// encryption from a perspective of a caller.
func TestHomomorphicEncryptionPrivateKey(t *testing.T) {
	for _, registeredGroup := range group.All() {
		t.Run(string(registeredGroup.Name()), func(t *testing.T) {
			// Even though we work with homomorphic encryption, first of all,
			// we still have to generate our encryption private key.

			// To generate encryption private key, user should provide:
			//	a) Group name
			// With provided group name user can obtain a registered Group
			// implementation
			name := string(registeredGroup.Name())
			g, err := group.Get(name)
			if err != nil {
				t.Fatal(err)
			}

			// Now user decides, which encryption scheme to use:
			//	a) ElGamal
			//	b) Paillier
			//	c) RSA
			// Since TIVI Core only implements ElGamal - it will be our choice
			r, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			params := elgamal.NewParameters(g)
			privKey, err := elgamal.NewPrivateKey(params, r)
			if err != nil {
				t.Fatal(err)
			}

			// ElGamal private key has been generated, and before we pass it
			// around, we should cast it to the interface
			var privKeyI crypto.DecryptionKey = privKey
			privKeyBytes, err := storeDERPKCS8PrivateKey(privKeyI)
			if err != nil {
				t.Fatal(err)
			}
			pubKeyBytes, err := storePEMX509PublicKey(privKeyI.EncryptionKey())
			if err != nil {
				t.Fatal(err)
			}

			// Now it is time to encrypt some data, for that, we need ElGamal
			// public key, which is currently X509 bytes.
			//
			// Note, that X509 unmarshalling doesn't care about public key
			// implementation, even scheme... It is just X509 bytes unmarshalling
			_, pubKeyRestoredI, err := x509DefaultUnmarshaller.Unmarshal(pubKeyBytes)
			if err != nil {
				t.Fatal(err)
			}

			// Now, let's encrypt some data, suppose we have PREDEFINED candidate
			// list as {John, Milana, Mike}, user has made his choice. NB! We are
			// only allowed to choose one candidate. So this time, we have chosen
			// Milana.
			voterChoice := []bool{false, true, false}

			// Homomorphic ElGamal concrete implementation
			homomorphicEncryption := homomorphic.NewElGamalHomomorphicEncryption(g, 0)
			// Homomorphic encryption interface, which suits both for distributed
			// and regular encryption schemes
			var homomorphicEncrypter crypto.HomomorphicEncrypter = homomorphicEncryption

			// Encrypt homomorphically. In terms of homomorphic encryption
			// ciphertext == choice
			choice, err := encryptHomomorphic(homomorphicEncrypter, pubKeyRestoredI, voterChoice)
			if err != nil {
				t.Fatal(err)
			}

			// We have our choice encrypted for now, next step is to decrypt
			// our choice to prove that after decryption matches the initial
			// choice
			//
			// But before we do it, obviously we need a private key, let's get it
			// from PKCS8
			_, privKeyRestoredI, err := pkcs8DefaultUnmarshaller.Unmarshal(privKeyBytes)
			if err != nil {
				t.Fatal(err)
			}

			// Homomorphic encryption has an amazing property - we can sum up
			// all choices made by all voters without decrypting each choice,
			// and therefore as a result of aggregation we will get a single
			// choice. But please note, that this choice is still encrypted,
			// BUT aggregated. To decrypt a final choice we use tally.
			var homomorphicAggregator crypto.HomomorphicAggregator = homomorphicEncryption
			aggregated, err := aggregateHomomorphic(homomorphicAggregator, pubKeyRestoredI, 1000, choice)
			if err != nil {
				t.Fatal(err)
			}

			// All voter choices are summed up into a single choice, so in
			// out example that single choice will have exactly 3 candidates
			// ({John, Milana, Mike}) and the sum of all votes for them. To
			// obtain decrypted voting results we use tally
			var homomorphicTallier crypto.HomomorphicTallier = homomorphicEncryption
			tallies, err := tallyHomomorphic(homomorphicTallier, []crypto.DecryptionKey{privKeyRestoredI}, []crypto.EncryptionKey{pubKeyRestoredI}, aggregated, nil, 3)
			if err != nil {
				t.Fatal(err)
			}

			// marks = {false, true, false} --> {-, Milana, -}
			marksInt := []uint64{0, 1, 0}
			for i, count := range marksInt {
				if tallies[i].Count() != count {
					t.Fatalf("Mark nr.%d count expected %d, got %d", i, count, tallies[i].Count())
				}
			}

			// In our example, a choice has 3 marks exactly. So, when we
			// tallied results each choice will have 3 decryption proofs (one
			// per mark). And if we have distributed encryption, then a ciphertext
			// is partially decrypted by amount of private key shares, so each choice
			// will have (threshold * len(marks)).
			var homomorphicVerifier crypto.HomomorphicVerifier = homomorphicEncryption
			// Each tally is a decrypted mark with a proof (regular encryption)
			// or proofs (distributed encryption, i.e. each private share decrypts
			// a mark independently)
			//
			// Note, that how we can make this verification all in parallel,
			// just the same way we did with aggregation
			wg := new(sync.WaitGroup)
			wg2 := new(sync.WaitGroup)
			for _, tally := range tallies {
				wg.Add(1)

				go func(tally crypto.HomomorphicDecryption) {
					tallyProofs, err := tally.Proofs()
					if err != nil {
						wg.Done()
						// Log error and return from that particular goroutine
						panic(err) // Don't ever panic in production like that
					}

					// For regular encryption there will only be a single proof, but
					// for distributed there will be threshold amount of proofs, so
					// we verify all of them
					for _, proof := range tallyProofs {
						wg2.Add(1)
						go func(proof []byte) {
							err = verifyHomomorphic(homomorphicVerifier, pubKeyRestoredI, proof, nil)
							if err != nil {
								wg.Done()
								wg2.Done()
								// Log error and return from that particular goroutine
								panic(err) // Don't ever panic in production like that
							}
							wg2.Done()
						}(proof)
					}
					wg.Done()
				}(tally)
			}

			// Wait all goroutines that verify proofs
			wg2.Wait()
			// Wait all goroutines that spread proofs for verification to wg2
			wg.Wait()

			// That's it! All main e2e workflow is done!
		})
	}
}

// TestDistributedHomomorphicEncryptionPrivateKey simulates homomorphic distributed
// ElGamal private key encryption from a perspective of a caller.
func TestDistributedHomomorphicEncryptionPrivateKey(t *testing.T) {
	for _, registeredGroup := range group.All() {
		t.Run(string(registeredGroup.Name()), func(t *testing.T) {
			// Even though we work with homomorphic encryption, first of all,
			// we still have to generate our distributed encryption private key.

			// To generate distributed encryption private key, user should provide:
			//	a) Group name
			//	b) threshold amount
			//	c) Parties amount
			// With provided group name user can obtain a registered Group
			// implementation
			name := string(registeredGroup.Name())
			g, err := group.Get(name)
			if err != nil {
				t.Fatal(err)
			}
			var threshold uint64 = 3
			var parties uint64 = 3

			// Now user decides, which distributed encryption scheme to use:
			//	a) ElGamal
			//	b) Paillier
			//	c) RSA
			// Since TIVI Core only implements ElGamal - it will be our choice
			privKeyShares, pubKey, err := distributed.NewPrivateKeyShares(g, parties, threshold)
			if err != nil {
				t.Fatal(err)
			}

			// ElGamal key shares has been generated, but before we pass it
			// around, we should cast it to the interface
			privKeySharesI := make([]crypto.DecryptionKey, len(privKeyShares))
			for i, privKeyShare := range privKeyShares {
				privKeySharesI[i] = privKeyShare
			}

			// Simulate storing private and public key shares in some other place
			// in a code, this is exactly what we do, when keys are generated
			privKeySharesBytes := make([][]byte, len(privKeySharesI))
			pubKeySharesBytes := make([][]byte, len(privKeySharesI))
			var privKeyShareI crypto.DecryptionKey
			var pubKeyShareI crypto.EncryptionKey

			for i, privKeyShare := range privKeyShares {
				// Only interface is passed in
				privKeyShareI = privKeyShare
				privKeySharesBytes[i], err = storePEMPKCS8PrivateKey(privKeyShareI)
				if err != nil {
					t.Fatal(err)
				}

				// Only interface is passed in
				pubKeyShareI = privKeyShare.EncryptionKey()
				pubKeySharesBytes[i], err = storeDERX509PublicKey(pubKeyShareI)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Interface
			var pubKeyI crypto.EncryptionKey = pubKey
			pubKeyBytes, err := x509DefaultMarshaller.Marshal(pubKeyI)
			if err != nil {
				t.Fatal(err)
			}

			// Now it is time to encrypt some data, for that, we need ElGamal
			// public key, which is currently X509 bytes.
			//
			// Note, that X509 unmarshalling doesn't care about public key
			// implementation, even scheme... It is just X509 bytes unmarshalling
			_, pubKeyRestoredI, err := x509DefaultUnmarshaller.Unmarshal(pubKeyBytes)
			if err != nil {
				t.Fatal(err)
			}

			// Now, let's encrypt some data, suppose we have PREDEFINED candidate
			// list as {John, Milana, Mike}, user has made his choice. NB! We are
			// only allowed to choose one candidate. So this time, we have chosen
			// Milana.
			voterChoice := []bool{false, true, false}

			// Homomorphic ElGamal concrete implementation
			homomorphicEncryption := homomorphic.NewElGamalHomomorphicEncryption(g, threshold)
			// Homomorphic encryption interface, which suits both for distributed
			// and regular encryption schemes
			var homomorphicEncrypter crypto.HomomorphicEncrypter = homomorphicEncryption

			// Encrypt homomorphically. In terms of homomorphic encryption
			// ciphertext == choice
			choice, err := encryptHomomorphic(homomorphicEncrypter, pubKeyRestoredI, voterChoice)
			if err != nil {
				t.Fatal(err)
			}

			// Homomorphic encryption has an amazing property - we can sum up
			// all choices made by all voters without decrypting each choice,
			// and therefore as a result of aggregation we will get a single
			// choice. But please note, that this choice is still encrypted,
			// BUT aggregated. To decrypt a final choice we use tally.
			var homomorphicAggregator crypto.HomomorphicAggregator = homomorphicEncryption
			aggregated, err := aggregateHomomorphic(homomorphicAggregator, pubKeyRestoredI, 1000, choice)
			if err != nil {
				t.Fatal(err)
			}

			// We have our choice encrypted for now, next step is to decrypt
			// our choice to prove that after decryption matches the initial
			// choice
			//
			// But before we do it, obviously we need distributed private key shares,
			// let's get them from PKCS8
			privKeySharesRestoredI := make([]crypto.DecryptionKey, threshold)
			for i, privKeyShareBytes := range privKeySharesBytes[:threshold] {
				_, privKeySharesRestoredI[i], err = pkcs8DistributedUnmarshaller.Unmarshal(privKeyShareBytes)
				if err != nil {
					t.Fatal(err)
				}
			}
			pubKeySharesRestoredI := make([]crypto.EncryptionKey, threshold)
			for i, pubKeyShareBytes := range pubKeySharesBytes[:threshold] {
				_, pubKeySharesRestoredI[i], err = x509DistributedUnmarshaller.Unmarshal(pubKeyShareBytes)
				if err != nil {
					t.Fatal(err)
				}
			}

			// All voter choices are summed up into a single choice, so in
			// out example that single choice will have exactly 3 candidates
			// ({John, Milana, Mike}) and the sum of all votes for them. To
			// obtain decrypted voting results we use tally
			var homomorphicTallier crypto.HomomorphicTallier = homomorphicEncryption
			tallies, err := tallyHomomorphic(homomorphicTallier, privKeySharesRestoredI, pubKeySharesRestoredI, aggregated, nil, 3)
			if err != nil {
				t.Fatal(err)
			}

			// marks = {false, true, false} --> {-, Milana, -}
			marksInt := []uint64{0, 1, 0}
			for i, count := range marksInt {
				if tallies[i].Count() != count {
					t.Fatalf("Mark nr.%d count expected %d, got %d", i, count, tallies[i].Count())
				}
			}

			// In our example, a choice has 3 marks exactly. So, when we
			// tallied results each choice will have 3 decryption proofs (one
			// per mark). And if we have distributed encryption, then a ciphertext
			// is partially decrypted by amount of private key shares, so each choice
			// will have (threshold * len(marks)).
			var homomorphicVerifier crypto.HomomorphicVerifier = homomorphicEncryption
			// Each tally is a decrypted mark with a proof (regular encryption)
			// or proofs (distributed encryption, i.e. each private share decrypts
			// a mark independently)
			//
			// Note, that how we can make this verification all in parallel,
			// just the same way we did with aggregation
			wg := new(sync.WaitGroup)
			wg2 := new(sync.WaitGroup)
			for _, tally := range tallies {
				wg.Add(1)

				go func(tally crypto.HomomorphicDecryption) {
					tallyProofs, err := tally.Proofs()
					if err != nil {
						wg.Done()
						// Log error and return from that particular goroutine
						panic(err) // Don't ever panic in production like that
					}

					// For regular encryption there will only be a single proof, but
					// for distributed there will be threshold amount of proofs, so
					// we verify all of them
					for i, pubKeyShareRestoredI := range pubKeySharesRestoredI {
						wg2.Add(1)
						go func(proof []byte, pubKeyShareRestoredI crypto.EncryptionKey) {
							err = verifyHomomorphic(homomorphicVerifier, pubKeyShareRestoredI, proof, nil)
							if err != nil {
								wg.Done()
								wg2.Done()
								// Log error and return from that particular goroutine
								panic(err) // Don't ever panic in production like that
							}
							wg2.Done()
						}(tallyProofs[i], pubKeyShareRestoredI)
					}
					wg.Done()
				}(tally)
			}

			// Wait all goroutines that verify proofs
			wg2.Wait()
			// Wait all goroutines that spread proofs for verification to wg2
			wg.Wait()

			// That's it! All main e2e workflow is done!
		})
	}
}

//	func TestSignAndVerify(t *testing.T) {
//		data := []byte("Hello World!")
//		for _, g := range group.All() {
//			// Skip non-NIST groups
//			if !strings.Contains(string(g.Name()), "NIST") {
//				continue
//			}
//
//			t.Run(string(g.Name()), func(t *testing.T) {
//				// Generate new ECDSA private key with only Group given, btw,
//				// we derive public key from a private
//				priv, err := ecdsa.New(g)
//				if err != nil {
//					t.Fatal(err)
//				}
//				// Cast to interface
//				var privI crypto.SigningPrivateKey = priv
//				var pubI = priv.EncryptionKey()
//
//				// Since generated keys is likely to be marshalled after creation (and
//				// later reused, on unmarshal), we simulate storing in I/O
//				privBytes, err := storePEMPKCS8PrivateKey(privI)
//				if err != nil {
//					t.Fatal(err)
//				}
//				if strings.Contains(string(g.Name()), "mod") {
//					os.WriteFile(fmt.Sprintf("pub %v", string(g.Name())), privBytes, 0777)
//				}
//				pubBytes, err := storePEMX509PublicKey(pubI)
//				if err != nil {
//					t.Fatal(err)
//				}
//				if strings.Contains(string(g.Name()), "mod") {
//					os.WriteFile(string(g.Name()), privBytes, 0777)
//				}
//				// Simulate reading private key from I/O to concrete implementation
//				// first
//				_, privIRestored, err := DecodePKCS8PrivateKey[crypto.SigningPrivateKey](privBytes)
//				if err != nil {
//					t.Fatal(err)
//				}
//
//				// SignAll custom data
//				signed, err := sign(privIRestored, data)
//				if err != nil {
//					t.Fatal(err)
//				}
//
//				// Simulate reading public key from I/O to concrete implementation
//				// first
//				_, pubIRestored, err := Unmarshal[crypto.SigningPublicKey](pubBytes)
//				if err != nil {
//					t.Fatal(err)
//				}
//
//				// VerifyAll signature
//				err = verify(pubIRestored, signed, data)
//				if err != nil {
//					t.Fatal(err)
//				}
//			})
//		}
//	}
//
//	func TestSignAndVerifyManyData(t *testing.T) {
//		data1 := []byte("Hello World!")
//		data2 := []byte("I'm Alice")
//		data3 := []byte("Good weather")
//		data4 := []byte("Happy coding :)")
//		dataSet := [][]byte{data1, data2, data3, data4}
//
//		for _, g := range group.All() {
//			// Skip non-NIST groups
//			if !strings.Contains(string(g.Name()), "NIST") {
//				continue
//			}
//			t.Run(string(g.Name()), func(t *testing.T) {
//				// Generate new ECDSA signing key
//				var signingKeyI crypto.SigningPrivateKey
//				signingKeyI, err := ecdsa.New(g)
//				if err != nil {
//					t.Fatal(err)
//				}
//
//				// SignAll many data
//				signature, err := crypto.SignAll(signingKeyI, dataSet...)
//				if err != nil {
//					t.Fatal(err)
//				}
//
//				// VerifyAll many data
//				err = crypto.VerifyAll(signingKeyI.EncryptionKey(), signature, dataSet...)
//				if err != nil {
//					t.Fatal(err)
//				}
//			})
//		}
//	}
//var base64ModP3072IVXVPublicKey = "MIIDOTCCAagGCSsGAQQBl1UCATCCAZkCggGBAP//////////yQ/aoiFowjTExmKLgNwc0SkCTgiKZ8x0Agu+pjsTmyJRSgh5jjQE3e+VGbPNOkMbMCsKbfJfFDdP4TVtbVHCReSFtXZiXn7G9ExC6aY37WsL/1y29Aa37e44a/taiZ+lrp8kEXxLH+ZJKGZR7ORbPcIAfLihY78FmNpINhxV05ppFj+o/STPX4NlXSPco62WHGLzViCFUrue1SkHcJaWbWcMNU5KvJgE8XRsCMoYIXwykF5GLjbOO+OedywYDoYDmyeDouwHoo+1xV3wb0xSyd4ry/aVWBcYOZVJfOqVauUV0iYYmPoFEBVyjlqKqsQtrTMXDQRQejOoVSGr3xy6ZOz7hQRY2+8KiupxV10GDH2zlw+FpuHkx6v1rozbCTPXHoyU4EolYZ3O49ImGtLua/Ev+gbZighk2HYCcz7IamRSHysYF3sgDLvhF1d6YV1sdwmIwLrZRuII4k+gdOWrMUPbW/zg/RCOS4LRIKk60sr//////////wIBAhsPVGVzdCBFbGVjdGlvbiAxA4IBiQAwggGEAoIBgEBFmDtb138D0DqHgzmmGJVgxKDr0IZUxWvURIbdcD/E2HlE19DwhY5HfvLco7c7S5atwDPKg900AKgzmoKi27YQmVSVzgdrvUZDUFUuB2exiy/HEXnrPEnB16STUeBPTCqnW3vWZuJePMBGDztnAkQPRIINu075bsAyQklXzkyyIpenn8FJRqNUl7GiwTkFYUwJdwngI/6nVMd3YnfmJxiofYiAy+vd8C8jR3e6to7/lGZ+VfvDnWgC8myF8OqwXgeMzqFWC+pESX9czYHI3b2gtwQFWfZXZ+MoYNwEoMg7nbU8MENGuPcW3ZeZ5jJDc2Dm6ycJXVpJgn/+tdZHp/TDa5RgQXEvokAIk+4aHz49MZA8vKx8oFxUlUOx7OWm+4vvPwvm7W249yhLOYmvT69qoTt2TxAT4fDr7Dl10HogbuXFELnlAt2gQMTebkrpMDUc74diG/mxnVEaW1LxrY/GF+1IrPmxfzo1ll3N6eZ0IhwU6XwGYSzfWiTHF3QDAQ=="
//var base64ModP3072IVXVPrivateKey = "MIIDNwIBADCCAagGCSsGAQQBl1UCATCCAZkCggGBAP//////////yQ/aoiFowjTExmKLgNwc0SkCTgiKZ8x0Agu+pjsTmyJRSgh5jjQE3e+VGbPNOkMbMCsKbfJfFDdP4TVtbVHCReSFtXZiXn7G9ExC6aY37WsL/1y29Aa37e44a/taiZ+lrp8kEXxLH+ZJKGZR7ORbPcIAfLihY78FmNpINhxV05ppFj+o/STPX4NlXSPco62WHGLzViCFUrue1SkHcJaWbWcMNU5KvJgE8XRsCMoYIXwykF5GLjbOO+OedywYDoYDmyeDouwHoo+1xV3wb0xSyd4ry/aVWBcYOZVJfOqVauUV0iYYmPoFEBVyjlqKqsQtrTMXDQRQejOoVSGr3xy6ZOz7hQRY2+8KiupxV10GDH2zlw+FpuHkx6v1rozbCTPXHoyU4EolYZ3O49ImGtLua/Ev+gbZighk2HYCcz7IamRSHysYF3sgDLvhF1d6YV1sdwmIwLrZRuII4k+gdOWrMUPbW/zg/RCOS4LRIKk60sr//////////wIBAhsPVGVzdCBFbGVjdGlvbiAxBIIBhASCAYBsSMVGGEoPLwC559hxFXZIuiHFpwFbEYm/0o6lZYuIS1e2hrPRcVJ0v0eUA41RB5giaMf7M/kBaq/c6QGJ6UVWWX7Xhnzk7xMwwBmfnDXzhfkKbMZzlrGASFsK6KcrroYEEFWkW6DYsG3NaWTV6hgXTAzisgC7WWvDtqAxF0ZTMg81MJj1yVgHklQ6rTPXfHBrFSi6XNKPSNNU0H66qICSnWcuwpUsILgq2vPHzjTcfjbr/PyRButIO0LrHlEInqK1uLaHyl1I6GWsvoJfFbm16rSLTFeRGTDTxjamwLI2PgEO5zN6KF2EVcmOoHEeTVguN4958VavDUDrK/sQalXwZrC6BoiIcbLgZF0xmm4EN9tzCk5uKBds6GSbDjkgCJsu8evnZBL6ces7VujW41E5q+OmZesWsJ99+ejuhT+vYbIz6B35q0/pUonYzNt/f4bBCS3AIhJCiCh/dnx5ekwsQCaAxUg7KYa5XI9a1mr9FpdmGCjqpZ48qvxLcUbW5D8="
//
//var base64NistP384IVXVPublicKey = "MIGPMCUGCSsGAQQBho0fATAYGwVQLTM4NBsPVGVzdCBFbGVjdGlvbiAxA2YAMGMEYQR7W65Z9VYdHrA3zC5wuXhQnD2nfuvTNRNJrrp4kDJ2U0/+tzptXuN3J/4heawE8eLvoKnV6IfOf+qKCfnoV3qByCdd7FCmYrxo0EYvgHz006BGWN/UvSYNETdcmS8MFWI="
//var base64NistP384IVXVPrivateKey = "MGECAQAwJQYJKwYBBAGGjR8BMBgbBVAtMzg0Gw9UZXN0IEVsZWN0aW9uIDEENQQzAjEAtsRtj3M7c4hln67tQCVZnoOYebz1jge3VxCMbmE1cJAC3WXYAy54rELm/ltbqiXW"
//
//type IVXVAlgorithmIdentifier struct {
//	algorithm  asn1.ObjectIdentifier
//	parameters ModPGroupParameters
//}
//
//type ModPGroupParameters struct {
//	P          *big.Int
//	g          *big.Int
//	ElectionID string
//}
//
//type ModPGroupKey struct {
//	Key big.Int
//}
//
//type IVXVPkcs8Unmarshaller[T group.Scalar] struct{}
//
//func NewIVXVPkcs8Unmarshaller[T group.Scalar]() IVXVPkcs8Unmarshaller[T] {
//	return IVXVPkcs8Unmarshaller[T]{}
//}
//
//func (x509 IVXVPkcs8Unmarshaller[T]) UnmarshalKeyParameters(algorithmIdentifier pkix.AlgorithmIdentifier) (parameters crypto.AlgorithmIdentifierParameters, err error) {
//	var params ModPGroupParameters
//	_, err = asn1.Unmarshal(algorithmIdentifier.Group.FullBytes, &params)
//	if err != nil {
//		return nil, err
//	}
//
//	var name string
//	if params.P.BitLen() == 3072 {
//		name = "RFC3526ModPGroup3072"
//	} else {
//		name = "NIST-P384"
//	}
//	g, err := group.Get(name)
//	if err != nil {
//		return nil, err
//	}
//
//	return elgamal.Group{g: g}, nil
//}
//
//func (x509 IVXVPkcs8Unmarshaller[T]) UnmarshalKeyElement(g group.g, bytes []byte) (element T, err error) {
//	//var key ModPGroupKey
//
//	c := cryptobyte.String(bytes)
//	var ss cryptobyte.String
//	//var sss cryptobyte.String
//	if !c.ReadASN1(&ss, asn_1.OCTET_STRING) {
//		return element, fmt.Errorf("Unmarshal SEQUENCE")
//	}
//	//
//	//if !ss.ReadAnyASN1(&sss, nil) {
//	//	return element, fmt.Errorf("Unmarshal SEQUENCE2")
//	//}
//
//	var builder cryptobyte.Builder
//	builder.AddASN1(asn_1.INTEGER, func(c *cryptobyte.Builder) {
//		c.AddBytes(ss)
//	})
//
//	b, err := builder.Bytes()
//	if err != nil {
//		return element, fmt.Errorf("Unmarshal INTEGER")
//	}
//
//	el, err := elgamal.ASN1UnmarshalPrivateElement(g, b)
//	if err != nil {
//		return element, err
//	}
//
//	return T(el), nil
//}
//
//type IVXVX509Unmarshaller[T group.Element] struct{}
//
//func NewIVXVX509Unmarshaller[T group.Element]() IVXVX509Unmarshaller[T] {
//	return IVXVX509Unmarshaller[T]{}
//}
//
//func (x509 IVXVX509Unmarshaller[T]) UnmarshalKeyParameters(algorithmIdentifier pkix.AlgorithmIdentifier) (parameters crypto.AlgorithmIdentifierParameters, err error) {
//	var params ModPGroupParameters
//	_, err = asn1.Unmarshal(algorithmIdentifier.Group.FullBytes, &params)
//	if err != nil {
//		return nil, err
//	}
//
//	var name string
//	if params.P.BitLen() == 3072 {
//		name = "RFC3526ModPGroup3072"
//	} else {
//		name = "NIST-P384"
//	}
//	g, err := group.Get(name)
//	if err != nil {
//		return nil, err
//	}
//
//	return elgamal.Group{g: g}, nil
//}
//
//func (x509 IVXVX509Unmarshaller[T]) UnmarshalKeyElement(g group.g, bytes []byte) (element T, err error) {
//	c := cryptobyte.String(bytes)
//	var ss cryptobyte.String
//	if !c.ReadASN1(&ss, asn_1.SEQUENCE) {
//		return element, fmt.Errorf("Unmarshal SEQUENCE")
//	}
//
//	el, err := elgamal.ASN1UnmarshalPublicElement(g, ss)
//	if err != nil {
//		return element, err
//	}
//
//	return el.(T), nil
//}
//
//func TestCustomUnmarshaller(t *testing.T) {
//	pem, err := base64.StdEncoding.DecodeString(base64ModP3072IVXVPublicKey)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ivxvPublicKeyUnmarshaller := NewUnmarshaller[crypto.EncryptionKey, group.Element](NewIVXVX509Unmarshaller[group.Element]())
//	_, pkey, err := ivxvPublicKeyUnmarshaller.Unmarshal(pem)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	pem2, err := base64.StdEncoding.DecodeString(base64ModP3072IVXVPrivateKey)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ivxvPrivateKeyUnmarshaller := NewUnmarshallerPKCS8[crypto.DecryptionKey, group.Scalar](NewIVXVPkcs8Unmarshaller[group.Scalar]())
//	_, key, err := ivxvPrivateKeyUnmarshaller.Unmarshal(pem2)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	plaintext := []byte("Hello World! =)")
//
//	fmt.Println(pkey.Marshal())
//	fmt.Println(key.EncryptionKey().Marshal())
//
//	// a part of ciphertext
//	rand, err := group.RandomScalar(pkey.Group().Group().Order())
//	ciphertext, err := encrypt(pkey, plaintext, rand)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	decrypted, err := key.Decrypt(ciphertext)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	plaintext2, err := decrypted.Plaintext()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if !bytes.Equal(plaintext, plaintext2) {
//		t.Fatal("Plaintext after decryption doesn't match the initial one")
//	}
//
//	fmt.Println(plaintext, plaintext2)
//
//}
