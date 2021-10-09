package mid

import (
	"context"
)

// MobileSignHash starts a Mobile-ID signing session to sigh hash.
func (c *Client) MobileSignHash(ctx context.Context, idCode, phone string, hash []byte,
	hashType string) (sesscode string, err error) {

	sesscode, err = c.startSession(ctx, sessSign, idCode, phone, hash, hashType)
	if err != nil {
		err = MobileSignHashError{Err: err}
		return
	}

	return
}

// GetMobileSignHashStatus queries the status of a Mobile-ID signing session.
// If err is nil and signature is empty, then the transaction is still
// outstanding.
func (c *Client) GetMobileSignHashStatus(ctx context.Context, sesscode string) (
	algorithm string, signature []byte, err error) {

	algorithm, signature, _, err = c.getSessionStatus(ctx, sessSign, sesscode)
	if err != nil {
		err = GetMobileSignHashStatusError{Err: err}
		return
	}
	return
}
