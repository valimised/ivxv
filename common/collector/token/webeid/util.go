package webeid

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"hash"
	"regexp"
)

// formatRegex checks whether Format field of a Web eID auth token is correct,
// according to the regex rules.
//
// If correct, then major release of a Web eID token is returned.
func formatRegex(format string) (string, error) {
	// format is expected to be web-eid:1.0
	re := regexp.MustCompile(formatRegexp)
	if !re.MatchString(format) {
		return "", FormatRegexError{Format: format,
			Description: _WEBEID_REGEX}
	}

	// [web-eid:1.0, 1, 0]
	return re.FindStringSubmatch(format)[1], nil
}

// formatMajorRelease ensures that backend and Web eID support
// the same major release, otherwise incompatible.
func formatMajorRelease(release string) error {
	if release != majorRelease {
		return FormatMajorReleaseError{
			Expected:    majorRelease,
			Got:         release,
			Description: _WEBEID_VER,
		}
	}
	return nil
}

// hashIt hashes it using hash algorithm algo.
func hashIt(it []byte, algo string) ([]byte, error) {
	// Web eID algo is a custom formatted string, so parse it first
	h, err := hashAlgorithm(algo)

	if err != nil {
		return nil, HashDataError{Err: err,
			Description: WEBEID_ALG}
	}

	// Hash it with algo hashing algorithm
	h.Write(it)
	return h.Sum(nil), nil
}

// algorithmRegex checks regex over Web eID auth token Algorithm field.
func algorithmRegex(algo string) (string, error) {
	re := regexp.MustCompile(algorithm)
	if !re.MatchString(algo) {
		return "", AlgorithmRegexError{Algorithm: algo,
			Description: WEBEID_ALG_REGEX}
	}

	// For example "ES512" is [ES512, ES, 512]
	return re.FindStringSubmatch(algo)[2], nil
}

// hashAlgorithm parses Web eID auth token Algorithm field, extract hash
// algorithm from it and returns it as a hash.Hash.
func hashAlgorithm(algo string) (hash.Hash, error) {
	// Check regex over Algorithm first
	sigHash, err := algorithmRegex(algo)
	if err != nil {
		return nil, MalformedAlgorithmError{Err: err,
			Description: WEBEID_ALG_REGEX}
	}

	switch sigHash {
	case "256":
		return sha256.New(), nil
	case "384":
		return sha512.New384(), nil
	case "512":
		return sha512.New(), nil
	default:
		return nil, UnsupportedAlgorithmError{Algorithm: algo,
			Description: WEBEID_ALG}
	}
}

//nolint:lll
/*
NB! IMPORTANT

THIS CONFRONTS WITH THE OFFICIAL DOCUMENTATION OF WEB EID,
SINCE CPP LIBRARY LIBELECTRONIC-ID FOR ID CARD SIGNING
https://github.com/web-eid/libelectronic-id/blob/0b5e58cf8141df49a1fd14b5e0c588bc6bae410a/src/electronic-ids/pcsc/EstEIDIDEMIA.cpp#L53
TRUNCATES A HASH THAT IS LONGER THAN 48 BYTES (SHA384) AND
APPENDS NULL BYTES IF HASH IS SHORTER THAN 32 BYTES (SHA256).

THEREFORE WE HAVE TO OVERCOME THAT BY HASHING RESULTING HASH AGAIN WITH SHA384.
*/
func anySignatureAlgorithmToSHA384(sigAlgo x509.SignatureAlgorithm) x509.SignatureAlgorithm {
	switch sigAlgo {
	// 256 --> 384
	case x509.ECDSAWithSHA256:
		return x509.ECDSAWithSHA384
	case x509.SHA256WithRSAPSS:
		return x509.SHA384WithRSAPSS
	case x509.SHA256WithRSA:
		return x509.SHA384WithRSA
	// 384 --> 384
	case x509.ECDSAWithSHA384, x509.SHA384WithRSAPSS, x509.SHA384WithRSA:
		return sigAlgo
	// 512 --> 384
	case x509.ECDSAWithSHA512:
		return x509.ECDSAWithSHA384
	case x509.SHA512WithRSAPSS:
		return x509.SHA384WithRSAPSS
	case x509.SHA512WithRSA:
		return x509.SHA384WithRSA
	default:
		return x509.ECDSAWithSHA384
	}
}
