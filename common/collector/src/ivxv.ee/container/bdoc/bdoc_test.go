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
		profile           Profile
		expectedFailure   bool
		expectedFileCount int
	}{
		// ID-card signature with TS profile and AIA OCSP response is
		// valid for BES and TS, but not TM.
		{"EIDTS", []string{"JÕEORG,JAAK-KRISTJAN,38001085718"}, BES, false, 1},
		{"EIDTS", []string{"JÕEORG,JAAK-KRISTJAN,38001085718"}, TS, false, 1},
		{"EIDTS", []string{"JÕEORG,JAAK-KRISTJAN,38001085718"}, TM, true, 1},

		// Mobile-ID signature with TS profile and TM OCSP response is
		// valid for BES and TS, but not TM (because of
		// SignaturePolicyIdentifier).
		{"MIDTS", []string{"O’CONNEŽ-ŠUSLIK TESTNUMBER,MARY ÄNN,60001018800"}, BES, false, 1},
		{"MIDTS", []string{"O’CONNEŽ-ŠUSLIK TESTNUMBER,MARY ÄNN,60001018800"}, TS, false, 1},
		{"MIDTS", []string{"O’CONNEŽ-ŠUSLIK TESTNUMBER,MARY ÄNN,60001018800"}, TM, true, 1},

		// ID-card signature with TM profile is valid for TM, but not
		// BES (because of SignaturePolicyIdentifier) and TS.
		{"EIDTM", []string{"MÄNNIK,MARI-LIIS,47101010033"}, TM, false, 1},
		{"EIDTM", []string{"MÄNNIK,MARI-LIIS,47101010033"}, BES, true, 1},
		{"EIDTM", []string{"MÄNNIK,MARI-LIIS,47101010033"}, TS, true, 1},

		// Mobile-ID signature with TM profile is valid for TM, but not
		// BES (because of SignaturePolicyIdentifier) and TS.
		{"MIDTM", []string{"O’CONNEŽ-ŠUSLIK,MARY ÄNN,11412090004"}, TM, false, 1},
		{"MIDTM", []string{"O’CONNEŽ-ŠUSLIK,MARY ÄNN,11412090004"}, BES, true, 1},
		{"MIDTM", []string{"O’CONNEŽ-ŠUSLIK,MARY ÄNN,11412090004"}, TS, true, 1},

		// ID-card signature with no qualification is valid for BES,
		// but not TM and TS.
		{"EIDBES", []string{"ŽAIKOVSKI,IGOR,37101010021"}, BES, false, 1},
		{"EIDBES", []string{"ŽAIKOVSKI,IGOR,37101010021"}, TM, true, 1},
		{"EIDBES", []string{"ŽAIKOVSKI,IGOR,37101010021"}, TS, true, 1},

		// Containers with multiple signers and files.
		{
			"MultipleSigners",
			[]string{"MÄNNIK,MARI-LIIS,47101010033", "ŽAIKOVSKI,IGOR,37101010021"},
			TM, false, 1,
		},
		{"MultipleFiles", []string{"ŽAIKOVSKI,IGOR,37101010021"}, TM, false, 2},

		// Containers with missing files.
		{"NoManifest", nil, BES, true, 1},
		{"NoSignatures", nil, BES, true, 1},
		{"NoFiles", []string{"ŽAIKOVSKI,IGOR,37101010021"}, BES, true, 0},

		// Containers with invalid OCSP response times.
		{"OCSPOld", nil, TS, true, 1},
		{"OCSPDelayed", nil, TS, true, 1},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s as %s", test.name, test.profile), func(t *testing.T) {
			// Open test bdoc file
			file, err := os.Open(filepath.Join("testdata",
				fmt.Sprintf("test%s.bdoc", test.name)))
			if err != nil {
				t.Fatal("Failed to open BDOC:", err)
			}
			defer file.Close()

			testOpener.profile = test.profile
			bdoc, err := testOpener.Open(file)
			switch {
			case test.expectedFailure && err != nil:
				return
			case test.expectedFailure && err == nil:
				t.Fatal("Expected failure verifying BDOC")
			case !test.expectedFailure && err != nil:
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
