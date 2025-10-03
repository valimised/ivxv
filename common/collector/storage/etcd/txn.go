package etcd

import (
	"context"
	"reflect"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"

	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/storage"
)

const expectedCastForTxnOp = "*txnOp"

// txnOp is an implementation for a transaction.
type txnOp struct {
	cmps   []clientv3.Cmp
	thens  []clientv3.Op
	elses  []clientv3.Op
	readyc chan bool
	errorc chan error
}

// Begin a transaction with lazy initialization.
func (c *client) Begin(ctx context.Context) (storage.TxnOp, error) {
	log.Debug(ctx, BeginTxn{Description: _ETCD_TXN_BEGIN})

	return &txnOp{
		readyc: make(chan bool, 1),
		errorc: make(chan error, 1),
	}, nil
}

// Ready sends bool over the ready chan and waits for an error
// on the error chan.
func (c *txnOp) Ready(ctx context.Context) error {
	// Send true to readyc
	c.readyc <- true

	// Once true is sent to readyc, start listening on errorc and ctx.Done.
	// It is also possible that by this time there is already an answer on
	// errorc. This is possible since errorc is a buffered channel.
	select {
	case err := <-c.errorc:
		log.Debug(ctx, ReadyReceivedOnErrorChannel{Err: err, Description: _ETCD_TXN_FAIL})
		return err
	case <-ctx.Done():
		return log.Alert(ReadyContextCancelled{Description: _ETCD_TXN_STOP})
	}
}

// AutoCommit waits for a bool on the ready chan and once received,
// attempts to Commit and send an error over the error chan.
func (c *client) AutoCommit(ctx context.Context, op storage.TxnOp) {
	// Don't check casting success here, instead work as a mediator and
	// just send op as it is to the Commit method
	unit := op.(*txnOp)

	// Move this function to the background (daemon).
	// Once select {...} block is passed - finish the daemon
	go func() {
		defer close(unit.readyc)
		defer close(unit.errorc)

		// Listen on readyc. It is possible that by this time there is already
		// an answer present. It is possible, since readyc is buffered channel
		select {
		case <-unit.readyc:
			unit.errorc <- c.Commit(ctx, op)
			log.Debug(ctx, AutoCommitSentOnErrorc{Description: _ETCD_TXN_STOP})
		case <-ctx.Done():
			// AutoCommit doesn't act on context deadline exceeded error,
			// instead it just exits the function
			log.Debug(ctx, AutoCommitContextCancelled{Description: _ETCD_CTX_DONE})
		}

		log.Debug(ctx, AutoCommitClose{Description: _ETCD_AUTOCOMMIT_UNSET})
		// Close readyc and errorc channels
	}()
}

// Commit a transaction.
func (c *client) Commit(ctx context.Context, op storage.TxnOp) error {
	// Get etcd key-value store. It is a same as an etcd client,
	// just with restricted operations, mostly CRUD-only
	kv, err := c.kv(ctx)
	if err != nil {
		return log.Alert(BeginKVError{Err: err,
			Description: _ETCD_CLIENT_N_CONN})
	}

	// Must cast to *txnOp
	unit, ok := op.(*txnOp)
	if !ok {
		return log.Alert(CommitCastToTxnOpError{
			Expected:    expectedCastForTxnOp,
			Got:         reflect.TypeOf(op),
			Description: _ETCD_CAST_REQ,
		})
	}

	log.Debug(ctx, CommitRequest{Description: _ETCD_TXN_START})

	// Include all If, Then, Else OpOptions to the Transaction COMMIT operation.
	// "All or nothing!" means that etcd either applies all changes or none of them
	resp, err := c.doRetry(ctx, func(ctx context.Context) (*clientv3.TxnResponse, error) {
		return kv.Txn(ctx).
			If(unit.cmps...).
			Then(unit.thens...).
			Else(unit.elses...).
			Commit()
	})
	if err != nil {
		return log.Alert(CommitError{Err: err,
			Description: _ETCD_ROLLBACK})
	}

	log.Debug(ctx, CommitResponse{
		Response:    resp,
		Success:     resp.Succeeded,
		Description: _ETCD_COMMIT,
	})

	if !resp.Succeeded {
		return storage.UnexpectedValueError{
			Err: CommitResponseError{Response: resp, Success: resp.Succeeded,
				Description: _ETCD_CMP}}
	}
	return nil
}

// If adds cmp into a list of predicates. List of predicates is said to be
// true if and only if all predicates are true, otherwise false.
func (c *txnOp) If(cmp clientv3.Cmp) {
	c.cmps = append(c.cmps, cmp)
}

// Then adds then into a list of operations, that are executed if list of
// predicates is true.
func (c *txnOp) Then(then clientv3.Op) {
	c.thens = append(c.thens, then)
}

// Else adds elze into a list of operations, that are executed if list of
// predicates is false.
func (c *txnOp) Else(elze clientv3.Op) {
	c.elses = append(c.elses, elze)
}

// Put adds key-value into a buffer to be further committed.
// Note, that only if each key is unique, value is added,
// due to KeyMissing policy.
func (c *txnOp) Put(key string, value []byte) {
	c.If(clientv3util.KeyMissing(key))
	c.Then(clientv3.OpPut(key, string(value)))
	c.Else(clientv3.OpGet(key))
}

// PutAll calls Put on each req in reqs.
func (c *txnOp) PutAll(reqs ...*storage.PutAllRequest) {
	for _, req := range reqs {
		c.Put(req.Key, req.Value)
	}
}

// PutForce adds key-value into a buffer without utilizing If and Else.
func (c *txnOp) PutForce(key string, value []byte) {
	c.Then(clientv3.OpPut(key, string(value)))
}

// CAS adds key-value into a buffer to be further committed.
// Note, that only if key's current value equals to old,
// then new value is added to key.
func (c *txnOp) CAS(key string, old, new []byte) { //nolint:revive
	c.If(clientv3.Compare(clientv3.Value(key), "=", string(old)))
	c.Then(clientv3.OpPut(key, string(new)))
	c.Else(clientv3.OpGet(key))
}
