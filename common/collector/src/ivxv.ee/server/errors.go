package server

import "errors"

// Server protocol errors. These are the possible errors that can be returned
// from a JSON-RPC call.
var (
	// General errors.
	ErrBadRequest      = errors.New("BAD_REQUEST")
	ErrCertificate     = errors.New("BAD_CERTIFICATE")
	ErrIneligible      = errors.New("INELIGIBLE_VOTER")
	ErrInternal        = errors.New("INTERNAL_SERVER_ERROR")
	ErrTooYoung        = errors.New("VOTER_TOO_YOUNG")
	ErrUnauthenticated = errors.New("UNAUTHENTICATED")
	ErrVotingEnd       = errors.New("VOTING_END")

	// Mobile ID errors.
	ErrMIDAbsent      = errors.New("MID_ABSENT")
	ErrMIDCanceled    = errors.New("MID_CANCELED")
	ErrMIDCertificate = errors.New("MID_BAD_CERTIFICATE")
	ErrMIDExpired     = errors.New("MID_EXPIRED")
	ErrMIDGeneral     = errors.New("MID_GENERAL")
	ErrMIDNotUser     = errors.New("MID_NOT_USER")
	ErrMIDOperator    = errors.New("MID_OPERATOR")

	// Smart ID errors.
	ErrSmartIDCanceled     = errors.New("SMARTID_CANCELED")
	ErrSmartIDCertificate  = errors.New("SMARTID_BAD_CERTIFICATE")
	ErrSmartIDExpired      = errors.New("SMARTID_EXPIRED")
	ErrSmartIDGeneral      = errors.New("SMARTID_GENERAL")
	ErrSmartIDVerification = errors.New("SMARTID_VERIFICATION")
	ErrSmartIDAccount      = errors.New("SMARTID_ACCOUNT")

	// Vote errors.
	ErrIdentityMismatch = errors.New("IDENTITY_MISMATCH")
	ErrOutdatedChoices  = errors.New("OUTDATED_CHOICES")
	ErrVotingRateLimit  = errors.New("RATE_LIMIT_EXCEEDED")
)
