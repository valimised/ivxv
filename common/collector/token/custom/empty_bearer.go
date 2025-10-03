package custom

import (
	"encoding/base64"
	"encoding/json"
)

// Header is not implemented.
func (t *fromEmpty) Header() (string, error) {
	return "", nil
}

// Payload will serialize t into []byte and base64([]byte).
func (t *fromEmpty) Payload() (string, error) {
	// Serialize t
	b, err := json.Marshal(t)
	if err != nil {
		return "", JSONMarshalError{Err: err,
			Description: _BEARER_PAYLOAD}
	}

	// Base64 encode marshalled bearer token.
	// Note, that no signature was added to the bearer token
	t.Data = base64.StdEncoding.EncodeToString(b)
	return t.Data, nil
}

// Signature is not implemented.
func (t *fromEmpty) Signature() ([]byte, error) {
	return nil, nil
}

// Verify is not implemented.
func (t *fromEmpty) Verify() error {
	return nil
}
