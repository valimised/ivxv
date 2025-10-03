// Package internal provides utility functions to operate on ElGamal package, e.g.
// to provide internal factory functions etc.
// It is a private package, caller couldn't call any of the code piece found
// in this package.
package internal

import (
	"encoding/asn1"
	"fmt"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	"tivi.io/core/crypto/elgamal/distributed"
	"tivi.io/core/math/group"
)

// NewPublicKey is a factory function that returns new crypto public key by provided
// key object identifier. Key object identifier must be supported by crypto library
// otherwise error.
//
// Currently only ElGamal, Distributed ElGamal and ECDSA-with-SHAKE256 public keys are supported.
func NewPublicKey(algorithm asn1.ObjectIdentifier) (func(params crypto.AlgorithmIdentifierParameters, pub group.Element) (crypto.EncryptionKey, error), error) {
	switch algorithm.String() {
	case crypto.ElGamalEncryptionOID().String():
		return func(params crypto.AlgorithmIdentifierParameters, pub group.Element) (crypto.EncryptionKey, error) {
			parameters, ok := params.(*elgamal.Parameters)
			if !ok {
				return nil, fmt.Errorf("cannot cast to ElGamal public key parameters type")
			}

			return elgamal.NewPublicKey(parameters, pub), nil
		}, nil
	case crypto.DistributedElGamalEncryptionOID.String():
		return func(params crypto.AlgorithmIdentifierParameters, pub group.Element) (crypto.EncryptionKey, error) {
			parameters, ok := params.(*distributed.Parameters)
			if !ok {
				return nil, fmt.Errorf("cannot cast to distributed ElGamal public key parameters type")
			}

			return distributed.NewPublicKeyShare(parameters, pub), nil
		}, nil
	case crypto.EcdsaWithSHAKE256.String():
		// TODO:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown public key algorithm: %s", algorithm.String())
	}
}

// NewPrivateKey is a factory function that returns new crypto private key by provided
// key object identifier. Key object identifier must be supported by crypto library
// otherwise error.
//
// Currently only ElGamal, Distributed ElGamal and ECDSA-with-SHAKE256 private keys are supported.
func NewPrivateKey(algorithm asn1.ObjectIdentifier) (func(params crypto.AlgorithmIdentifierParameters, priv *group.Scalar) (any, error), error) {
	switch algorithm.String() {
	case crypto.ElGamalEncryptionOID().String():
		return func(params crypto.AlgorithmIdentifierParameters, priv *group.Scalar) (any, error) {
			parameters, ok := params.(*elgamal.Parameters)
			if !ok {
				return nil, fmt.Errorf("cannot cast to ElGamal private key parameters type")
			}

			key, err := elgamal.NewPrivateKey(parameters, priv)
			if err != nil {
				return nil, fmt.Errorf("failed to create new ElGamal private key: %v", err)
			}

			return key, nil
		}, nil
	case crypto.DistributedElGamalEncryptionOID.String():
		return func(params crypto.AlgorithmIdentifierParameters, priv *group.Scalar) (any, error) {
			parameters, ok := params.(*distributed.Parameters)
			if !ok {
				return nil, fmt.Errorf("cannot cast to distributed ElGamal private key parameters type")
			}

			key, err := distributed.NewPrivateKeyShareWithSecret(parameters, priv)
			if err != nil {
				return nil, fmt.Errorf("failed to create new distributed ElGamal private key: %v", err)
			}

			return key, nil
		}, nil
	case crypto.EcdsaWithSHAKE256.String():
		// TODO:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown private key algorithm: %s", algorithm.String())
	}
}
