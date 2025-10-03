package bdoc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"ivxv.ee/common/collector/container"
	"ivxv.ee/common/collector/yaml"
)

const (
	// Path to the trust.yml for BDOC TS profile.
	trustConfTS = "testdata/trustTS.yaml"
	// Path to the trust.yml for deprecated BDOC TM profile.
	trustConfTM = "testdata/trustTM.yaml"
	// Path to the trust.yml for BDOC BES profile.
	trustConfBES = "testdata/trustBES.yaml"

	dataKey   = "test.txt"
	dataValue = "Test data"

	// Deprecated: TM
	//
	// BDOC Opener should not be configured to support TM profiles anymore.
	//
	// SK ID Solutions:
	//
	// The modified service cannot be used to create signatures in BDOC-TM
	// format. Signatures created in BDOC-TM format from November 1, 2023 can
	// no longer be validated.
	//
	// Reference:
	// https://www.skidsolutions.eu/news/important-changes-in-validity-confirmation-service/
	TM = "TM"

	// In unzipped BDOC container, in META-INF/signatures0.xml, <ds:Signature Id="S0">
	signatureIDS0 = "S0"
	// In unzipped BDOC container, in META-INF/signatures1.xml, <ds:Signature Id="S1">
	signatureIDS1 = "S1"
)

const expectedEqualObjects = "Expected o1 == o2, got o1: %v, o2: %v\n"

// Supplier is an any function that takes no parameters and returns some value.
type Supplier func() interface{}

func unsupportedProfileErrorMessageSupplier(profile Profile) Supplier {
	return func() interface{} {
		profileErr := new(UnsupportedProfileError)
		profileErr.Profile = profile
		profileErr.Description = _CONTAINER_BDOC_PROFILE
		return *profileErr
	}
}

func tsProfileOCSPDelayedErrorMessageSupplier(signatureID string) Supplier {
	return func() interface{} {
		signatureError := new(CheckSignatureError)
		signatureError.Signature = signatureID
		timestampErr := new(TimestampAndOCSPTimeMismatchError)
		timestampErr.OCSPProducedAt = time.Date(2021, 4, 13, 9, 56, 6, 0, time.UTC)
		timestampErr.TimestampGenTime = time.Date(2021, 4, 8, 8, 20, 41, 0, time.UTC)
		timestampErr.Description = _CONTAINER_BDOC_TS
		signatureError.Err = *timestampErr
		signatureError.Description = _CONTAINER_BDOC_CERT_OCSP
		return *signatureError
	}
}

func tsProfileOCSPOldErrorMessageSupplier(signatureID string) Supplier {
	return func() interface{} {
		signatureError := new(CheckSignatureError)
		signatureError.Signature = signatureID
		timestampErr := new(TimestampAndOCSPTimeMismatchError)
		timestampErr.TimestampGenTime = time.Date(2021, 4, 13, 9, 56, 6, 0, time.UTC)
		timestampErr.OCSPProducedAt = time.Date(2021, 4, 8, 8, 20, 41, 0, time.UTC)
		timestampErr.Description = _CONTAINER_BDOC_TS
		signatureError.Err = *timestampErr
		signatureError.Description = _CONTAINER_BDOC_CERT_OCSP
		return *signatureError
	}
}

func noDataFilesErrorMessageSupplier() Supplier {
	return func() interface{} {
		signatureError := new(OpenBDOCContainerError)
		manifestError := new(NoDataFilesError)
		manifestError.Description = _CONTAINER_BDOC_DATA
		signatureError.Err = *manifestError
		signatureError.Description = _CONTAINER_BDOC_OPEN
		return *signatureError
	}
}

func noSignatureErrorMessageSupplier() Supplier {
	return func() interface{} {
		signatureError := new(OpenBDOCContainerError)
		manifestError := new(NoSignaturesError)
		manifestError.Description = _CONTAINER_BDOC_DATA
		signatureError.Err = *manifestError
		signatureError.Description = _CONTAINER_BDOC_OPEN
		return *signatureError
	}
}

func manifestErrorMessageSupplier() Supplier {
	return func() interface{} {
		signatureError := new(OpenBDOCContainerError)
		manifestError := new(MissingManifestError)
		manifestError.Description = _CONTAINER_BDOC_MANI
		signatureError.Err = *manifestError
		signatureError.Description = _CONTAINER_BDOC_OPEN
		return *signatureError
	}
}

func tmProfilePolicyErrorMessageSupplier(signatureID string) Supplier {
	return func() interface{} {
		signatureError := new(CheckSignatureError)
		signatureError.Signature = signatureID
		policyErr := new(UnexpectedSignaturePolicyIdentifierError)
		policyErr.Identifier = "urn:oid:1.3.6.1.4.1.10015.1000.3.2.1"
		policyErr.Description = _CONTAINER_BDOC_SIG_POLICY
		signatureError.Err = *policyErr
		signatureError.Description = _CONTAINER_BDOC_CERT_OCSP
		return *signatureError
	}
}

func tmProfileTimestampErrorMessageSupplier(signatureID string) Supplier {
	return func() interface{} {
		signatureError := new(CheckSignatureError)
		signatureError.Signature = signatureID
		timestampErr := new(TimestampMissingError)
		timestampErr.Description = _CONTAINER_BDOC_CERT_N_TSA
		signatureError.Err = *timestampErr
		signatureError.Description = _CONTAINER_BDOC_CERT_OCSP
		return *signatureError
	}
}

// loadTrustConf loads necessary trust.yml for the given profile.
func loadTrustConf(profile Profile) (*Opener, error) {
	var confPath string

	switch profile {
	case TS:
		confPath = trustConfTS
	case TM:
		confPath = trustConfTM
	case BES:
		confPath = trustConfBES
	}

	return testLoadConf(confPath)
}

