package bdoc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"ivxv.ee/container"
	"ivxv.ee/yaml"
)

const (
	yamlConfLoc = "testdata/conf.yaml"

	dataKey   = "test.txt"
	dataValue = "Test data"
)

var testOpener *Opener

func init() {
	var err error

	if testOpener, err = testLoadConf(); err != nil {
		fmt.Printf("Failed to parse conf: %v", err)
		os.Exit(1)
	}
}

func testLoadConf() (o *Opener, err error) {
	yamlTestConf, err := os.Open(yamlConfLoc)
	if err != nil {
		return
	}
	node, err := yaml.Parse(yamlTestConf, nil)
	if err != nil {
		return
	}
	var c Conf
	if err = yaml.Apply(node, &c); err != nil {
		return
	}

	return New(&c)
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name              string
		signers           []string
		checkTM           bool
		expectedFailure   bool
		expectedFileCount int
	}{
		// TM signatures
		{"EID", []string{"MÄNNIK,MARI-LIIS,47101010033"}, true, false, 1},
		{"MID", []string{"O’CONNEŽ-ŠUSLIK,MARY ÄNN,11412090004"}, true, false, 1},
		{
			"MultipleSigners",
			[]string{"MÄNNIK,MARI-LIIS,47101010033", "ŽAIKOVSKI,IGOR,37101010021"},
			true, false, 1,
		},
		{"MultipleFiles", []string{"ŽAIKOVSKI,IGOR,37101010021"}, true, false, 2},
		// BES signatures
		{"CheckOCSP", []string{"ŽAIKOVSKI,IGOR,37101010021"}, true, true, 1},
		{"CheckOCSP", []string{"ŽAIKOVSKI,IGOR,37101010021"}, false, false, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Open test bdoc file
			file, err := os.Open(filepath.Join("testdata",
				fmt.Sprintf("test%s.bdoc", test.name)))
			if err != nil {
				t.Fatal("Failed to open BDOC:", err)
			}
			defer file.Close() // nolint: errcheck, ignore close failure of read-only fd.

			testOpener.checkTM = test.checkTM
			bdoc, err := testOpener.Open(file)
			if test.expectedFailure && err == nil {
				t.Fatal("Expected failure verifying BDOC")
			} else if test.expectedFailure && err != nil {
				return
			} else if !test.expectedFailure && err != nil {
				t.Fatalf("Failure verifying BDOC: %v", err)
			}

			s := bdoc.Signatures()
			if len(s) != len(test.signers) {
				t.Fatal("unexpected signers count:", len(s))
			}
			if ret := testCompareNames(s, test.signers); ret != "" {
				t.Fatalf("Signer common name error: %s", ret)
			}

			doc := bdoc.Data()
			if len(doc) != test.expectedFileCount {
				t.Fatal("unexpected data key count:", len(doc))
			}
			if len(doc) == 1 {
				if val, ok := doc[dataKey]; !ok {
					t.Fatalf("missing data key %q", dataKey)
				} else if !bytes.Equal(val, []byte(dataValue)) {
					t.Fatalf("unexpected data value of key %q: %x", dataKey, val)
				}
			}
		})
	}
}

func testCompareNames(signatures []container.Signature, signers []string) string {
	for _, signer := range signers {
		var found bool
		for _, signature := range signatures {
			if signature.Signer.Subject.CommonName == signer {
				found = true
				break
			}
		}
		if !found {
			return fmt.Sprintf("Expected name %s not found", signer)
		}
	}
	return ""
}
