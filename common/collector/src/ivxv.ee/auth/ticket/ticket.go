/*
Package ticket implements voter authentication using cryptographic cookies.

An authentication service verifies the identity of the voter and then issues
encrypted and authenticated cookies called tickets. Other services that trust
the authentication service, i.e., share the encryption key with it, need only
verify this ticket and not perform the entire authentication protocol
themselves.

To ensure that a ticket can only be used to store a vote once, it contains the
vote identifier that will be assigned to the vote: once a vote with this
identifier exists, the ticket cannot be used to store a new vote.
*/
package ticket // import "ivxv.ee/auth/ticket"

import (
	"context"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/asn1"
	"io/ioutil"

	"ivxv.ee/auth"
	"ivxv.ee/cookie"
	"ivxv.ee/yaml"
)

func init() {
	auth.Register(auth.Ticket, func(_ yaml.Node) (auth.Verifier, error) {
		// Wrap NewFromSystem to convert *T to auth.Verifier.
		return NewFromSystem()
	})
}

// Conf is the ticket authentication method configuration.
type Conf struct {
	Key cookie.Key
}

// T is a ticket manager which can issue and verify authentication tickets.
type T struct {
	cookie *cookie.C
}

// New creates a new ticket manager with the provided cookie key.
func New(key cookie.Key) (t *T, err error) {
	c, err := cookie.New(key)
	if err != nil {
		return nil, NewCookieError{Err: err}
	}
	return &T{cookie: c}, nil
}

// NewFromSystem creates a new ticket manager with the key read from the
// filesystem.
func NewFromSystem() (t *T, err error) {
	key, err := ioutil.ReadFile("/var/lib/ivxv/service/ticket.key")
	if err != nil {
		return nil, ReadKeyError{Err: err}
	}
	if t, err = New(key); err != nil {
		return nil, NewTicketError{Err: err}
	}
	return
}

type tt struct {
	Subject pkix.RDNSequence
	VoteID  []byte
}

// Create issues a new authentication ticket for the subject.
func (t *T) Create(subject pkix.Name) (ticket []byte, err error) {
	vid := make([]byte, 16)
	if _, err = rand.Read(vid); err != nil {
		return nil, GenerateVoteIDError{Err: err}
	}
	subject.ExtraNames = subject.Names // Also marshal unrecognized names.
	plain, err := asn1.Marshal(tt{Subject: subject.ToRDNSequence(), VoteID: vid})
	if err != nil {
		return nil, MarshalTicketError{Err: err}
	}
	return t.cookie.Create(plain), nil
}

// Verify implements the ivxv.ee/auth.Verifier interface. The token must be a
// ticket issued with the same cookie key as t was configured with.
func (t *T) Verify(_ context.Context, token []byte) (name *pkix.Name, err error) {
	ticket, err := t.open(token)
	if err != nil {
		return nil, VerifyOpenError{Err: err}
	}
	name = new(pkix.Name)
	name.FillFromRDNSequence(&ticket.Subject)
	return
}

// VoteIdentifier implements the ivxv.ee/auth.VoteIdentifier interface. The
// token must be a ticket issued with the same cookie key as t was configured
// with.
func (t *T) VoteIdentifier(token []byte) (voteID []byte, err error) {
	ticket, err := t.open(token)
	if err != nil {
		return nil, VoteIdentifierOpenError{Err: err}
	}
	return ticket.VoteID, nil
}

func (t *T) open(token []byte) (ticket tt, err error) {
	plain, err := t.cookie.Open(token)
	if err != nil {
		return ticket, OpenTicketError{Err: err}
	}
	rest, err := asn1.Unmarshal(plain, &ticket)
	if err != nil {
		return ticket, UnmarshalTicketError{Err: err}
	}
	if len(rest) > 0 {
		return ticket, TrailingDataError{Rest: rest}
	}
	return
}
