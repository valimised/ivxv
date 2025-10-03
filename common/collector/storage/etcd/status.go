package etcd

import (
	"context"
	"strconv"

	clientv3 "go.etcd.io/etcd/client/v3"

	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/storage"
)

const (
	emptyLeaseID = ""
	base10       = 10
)

func (c *client) GetWithLease(ctx context.Context, key string) ([]byte, string, error) {
	// Get etcd key-value store. It is a same as an etcd client,
	// just with restricted operations, mostly CRUD-only
	kv, err := c.kv(ctx)
	if err != nil {
		return nil, emptyLeaseID, log.Alert(GetWithLeaseKVError{Err: err,
			Description: _ETCD_CLIENT_N_CONN})
	}

	log.Debug(ctx, GetWithLeaseRequest{Key: key, Description: _ETCD_GET_W_LEASE})

	// If etcd doesn't respond, then don't hung longer than c.optime
	ctx, cancel := context.WithTimeout(ctx, c.optime)
	defer cancel()

	// Send GET query
	resp, err := kv.Get(ctx, key)
	if err != nil {
		return nil, emptyLeaseID, log.Alert(GetWithLeaseError{
			Key:         key,
			Err:         err,
			Description: _ETCD_GET_W_LEASE_DB_FAIL,
		})
	}

	// There is no such a key in etcd
	if resp.Kvs == nil {
		log.Debug(ctx, GetWithLeaseEmptyResponse{Key: key,
			Description: _ETCD_GET_NONE})
		return nil, emptyLeaseID, nil
	}

	// Convert from int64 to UTF-8 string.
	// This helps to prevent any numerical casting issues,
	// it is probably safer to hold numerical values in a UTF-8 string
	leaseID := strconv.FormatInt(resp.Kvs[0].Lease, base10)

	log.Debug(ctx, GetWithLeaseResponse{
		Response:    resp,
		Key:         key,
		Value:       resp.Kvs[0].Value,
		LeaseID:     leaseID,
		Description: _ETCD_GET_W_LEASE_COMMITTED,
	})

	return resp.Kvs[0].Value, leaseID, err
}

func (c *client) PutForceWithOpts(ctx context.Context, key string, value []byte,
	opts interface{}) (err error) {
	// Currently only TTL option is expected as opts,
	// however nothing stops you from switch opts.(type) {...}
	putOptsWithTTL, ok := opts.(*storage.PutOpOptionWithTTL)
	if !ok {
		return log.Alert(PutForceWithOptsCastToPutOpOptionsWithTTLError{
			Key:         key,
			Err:         err,
			Description: _ETCD_PUT_OPTS,
		})
	}

	// Get etcd client key-value store
	kv, err := c.kv(ctx)
	if err != nil {
		return log.Alert(PutForceWithOptsKVError{Err: err,
			Description: _ETCD_GET_NONE})
	}

	// Ensure that LeaseID is not "", otherwise Go will panic at strconv.ParseInt
	if putOptsWithTTL.LeaseID == "" {
		log.Debug(ctx, PutForceWithOptsEmptyLeaseID{
			Key:         key,
			Value:       value,
			LeaseID:     putOptsWithTTL.LeaseID,
			TTL:         putOptsWithTTL.TTL,
			Description: _ETCD_PUT_LEASE,
		})

		putOptsWithTTL.LeaseID = "0"
	}

	// LeaseID is at least "0"
	leaseID, err := strconv.ParseInt(putOptsWithTTL.LeaseID, 10, 64)
	if err != nil {
		return log.Alert(PutForceWithOptsConvertLeaseIDToInt64Error{
			LeaseID:     putOptsWithTTL.LeaseID,
			Description: _ETCD_PUT_LEASE_INT,
		})
	}

	// Ensure that TTL is not "", otherwise Go will panic at strconv.ParseInt
	if putOptsWithTTL.TTL == "" {
		log.Debug(ctx, PutForceWithOptsEmptyTTL{
			Key:         key,
			Value:       value,
			LeaseID:     putOptsWithTTL.LeaseID,
			TTL:         putOptsWithTTL.TTL,
			Description: _ETCD_PUT_LEASE_TTL,
		})

		putOptsWithTTL.TTL = "0"
	}

	// TTL is "0" if not set, which means that value will disappear immediately
	ttl, err := strconv.ParseInt(putOptsWithTTL.TTL, 10, 64)
	if err != nil {
		return log.Alert(PutForceWithOptsConvertTTLToInt64Error{
			LeaseID:     putOptsWithTTL.LeaseID,
			Description: _ETCD_PUT_LEASE_TTL_INT,
		})
	}

	log.Debug(ctx, PutForceWithOptsRequest{
		Key:         key,
		Value:       value,
		LeaseID:     putOptsWithTTL.LeaseID,
		TTL:         ttl,
		Description: _ETCD_PUT,
	})

	resp, err := c.putForceWithOpts(ctx, kv, storage.PutAllRequestWithTTL{
		Key:     key,
		Value:   value,
		TTL:     ttl,
		LeaseID: leaseID,
	})
	if err != nil || !resp.Succeeded {
		return log.Alert(PutForceWithOptsError{
			Key:         key,
			Err:         err,
			Description: _ETCD_PUT_DB_FAIL,
		})
	}

	log.Debug(ctx, PutForceWithOptsResponse{
		Response:    resp,
		Success:     resp.Succeeded,
		Description: _ETCD_PUT_REQ_COMMITTED,
	})
	return
}

