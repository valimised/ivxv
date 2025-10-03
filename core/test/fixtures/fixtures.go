package fixtures

// KeyPairProof proves to auditor that public key used in vote encryption is a pair of
// private key used in vote decryption.
type KeyPairProof struct {
	Proof []byte
}

// CiphertextProof proves to auditor that encrypted vote (ciphertext) produced by public key,
// private key used in vote decryption.
type CiphertextProof struct {
	Ciphertext []byte
	Decryption []byte
	Proof      []byte
}