// testLoadConf configures BDOC Opener using confPath.
func testLoadConf(confPath string) (o *Opener, err error) {
	yamlTestConf, err := os.Open(confPath)
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

//nolint:lll
func TestOpen(t *testing.T) {
	tests := []struct {
		name              string
		signers           []string
		profile           Profile
		expectedFailure   bool
		expectedFileCount int
		failure           Supplier // nil means no error is expected
	}{
		// ID-card signature with TS profile and AIA OCSP response is
		// valid for BES and TS.
		{"EIDTS", []string{"JÕEORG,JAAK-KRISTJAN,38001085718"}, BES, false, 1, nil},
		{"EIDTS", []string{"JÕEORG,JAAK-KRISTJAN,38001085718"}, TS, false, 1, nil},
		// TM profile is not supported anymore
		{"EIDTS", []string{"JÕEORG,JAAK-KRISTJAN,38001085718"}, TM, true, 1, unsupportedProfileErrorMessageSupplier(TM)},

		// Mobile-ID signature with TS profile and non-AIA OCSP response is
		// valid for BES and TS.
		{"MIDTS", []string{"O’CONNEŽ-ŠUSLIK TESTNUMBER,MARY ÄNN,60001018800"}, BES, false, 1, nil},
		{"MIDTS", []string{"O’CONNEŽ-ŠUSLIK TESTNUMBER,MARY ÄNN,60001018800"}, TS, false, 1, nil},
		// TM profile is not supported anymore
		{"MIDTS", []string{"O’CONNEŽ-ŠUSLIK TESTNUMBER,MARY ÄNN,60001018800"}, TM, true, 1, unsupportedProfileErrorMessageSupplier(TM)},

		// TM profiles are not supported anymore, and even if operator configured
		// BDOC Opener to support only TS or BES profiles via trust.yml config
		// file, it doesn't restrict user to sign BDOC containers with ID card TM
		// profile. Backend should fail with SignaturePolicyIdentifier
		{"EIDTM", []string{"MÄNNIK,MARI-LIIS,47101010033"}, TM, true, 1,
			unsupportedProfileErrorMessageSupplier(TM)},
		{"EIDTM", []string{"MÄNNIK,MARI-LIIS,47101010033"}, BES, true, 1,
			tmProfilePolicyErrorMessageSupplier(signatureIDS0)},
		{"EIDTM", []string{"MÄNNIK,MARI-LIIS,47101010033"}, TS, true, 1,
			tmProfilePolicyErrorMessageSupplier(signatureIDS0)},

		// TM profiles are not supported anymore, and even if operator configured
		// BDOC Opener to support only TS or BES profiles via trust.yml config
		// file, it doesn't restrict user to sign BDOC containers with Mobile-ID
		// TM profile. Backend should fail with SignaturePolicyIdentifier
		{"MIDTM", []string{"O’CONNEŽ-ŠUSLIK,MARY ÄNN,11412090004"}, TM, true, 1, unsupportedProfileErrorMessageSupplier(TM)},
		{"MIDTM", []string{"O’CONNEŽ-ŠUSLIK,MARY ÄNN,11412090004"}, BES, true, 1, tmProfilePolicyErrorMessageSupplier(signatureIDS1)},
		{"MIDTM", []string{"O’CONNEŽ-ŠUSLIK,MARY ÄNN,11412090004"}, TS, true, 1, tmProfilePolicyErrorMessageSupplier(signatureIDS1)},

		// ID-card signature with no qualification are only valid for BES
		{"EIDBES", []string{"ŽAIKOVSKI,IGOR,37101010021"}, BES, false, 1, nil},
		{"EIDBES", []string{"ŽAIKOVSKI,IGOR,37101010021"}, TM, true, 1, unsupportedProfileErrorMessageSupplier(TM)},
		{"EIDBES", []string{"ŽAIKOVSKI,IGOR,37101010021"}, TS, true, 1, tmProfileTimestampErrorMessageSupplier(signatureIDS0)},

		// Containers with multiple signers and files.
		{
			"MultipleSigners",
			[]string{"ORAV,IVAN,30809010001", "ROPKA,KIVIVALVUR,32608320001"},
			TS, false, 1, nil,
		},
		{"MultipleFiles", []string{"ROPKA,KIVIVALVUR,32608320001"}, TS, false, 2, nil},

		// Containers with missing files.
		{"NoManifest", nil, BES, true, 1, manifestErrorMessageSupplier()},
		{"NoSignatures", nil, BES, true, 1, noSignatureErrorMessageSupplier()},
		{"NoFiles", []string{"ŽAIKOVSKI,IGOR,37101010021"}, BES, true, 0, noDataFilesErrorMessageSupplier()},

		// Containers with invalid OCSP response times.
		{"OCSPOld", nil, TS, true, 1, tsProfileOCSPOldErrorMessageSupplier(signatureIDS0)},
		{"OCSPDelayed", nil, TS, true, 1, tsProfileOCSPDelayedErrorMessageSupplier(signatureIDS0)},
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

			testOpener, err := loadTrustConf(test.profile)
			if err != nil {
				fail := test.failure()
				if !reflect.DeepEqual(fail, err) {
					errMsg := fmt.Sprintf(expectedEqualObjects, fail, err)
					t.Fatal(errMsg)
				}
				return
			}

			bdoc, err := testOpener.Open(file)
			switch {
			case test.expectedFailure && err != nil:
				fail := test.failure()
				if !reflect.DeepEqual(fail, err) {
					errMsg := fmt.Sprintf(expectedEqualObjects, fail, err)
					t.Fatal(errMsg)
				}
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
