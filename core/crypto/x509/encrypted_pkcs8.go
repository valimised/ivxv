package x509

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"hash"
	"strconv"

	crypto2 "tivi.io/core/crypto"
	asn_1 "tivi.io/core/crypto/asn1"
	"tivi.io/core/crypto/x509/internal"
)

const (
	aes256KeySize = 32
)

var (
	// https://tools.ietf.org/html/rfc4231#section-3.1
	oidHMACWithSHA256 = asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 9}
	// https://tools.ietf.org/html/rfc3565#section-4.1
	oidAES256CBC = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 42}
)

func aes256CBCDecrypt(encrypted, key, iv []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key) // We already checked key length, ignore err.
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted = make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)
	return
}

// https://tools.ietf.org/html/rfc5208#section-6
type encryptedPrivateKeyInfo struct {
	EncryptionAlgorithm pkix.AlgorithmIdentifier
	EncryptedData       []byte
}

// https://tools.ietf.org/html/rfc2898#appendix-A.2
var oidPBKDF2 = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 12}

type pbkdf2Params struct {
	Salt           []byte
	IterationCount int
	KeyLength      int `asn1:"optional"`
	PRF            pkix.AlgorithmIdentifier
}

// https://tools.ietf.org/html/rfc2898#appendix-A.4
var oidPBES2 = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 13}

type pbes2Params struct {
	KeyDerivationFunc pkix.AlgorithmIdentifier
	EncryptionScheme  pkix.AlgorithmIdentifier
}

