package rpc

import (
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/status"
	api "ivxv.ee/sessionstatus/api/rpc"
)

// NewHandler returns an RPC handler that is ready to pass as rcvr
// parameter into rpc.NewServer().Register(rcvr), which in turn means
// that all rules applied to rcvr must also apply to a returned handler.
func NewHandler(status status.Status) *RPC {
	return &RPC{status: status}
}

// RPC is a handler to process microservices' session status requests.
type RPC struct {
	status status.Status
}

// SessionStatusRead is an RPC endpoint to provide an information
// about server.Header.SessionID of the client.
//
// Client should pass a request req, that includes req.server.Header.SessionID,
// then response resp will be an information includes previous RPC.Method being
// called with this server.Header.SessionID and Auth, which is a detailed
// authentication method being used ("id" for ID-card, "mid" for Mobile-ID,
// "sid" for Smart-ID, "wid" for Web eID). Lease is used to keep track of
// the TTL value of that particular server.Header.SessionID in a database.
func (r *RPC) SessionStatusRead(req api.StatusReadReq, resp *api.StatusReadResp) error {
	log.Log(req.Ctx, SessionStatusReadReq{Description: _SESSIONSTATUS_SESSSTATREADREQ})

	// s is either nil or *StatusReadResp.
	// nil s is only possible if error
	s, err := r.status.Read(req.Ctx, &req)
	if err != nil {
		return SessionStatusReadError{Err: err,
			Description: _SESSIONSTATUS_READ}
	}

	// s should be castable to *SessionStatusReadResp.
	// There is no way to s being non-castable
	// to *StatusReadResp, if so, then code has bugs
	readResp, err := castAnyToSessionStatusReadResp(s)
	if err != nil {
		return CastAnyToSessionStatusReadRespError{Err: err,
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATREADRESP}
	}

	resp.Header = readResp.Header
	resp.Caller = readResp.Caller
	resp.Auth = readResp.Auth
	resp.Lease = readResp.Lease

	log.Log(req.Ctx, SessionStatusReadResp{
		Caller:      resp.Caller,
		Auth:        resp.Auth,
		LeaseID:     resp.Lease,
		Description: _SESSIONSTATUS_SESSSTATREADRESP,
	})
	return nil
}

// SessionStatusUpdate is an RPC endpoint to update current information of a
// server.Header.SessionID client in the database. Information to be updated
// is passed in req. On success resp is returned, otherwise error.
func (r *RPC) SessionStatusUpdate(req api.StatusUpdateReq, resp *api.StatusUpdateResp) error {
	log.Log(req.Ctx, SessionStatusUpdateReq{
		Caller:      req.Caller,
		Auth:        req.Auth,
		LeaseID:     req.Lease,
		Description: _SESSIONSTATUS_SESSSTATUPDATEREQ,
	})

	// Update session status in a database
	err := r.status.Update(req.Ctx, &req)
	if err != nil {
		return SessionStatusUpdateError{
			SessionID:   req.SessionID,
			Caller:      req.Caller,
			Auth:        req.Auth,
			Err:         err,
			Description: _SESSIONSTATUS_PUT,
		}
	}

	resp.Ok = true

	log.Log(req.Ctx, SessionStatusUpdateResp{Success: resp.Ok,
		Description: _SESSIONSTATUS_SESSSTATUPDATERESP})
	return nil
}

// SessionStatusDelete is an RPC endpoint to delete current information of a
// server.Header.SessionID client in the database. Information to be deleted
// is passed in req. On success resp is returned, otherwise error.
func (r *RPC) SessionStatusDelete(req api.StatusDeleteReq, resp *api.StatusDeleteResp) error {
	log.Log(req.Ctx, SessionStatusDeleteReq{Description: _SESSIONSTATUS_SESSSTATDELETEREQ})

	// Delete session status in a database
	err := r.status.Delete(req.Ctx, &req)
	if err != nil {
		return SessionStatusDeleteError{Err: err,
			Description: _SESSIONSTATUS_DEL}
	}

	resp.Ok = true

	log.Log(req.Ctx, SessionStatusDeleteResp{Success: resp.Ok,
		Description: _SESSIONSTATUS_SESSSTATDELETERESP})
	return nil
}
