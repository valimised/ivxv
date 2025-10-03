package rpc

import (
	"context"
	"encoding/base64"

	"ivxv.ee/common/collector/status"
	"ivxv.ee/common/collector/storage"
	api "ivxv.ee/sessionstatus/api/rpc"
)

const (
	sessionIDPrefix             = "/session"
	separator                   = "\x1f"
	statusReadRespDBRecordCount = 2
)

type client struct {
	repository storage.SessionStatusRepository
}

// NewStatusRepository initializes storage client r for a session
// status server. TTL value can be used for assigning expiration time
// for a key in a database.
func NewStatusRepository(r storage.SessionStatusRepository) status.Status {
	return &client{repository: r}
}

func (c *client) Read(ctx context.Context, data interface{}) (interface{}, error) {
	// Any data that is passed here, must cast to *SessionStatusReadReq
	req, err := castAnyToSessionStatusReadReq(data)
	if err != nil {
		return nil, CastAnyToSessionStatusReadReqError{Err: err,
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSREADREQ}
	}

	key := toSessionStorageKey(req.Header.SessionID)

	// GET database query.
	// val can be either nil or []byte.
	// nil means an error or absence of a value
	val, lease, err := c.repository.GetWithLease(ctx, key)
	if err != nil {
		return nil, GetWithLeaseError{
			Key:         key,
			Err:         err,
			Description: _SESSIONSTATUS_READ,
		}
	}

	// nil means that val doesn't exist in a database,
	// for a caller it means either req.Header.SessionID key
	// is a brand-new session ID or existing but tampered session ID.
	// Decision should be made by a caller!
	if val == nil {
		return &api.StatusReadResp{Header: req.Header}, nil
	}

	// Verify val correctness
	array, err := parseSessionStatus(val)
	if err != nil {
		return nil, ParseSessionStatusError{
			Value:       val,
			Err:         err,
			Description: _SESSIONSTATUS_RESP,
		}
	}

	return &api.StatusReadResp{
		Header: req.Header,
		Caller: array[0],
		Auth:   array[1],
		Lease:  lease,
	}, nil
}

func (c *client) Update(ctx context.Context, data interface{}) error {
	// Any data that is passed here, must cast to *SessionStatusUpdateReq
	req, err := castAnyToSessionStatusUpdateReq(data)
	if err != nil {
		return CastAnyToSessionStatusUpdateReqError{Err: err,
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSUPDATEDREQ}
	}

	// Value should be stored in a database as
	// base64(Caller\x1FAuth)
	concat := req.Caller + separator + req.Auth
	val := []byte(base64.StdEncoding.EncodeToString([]byte(concat)))

	// TTL value is int and measured in seconds!
	ttl := &storage.PutOpOptionWithTTL{TTL: req.TTL, LeaseID: req.Lease}

	key := toSessionStorageKey(req.Header.SessionID)

	// PUT query to a database, where key will be deleted after ttl amount
	// of seconds
	err = c.repository.PutForceWithOpts(ctx, key, val, ttl)
	if err != nil {
		return PutForceWithOptsError{
			Key:         key,
			Value:       val,
			Err:         err,
			Description: _SESSIONSTATUS_PUT,
		}
	}

	return nil
}

func (c *client) Delete(ctx context.Context, data interface{}) error {
	// Any data that is passed here, must cast to *SessionStatusDeleteReq
	req, err := castAnyToSessionStatusDeleteReq(data)
	if err != nil {
		return CastAnyToSessionStatusDeleteReqError{Err: err,
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSDELETEDREQ}
	}

	key := toSessionStorageKey(req.Header.SessionID)

	// DELETE query to a database
	err = c.repository.Delete(ctx, key)
	if err != nil {
		return DeleteError{
			Key:         key,
			Err:         err,
			Description: _SESSIONSTATUS_DEL,
		}
	}

	return nil
}