func (c *client) putForceWithOpts(ctx context.Context, kv clientv3.KV,
	req storage.PutAllRequestWithTTL) (
	resp *clientv3.TxnResponse, err error) {
	var ops []clientv3.Op

	// Translate int64 to clientv3.LeaseID
	var leaseID clientv3.LeaseID
	leaseID = clientv3.LeaseID(req.LeaseID)

	// If LeaseID doesn't exist for a key
	if leaseID == 0 {
		// If etcd doesn't respond, then don't hung longer than c.optime
		ctx1, cancel1 := context.WithTimeout(ctx, c.optime)

		// Create new LeaseID for a key with an expiration time of req.TTL seconds
		lease, err := c.cli.Grant(ctx1, req.TTL)
		cancel1()
		if err != nil {
			return nil, GrantNewLeaseIDError{
				Err:         err,
				Key:         req.Key,
				Description: _ETCD_LEASE_GRANT,
			}
		}
		leaseID = lease.ID
	}

	// Create Lease OpOption for a transaction
	leaseOp := clientv3.WithLease(leaseID)

	// Transaction will include only PUT with TTL query
	ops = append(ops, clientv3.OpPut(req.Key, string(req.Value), leaseOp))

	// doRetry will wrap ctx to context.WithTimeout(ctx)
	return c.doRetry(ctx, func(ctx context.Context) (*clientv3.TxnResponse, error) {
		return kv.Txn(ctx).Then(ops...).Commit()
	})
}

func (c *client) Delete(ctx context.Context, key string) error {
	kv, err := c.kv(ctx)
	if err != nil {
		return log.Alert(DeleteKVError{Err: err,
			Description: _ETCD_CLIENT_N_CONN})
	}

	log.Debug(ctx, DeleteRequest{Key: key, Description: _ETCD_DELETE})

	// Delete a key
	resp, err := c.delete(ctx, kv, key)
	if err != nil || !resp.Succeeded {
		return log.Alert(DeleteError{
			Key:         key,
			Err:         err,
			Description: _ETCD_DELETE_DB_FAIL,
		})
	}

	log.Debug(ctx, DeleteResponse{
		Response:    resp,
		Success:     resp.Succeeded,
		Description: _ETCD_DELETE_REQ_COMMITTED,
	})
	return nil
}

func (c *client) delete(ctx context.Context, kv clientv3.KV, key string) (
	resp *clientv3.TxnResponse, err error) {
	var ops []clientv3.Op

	// Transaction will include only DELETE query
	ops = append(ops, clientv3.OpDelete(key))

	// doRetry will wrap ctx to context.WithTimeout(ctx)
	return c.doRetry(ctx, func(ctx context.Context) (*clientv3.TxnResponse, error) {
		return kv.Txn(ctx).Then(ops...).Commit()
	})
}
