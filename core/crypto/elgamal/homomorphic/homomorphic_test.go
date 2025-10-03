package homomorphic

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"tivi.io/core/crypto"
	oelgamal "tivi.io/core/crypto/elgamal"
	"tivi.io/core/crypto/elgamal/distributed"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

func TestMarkCompleteness(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"zero", false},
		{"one", true},
	}

	for _, tc := range tests {
		for _, g := range group.All() {
			opts := generateOptions(t, g)
			t.Run(tc.name, func(t *testing.T) {
				mark, _, err := NewMark(opts.Public(), tc.value)
				if err != nil {
					t.Fatalf("new Mark error: %s", err)
				}
				verified := mark.Verify(opts.Public())
				if verified != nil {
					t.Fatalf("Mark verification error: %s", verified)
				}
			})
		}
	}
}

func TestChoiceCompleteness(t *testing.T) {
	for _, g := range group.All() {
		tests := []struct {
			name  string
			marks []bool
			fails bool
		}{
			{"no_marks", []bool{}, true},
			{"no_true_mark", []bool{false, false, false, false, false}, true},
			{"single_mark", []bool{false, false, false, true, false}, false},
			{"two_marks", []bool{true, false, true, false, false}, true},
		}
		opts := generateOptions(t, g)
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ch, err := Encrypt(opts.Public(), tc.marks)
				if err != nil {
					t.Fatal("error constructing Choice", err)
				}
				if err := ch.Verify(opts.Public()); (err != nil) != tc.fails {
					fmt.Println(err)
					t.Fatalf("expected verification: %t, got %t", tc.fails, err != nil)
				}
			})
		}
	}
}

func TestChoiceAdd(t *testing.T) {
	for _, g := range group.All() {
		tests := []struct {
			name  string
			marks [][]bool
			fails bool
		}{
			{"compatible_choices", [][]bool{{false, false, true}, {true, false, false}}, false},
			{"incompatible_choices", [][]bool{{false, false, true, false}, {false, true, false}}, true},
			{"many_choices", [][]bool{{false, true}, {true, false}, {false, true}, {true, false}, {false, true}}, false},
			{"no_choices", [][]bool{}, true},
		}
		opts := generateOptions(t, g)
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var err error
				chs := make([][]byte, len(tc.marks))
				for i, marks := range tc.marks {
					tmp, err := Encrypt(opts.Public(), marks)
					if err != nil {
						t.Fatal("error creating Choice", err)
					}
					chs[i], err = tmp.MarshalASN1()
					if err != nil {
						t.Fatal("error creating Choice", err)
					}
				}
				_, err = Aggregate(opts.Public(), true, chs...)
				if (err != nil) != tc.fails {
					t.Fatalf("expected fails %t, got %t", tc.fails, err != nil)
				}
			})
		}
	}
}

func TestDiscreteLog(t *testing.T) {
	for _, g := range group.All() {
		maxCount := 1024

		dl0, err := g.Identity().Scale(group.OneScalar(g.Order()))
		if err != nil {
			t.Fatal(err)
		}

		dl1 := g.Generator()

		dl1024, err := g.Generator().Scale(group.NewScalar(big.NewInt(1024), g.Order()))
		if err != nil {
			t.Fatal(err)
		}

		dl1025, err := g.Generator().Scale(group.NewScalar(big.NewInt(1025), g.Order()))
		if err != nil {
			t.Fatal(err)
		}

		dlrand, err := group.RandomElement(g)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			name  string
			dec   group.Element
			count int
			fails bool
		}{
			{"DL-0", dl0, 0, false},
			{"DL-1", dl1, 1, false},
			{"DL-1024", dl1024, 1024, false},
			{"DL-1025", dl1025, 0, true},
			{"nil", nil, 0, true},
			{"random", dlrand, 0, true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				count, err := DiscreteLog(g, maxCount, tc.dec)
				if (err != nil) != tc.fails {
					t.Fatalf("expected failure %t, got %t", tc.fails, (err != nil))
				}
				if count != tc.count {
					t.Fatalf("expected count %d, got %d", tc.count, count)
				}
			})
		}
	}
}

