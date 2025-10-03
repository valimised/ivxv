package cades

import (
	"crypto/x509"
	"encoding/asn1"
	"fmt"

	"tivi.io/core/crypto/cms"
	"tivi.io/core/crypto/timestamp"
)

const idSignatureTimestamp = "1.2.840.113549.1.9.16.2.14"

var otherUnsignedAttrs = make(map[string]cms.AttrToVerify, 0)

func CAdESTVerifySignedData(signedData []byte, certs []*x509.Certificate, detachedData, nonce []byte) error {
	otherUnsignedAttrs[idSignatureTimestamp] = cms.AttrToVerify{
		AttributeID: idSignatureTimestamp,
		VerifyCallback: func(b []byte) error {
			var ctInfo cms.ContentInfo
			rest, err := asn1.Unmarshal(signedData, &ctInfo)
			if err != nil {
				return fmt.Errorf("unmarshal content info: %s", err)
			}
			if len(rest) > 0 {
				return fmt.Errorf("unmarshal content info excess bytes: %d", len(rest))
			}
			for _, signer := range ctInfo.Content.SignerInfos {
				_, err := timestamp.VerifyTimestamp(b, signer.Signature, nonce, timestamp.Conf{Signers: certs})
				if err != nil {
					return fmt.Errorf("verify timestamp: %w", err)
				}
			}
			return nil
		},
	}
	return cadesBVerifySignedData(signedData, certs, detachedData, map[string]cms.AttrToVerify{}, otherUnsignedAttrs)
}