// EncryptedPKCS8Unmarshal decrypts encrypted with password and then passes the
// result to ParsePKCS8. The encrypted data can be encoded as either DER or PEM
// with the block type ENCRYPTED PRIVATE KEY.
//
// Currently only keys encrypted using PBES2 with PBKDF2 as the key derivation
// function and AES-256-CBC as the encryption scheme are supported (see RFC
// 2898: PKCS #5 at https://tools.ietf.org/html/rfc2898).
func (pkcs8 *UnmarshallerPKCS8[T]) EncryptedPKCS8Unmarshal(encrypted, password []byte) (params crypto2.AlgorithmIdentifierParameters, key T, err error) {
	// Unmarshal PEM to DER bytes
	der, err := internal.Pem2Der(encrypted, internal.EncryptedPKCS8Header)
	if err != nil {
		return nil, key, fmt.Errorf("PEM to DER conversion failed: %v", err)
	}

	// Unlock DER bytes using provided password
	var encryptedKeyInfo encryptedPrivateKeyInfo
	var rest []byte
	if rest, err = asn1.Unmarshal(der, &encryptedKeyInfo); err != nil {
		return nil, key, fmt.Errorf("failed to ASN.1 unmarshal encrypted private key")
	} else if len(rest) > 0 {
		return nil, key, fmt.Errorf("ASN.1 unmarshalled encrypted private key left trailing bytes")
	}

	// Make sure the key is encrypted using PBES2 with PBKDF2 as the key
	// derivation function and AES-256-CBC as the encryption scheme.
	encryptionOID := encryptedKeyInfo.EncryptionAlgorithm.Algorithm
	if !encryptionOID.Equal(oidPBES2) {
		return nil, key, fmt.Errorf("unsupported encrypted private key alogirthm: %s", encryptionOID.String())
	}

	var keyParams pbes2Params
	if rest, err = asn1.Unmarshal(encryptedKeyInfo.EncryptionAlgorithm.Parameters.FullBytes, &params); err != nil {
		return nil, key, fmt.Errorf("failed to ASN.1 unmarshal encrypted private key parameters")
	} else if len(rest) > 0 {
		return nil, key, fmt.Errorf("ASN.1 unmarshalled encrypted private key parameters left trailing bytes")
	}

	keyDerivationOID := keyParams.KeyDerivationFunc.Algorithm
	if !keyDerivationOID.Equal(oidPBKDF2) {
		return nil, key, fmt.Errorf("unsupported encrytion private key derivation alogirthm: %s", keyDerivationOID.String())
	}

	encryptionSchemeOID := keyParams.EncryptionScheme.Algorithm
	if !encryptionSchemeOID.Equal(oidAES256CBC) {
		return nil, key, fmt.Errorf("unsupported encrytion private key scheme: %s", keyDerivationOID.String())
	}

	// Parse PBKDF2 params and ensure IterationCount is positive, KeyLength
	// (if present) matches AES-256 key size, and the PRF is HMAC-SHA256.
	var derivationParams pbkdf2Params
	if rest, err = asn1.Unmarshal(keyParams.KeyDerivationFunc.Parameters.FullBytes, &derivationParams); err != nil {
		return nil, key, fmt.Errorf("failed to ASN.1 unmarshal encrypted private key PBKDF2 parameters")
	} else if len(rest) > 0 {
		return nil, key, fmt.Errorf("ASN.1 unmarshalled encrypted private key PBKDF2 parameters left trailing bytes")
	}

	if count := derivationParams.IterationCount; count <= 0 {
		return nil, key, fmt.Errorf("encrypted private key derivation count <= 0")
	}

	if keyLength := derivationParams.KeyLength; keyLength == 0 {
		derivationParams.KeyLength = aes256KeySize
	} else if keyLength != aes256KeySize {
		return nil, key, fmt.Errorf("expected encrypted private key length %s, but got %s", strconv.Itoa(aes256KeySize), strconv.Itoa(keyLength))
	}

	hmacOID := derivationParams.PRF.Algorithm
	if !hmacOID.Equal(oidHMACWithSHA256) {
		return nil, key, fmt.Errorf("unsupported encrytion private key PRF alogirthm: %s", keyDerivationOID.String())
	}
	h := sha256.New

	// Parse AES params and ensure IV length matches AES block size and
	// length of the encrypted data is a multiple of it.
	var encryptionParams []byte // The initialization vector.
	if rest, err = asn1.Unmarshal(keyParams.EncryptionScheme.Parameters.FullBytes, &encryptionParams); err != nil {
		return nil, key, fmt.Errorf("failed to ASN.1 unmarshal encrypted private key scheme parameters")
	} else if len(rest) > 0 {
		return nil, key, fmt.Errorf("ASN.1 unmarshalled encrypted private key scheme parameters left trailing bytes")
	}
	if len(encryptionParams) != aes.BlockSize {
		return nil, key, fmt.Errorf("invalid encrypted private key scheme parameters bytes length %s", strconv.Itoa(len(encryptionParams)))
	}
	if length := len(encryptedKeyInfo.EncryptedData); length%aes.BlockSize != 0 {
		return nil, key, fmt.Errorf("invalid encrypted private key data bytes length %s", strconv.Itoa(length))
	}

	// Derive the encryption key and decrypt the data.
	aesKey := pbkdf2(h, password, derivationParams.Salt,
		derivationParams.IterationCount, derivationParams.KeyLength)
	keyInfo := aes256CBCDecrypt(encryptedKeyInfo.EncryptedData, aesKey, encryptionParams)

	// Strip the PKCS #5/PKCS #7 padding from keyInfo
	// (https://tools.ietf.org/html/rfc2315#section-10.3) and parse it.
	// Don't worry about creating a padding oracle, because this is used on
	// local files anyway.
	paddingSize := int(keyInfo[len(keyInfo)-1])
	if paddingSize <= 0 || paddingSize > aes.BlockSize {
		return nil, key, fmt.Errorf("invalid encrypted private key data padding size: %s, while max. is %s", strconv.Itoa(paddingSize), strconv.Itoa(aes.BlockSize))
	}
	padding := keyInfo[len(keyInfo)-paddingSize:]
	if !bytes.Equal(padding, bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)) {
		return nil, key, fmt.Errorf("invalid encrypted private key data padding: %s", asn_1.HexN(padding, aes.BlockSize))
	}
	keyInfo = keyInfo[:len(keyInfo)-paddingSize]

	// Once DER bytes are unlocked, unmarshal it to crypto private key
	params, key, err = pkcs8.Unmarshal(keyInfo)
	if err != nil {
		return nil, key, fmt.Errorf("failed to ASN.1 unmarshal unlocked private key: %v", err)
	}
	return
}

// https://tools.ietf.org/html/rfc2898#section-5.2
func pbkdf2(h func() hash.Hash, password, salt []byte, iter, keylen int) (key []byte) {
	prf := hmac.New(h, password)
	hlen := prf.Size()
	l := (keylen + hlen - 1) / hlen // Number of blocks in keylen.

	ibytes := make([]byte, 4)
	for i := 1; i <= l; i++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(ibytes, uint32(i))
		prf.Write(ibytes)
		block := prf.Sum(nil) // U_1.
		key = append(key, block...)

		for c := 2; c <= iter; c++ { // Repeat the PRF iter times.
			prf.Reset()
			prf.Write(block)
			block = prf.Sum(nil) // U_c.
			for n, u := range block {
				key[(i-1)*hlen+n] ^= u // XOR U_c into T_i.
			}
		}
	}
	return key[:keylen]
}