func generateOptionsWithParameters2(t *testing.T, gr group.Group) *oelgamal.PrivateKey {
	t.Helper()
	params := oelgamal.NewParameters(gr)
	rnd, err := group.RandomScalar(params.Group().Order())
	priv, err := oelgamal.NewPrivateKey(params, rnd)
	if err != nil {
		t.Fatalf("error generating public key: %s", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	return priv
}

func generateOptions(t *testing.T, g group.Group) *oelgamal.PrivateKey {
	return generateOptionsWithParameters2(t, g)
}
func generateNonStandardOptions(t *testing.T, g group.Group) *oelgamal.PrivateKey {
	return generateOptionsWithParameters2(t, g)
}

func testDistributedHomomorphicEncryption(t *testing.T, g group.Group, _ *oelgamal.PrivateKey) {
	tests := []struct {
		name     string
		marks    [][]bool
		expected []int
	}{
		{"single_choice", [][]bool{{false, false, true}}, []int{0, 0, 1}},
		{"multiple_choices-single_mark", [][]bool{{true, false, false}, {false, true, false}, {false, false, true}}, []int{1, 1, 1}},
		{"multiple_choices-multiple_marks", [][]bool{{true, false, false}, {true, false, false}}, []int{2, 0, 0}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// First of all, we start with a workflow, when user generates private
			// key automatically
			//
			// User decides which private key to create:
			//	a) Regular
			//	b) Distributed

			// This time we decided to create Distributed one.
			//
			// Note that we say explicitly, that Distributed encryption is
			// ElGamal. However, further in a code we cast these concrete types
			// to the interfaces.
			// When, for example, you don't want to use ElGamal in a future,
			// you easily could replace Distributed ElGamal with, for example,
			// Distributed Paillier or Distributed RSA, if such exist.
			//
			// Keeping that in mind, it is strongly advised to use concrete types
			// in a specific place inside a code where initialization is done, i.e.:
			//	a) init() func
			//	b) main() func
			//	c) REST controller/service
			//
			// Rule of thumb - don't let concrete types to flow deeply into a code,
			// keep them in a one single place and always pass inside the interfaces.
			privSs, pk, err := distributed.NewPrivateKeyShares(g, 3, 2)

			// Cast to interfaces
			//	var elgamalPubKey crypto.EncryptionKey = pk
			privShares := make([]crypto.DecryptionKey, len(privSs))
			for i, privS := range privSs {
				privShares[i] = privS
			}

			// Derive public key shares from private key shares
			pubShares := make([]crypto.EncryptionKey, len(privShares))
			for i, share := range privShares {
				pubShares[i] = share.EncryptionKey()
			}

			// This is the place where we have finished with Distributed ElGamal
			// private key creation, and it is a time to store the material in I/O,
			// since most likely, it will be returned to the client via either
			// disk storing or Web call.

			// Store private key shares in I/O
			privSharesBytes := make([][]byte, len(privShares))
			for i, share := range privShares {
				privSharesBytes[i], err = share.Marshal()
				if err != nil {
					t.Fatal(err)
				}
			}

			// Store public key shares in I/O
			pubSharesBytes := make([][]byte, len(pubShares))
			for i, share := range pubShares {
				pubSharesBytes[i], err = share.Marshal()
				if err != nil {
					t.Fatal(err)
				}
			}
			zk := &ElGamalHomomorphicEncryption{g: g, threshold: 3}
			// Cast to interface
			var pubkeyRestored crypto.HomomorphicEncryption = zk

			// Homomorphic ciphertexts == choices
			choices := make([][]byte, len(tc.marks))
			for i, marks := range tc.marks {

				// Homomorphically encrypt user choices
				choices[i], err = pubkeyRestored.Encrypt(pk, marks)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Aggregate choices in parallel
			wg := new(sync.WaitGroup)
			batchSize := 1000
			var aggregated []byte
			lock := new(sync.Mutex)

			// Use homomorphic property of each choice to aggregate them together
			for i := 0; i < len(choices); i += batchSize {
				end := i + batchSize
				if end > len(choices) {
					end = len(choices)
				}
				batch := choices[i:end]

				// Each goroutine call will be recorded and using wg.Wait()
				// will be awaited
				wg.Add(1)

				// Use goroutines to make it parallel
				go func() {
					// Aggregate batches (use as large batch as you can)
					aggregated, err = pubkeyRestored.Aggregate(pk, true, batch...)
					if err != nil {
						// Log error, continue
						return
					}

					// This is a blocking goroutine
					go func() {
						lock.Lock()
						defer lock.Unlock()

						// Aggregate aggregated batches
						aggregated, err = pubkeyRestored.Aggregate(pk, false, aggregated)
						if err != nil {
							// Log error, continue
							return
						}

						// Mark goroutine as finished
						wg.Done()
					}()
				}()
			}

			// Block and wait for all unfinished goroutines (code didn't reach
			// wg.Done() statement
			wg.Wait()

			// DecryptAll+tally aggregated choices
			//
			// Tallies contain voting result
			talllies, err := pubkeyRestored.Tally(privShares, pubShares, aggregated, len(choices), nil)
			if err != nil {
				t.Fatal(err)
			}

			for i, count := range tc.expected {
				if talllies[i].Count() != uint64(count) {
					t.Fatalf("Mark nr.%d count expected %d, got %d", i, count, talllies[i].Count())
				}
			}

			// Store tally in I/O for later auditing (proofs)
			talliesBytes := make([][]byte, len(talllies))
			for i, tally := range talllies {
				talliesBytes[i], err = tally.Marshal()
				if err != nil {
					t.Fatal(err)
				}
			}

			for _, tally := range talllies {
				tallyProofsBytes, err := tally.Proofs()
				if err != nil {
					t.Fatal(err)
				}
				for i, pubShare := range pubShares {
					err = pubkeyRestored.Verify(pubShare, tallyProofsBytes[i], nil)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}

}

func testHomomorphicEncryption(t *testing.T, g group.Group, _ *oelgamal.PrivateKey) {
	tests := []struct {
		name     string
		marks    [][]bool
		expected []int
	}{
		{"single_choice", [][]bool{{false, false, true}}, []int{0, 0, 1}},
		{"multiple_choices-single_mark", [][]bool{{true, false, false}, {false, true, false}, {false, false, true}}, []int{1, 1, 1}},
		{"multiple_choices-multiple_marks", [][]bool{{true, false, false}, {true, false, false}}, []int{2, 0, 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// User decides which private key to create:
			//	a) Regular
			//	b) Distributed

			// This time
			params := oelgamal.NewParameters(g)
			rnd, err := group.RandomScalar(params.Group().Order())
			privKey, err := oelgamal.NewPrivateKey(params, rnd)
			if err != nil {
				t.Fatal(err)
			}

			// User decides which homomorphic public key to create
			// based on existing private key
			homoKey := NewElGamalHomomorphicEncryption(g, 1)
			choices := make([][]byte, len(tc.marks))
			for i, marks := range tc.marks {
				// Homomorphically Encrypt user choices
				choices[i], err = homoKey.Encrypt(privKey.EncryptionKey(), marks)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Aggregate choices in parallel
			wg := new(sync.WaitGroup)
			batchSize := 1000
			var aggregated []byte
			lock := new(sync.Mutex)

			// Use homomorphic property of each Choice to Aggregate them together
			for i := 0; i < len(choices); i += batchSize {
				end := i + batchSize
				if end > len(choices) {
					end = len(choices)
				}
				batch := choices[i:end]

				// Each goroutine call will be recorded and using wg.Wait()
				// will be awaited
				wg.Add(1)

				// Use goroutines to make it parallel
				go func() {
					// Aggregate batches (use as large batch as you can)
					aggregated, err = homoKey.Aggregate(privKey.EncryptionKey(), true, batch...)
					if err != nil {
						t.Fatal(err)
						wg.Done()
					}

					// This is a blocking goroutine
					go func() {
						lock.Lock()
						defer lock.Unlock()

						// Aggregate aggregated batches
						aggregated, err = homoKey.Aggregate(privKey.EncryptionKey(), false, aggregated)
						if err != nil {
							t.Fatal(err)
						}

						// Mark goroutine as finished
						wg.Done()
					}()
				}()
			}

			// Block and wait for all unfinished goroutines (code didn't reach
			// wg.Done() statement
			wg.Wait()

			// DecryptAll+tally aggregated choices
			//
			// Tally contains decrypted Result and proof
			privKeys := []crypto.DecryptionKey{privKey}
			pubKeys := []crypto.EncryptionKey{privKey.Public()}
			tally, err := homoKey.Tally(privKeys, pubKeys, aggregated, len(choices), nil)
			if err != nil {
				t.Fatal(err)
			}

			for j, count := range tc.expected {
				if tally[j].Count() != uint64(count) {
					t.Fatalf("Mark %d count expected %d, got %d", j, count, tally[j].Count())
				}
			}

			// Store tally in I/O for later auditing (proof)

			for _, a := range tally {
				go func(a crypto.HomomorphicDecryption) {
					proofsPerTally, err := a.Proofs()
					if err != nil {
						// Log error, continue
						return
					}
					for _, proofPerTally := range proofsPerTally {
						go func(proofPerTally []byte) {
							err = homoKey.Verify(privKey.EncryptionKey(), proofPerTally, nil)
							if err != nil {
								// Log error, continue
								return
							}
						}(proofPerTally)
					}
				}(a)
			}
		})
	}

}

func TestTallyCompletenessDistributedHomo(t *testing.T) {
	for _, g := range group.All() {
		priv := generateOptions(t, g)
		testDistributedHomomorphicEncryption(t, g, priv)
	}
}

func TestTallyCompletenessRegularHomo(t *testing.T) {
	for _, g := range group.All() {
		priv := generateOptions(t, g)
		testHomomorphicEncryption(t, g, priv)
	}
}

func TestTallyCompletenessNonStandard(t *testing.T) {
	for _, g := range group.All() {
		priv := generateNonStandardOptions(t, g)
		testDistributedHomomorphicEncryption(t, g, priv)
	}
}
