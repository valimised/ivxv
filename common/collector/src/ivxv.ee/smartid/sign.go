package smartid

import (
	"context"
)

// SignHash starts a Smart-ID signing session to sigh hash.
func (c *Client) SignHash(ctx context.Context, documentno string, hash []byte,
	hashType string) (sesscode string, err error) {

	sesscode, err = c.startSession(ctx, sessSign, documentno, hash, hashType)
	if err != nil {
		err = SignHashError{Err: err}
		return
	}

	return
}

// GetSignHashStatus queries the status of a Smart-ID signing session.
// If err is nil and signature is empty, then the transaction is still
// outstanding.
func (c *Client) GetSignHashStatus(ctx context.Context, sesscode string) (
	algorithm string, signature []byte, err error) {

	_, algorithm, signature, _, err = c.getSessionStatus(ctx, sesscode)
	if err != nil {
		err = GetSignHashStatusError{Err: err}
		return
	}
	return
}
