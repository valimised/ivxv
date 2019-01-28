/*
Package etcd implements a storage protocol which reads and writes data from an
etcd cluster.
*/
package etcd // import "ivxv.ee/storage/etcd"

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"runtime"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"

	"ivxv.ee/conf"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/log"
	"ivxv.ee/storage"
	"ivxv.ee/yaml"
)

func init() {
	storage.Register(storage.Etcd, func(n yaml.Node, services *storage.Services) (
		s storage.PutGetter, err error) {

		c := new(client)

		if len(services.Servers) == 0 {
			return nil, NoStorageServersError{}
		}
		c.endpoints = services.Servers

		var cfg Conf
		if err = yaml.Apply(n, &cfg); err != nil {
			return nil, ConfigurationError{Err: err}
		}
		ca, err := cryptoutil.PEMCertificatePool(cfg.CA)
		if err != nil {
			return nil, ConfigurationCAError{Err: err}
		}

		cert, err := tls.LoadX509KeyPair(conf.TLS(services.Sensitive))
		if err != nil {
			return nil, TLSKeyPairError{Err: err}
		}

		// Parse the leaf certificate and verify that it is issued by
		// CA and can be used for client authentication.
		if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
			return nil, ParseTLSCertificateError{Err: err}
		}
		if _, err = cert.Leaf.Verify(x509.VerifyOptions{
			Roots:     ca,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}); err != nil {
			return nil, VerifyTLSCertificateError{Err: err}
		}

		c.tls = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			RootCAs:      ca,
			Certificates: []tls.Certificate{cert},
		}

		c.conntime = time.Duration(cfg.ConnTimeout) * time.Second
		c.optime = time.Duration(cfg.OpTimeout) * time.Second
		return c, nil
	})
}

// Conf is the etcd storage protocol configuration.
type Conf struct {
	// CA is the PEM-encoding of the CA certificate which issued etcd TLS
	// certificates.
	CA string

	// ConnTimeout is the timeout for connecting to etcd endpoints in seconds.
	ConnTimeout int64

	// OpTimeout is the timeout for performing a single operation in seconds.
	OpTimeout int64

	// Bootstrap lists the IDs of etcd servers that are used to bootstrap
	// the cluster. Necessary so a server knows if they are part of the
	// bootstrap or joining an existing cluster.
	Bootstrap []string
}

type client struct {
	endpoints []string
	tls       *tls.Config
	conntime  time.Duration
	optime    time.Duration

	// cli is the etcd client and clierr is the (potential) error that we
	// got when creating the client. clionce ensures we only initialize the
	// client once.
	cli     *clientv3.Client
	clierr  error
	clionce sync.Once
}

func (c *client) BatchSize() int {
	return 128 // Not configurable in etcd v3.1.
}

// kv returns the KV interface of the client, initializing it if not already
// done. If initialization failed, then all calls to kv will return that error.
func (c *client) kv(ctx context.Context) (kv clientv3.KV, err error) {
	c.clionce.Do(func() {
		// Strip connection and session ID of the first connection to
		// call kv from the context, because the logger will be reused
		// for all future connections.
		ctx = log.WithConnectionID(ctx, "")
		ctx = log.WithSessionID(ctx, "")
		clientv3.SetLogger(newLogger(ctx))

		c.cli, c.clierr = clientv3.New(clientv3.Config{
			Endpoints:   c.endpoints,
			DialTimeout: c.conntime,
			TLS:         c.tls,
		})
		if c.clierr == nil {
			log.Log(ctx, Connected{Endpoints: c.endpoints})
		}
	})
	if err = c.clierr; err == nil {
		kv = c.cli.KV
	}
	return
}

func (c *client) Put(ctx context.Context, key string, value []byte) (err error) {
	kv, err := c.kv(ctx)
	if err != nil {
		return log.Alert(PutKVError{Err: err})
	}

	log.Debug(ctx, PutRequest{Key: key, Value: value})
	resp, err := c.put(ctx, kv, storage.PutAllRequest{Key: key, Value: value})
	if err != nil {
		return log.Alert(PutError{Key: key, Err: err})
	}
	log.Debug(ctx, PutResponse{Response: resp})

	if !resp.Succeeded {
		return storage.ExistError{Key: key, Err: PutExistingKeyError{}}
	}
	return
}

func (c *client) PutAll(ctx context.Context, reqs ...storage.PutAllRequest) (err error) {
	kv, err := c.kv(ctx)
	if err != nil {
		return log.Alert(PutAllKVError{Err: err})
	}

	log.Debug(ctx, PutAllRequest{Count: len(reqs)})
	resp, err := c.put(ctx, kv, reqs...)
	if err != nil {
		return log.Alert(PutAllError{Err: err})
	}
	log.Debug(ctx, PutAllResponse{Response: resp})

	if !resp.Succeeded {
		var key string
		for _, rop := range resp.Responses { // Find the first existing key.
			if kvs := rop.GetResponseRange().GetKvs(); len(kvs) > 0 {
				key = string(kvs[0].Key)
				break
			}
		}
		return storage.ExistError{Key: key, Err: PutAllExistingKeyError{}}
	}
	return
}

func (c *client) put(ctx context.Context, kv clientv3.KV, reqs ...storage.PutAllRequest) (
	resp *clientv3.TxnResponse, err error) {

	var cmps []clientv3.Cmp
	var ops []clientv3.Op
	var fail []clientv3.Op
	for _, req := range reqs {
		// clientv3/clientv3util.KeyMissing in newer etcd.
		cmps = append(cmps, clientv3.Compare(clientv3.Version(req.Key), "=", 0))
		ops = append(ops, clientv3.OpPut(req.Key, string(req.Value)))
		fail = append(fail, clientv3.OpGet(req.Key, clientv3.WithKeysOnly()))
	}

	return c.doRetry(ctx, func(ctx context.Context) (*clientv3.TxnResponse, error) {
		return kv.Txn(ctx).If(cmps...).Then(ops...).Else(fail...).Commit()
	})
}

func (c *client) Get(ctx context.Context, key string) (value []byte, err error) {
	kv, err := c.kv(ctx)
	if err != nil {
		return nil, log.Alert(GetKVError{Err: err})
	}

	log.Debug(ctx, GetRequest{Key: key})
	ctx, cancel := context.WithTimeout(ctx, c.optime)
	defer cancel()
	resp, err := kv.Get(ctx, key)
	if err != nil {
		return nil, log.Alert(GetError{Key: key, Err: err})
	}
	log.Debug(ctx, GetResponse{Response: resp})

	if len(resp.Kvs) == 0 {
		return nil, storage.NotExistError{Key: key, Err: GetMissingKeyError{}}
	}
	value = resp.Kvs[0].Value
	return
}

func (c *client) GetAll(ctx context.Context, keys ...string) (values map[string][]byte, err error) {
	kv, err := c.kv(ctx)
	if err != nil {
		return nil, log.Alert(GetAllKVError{Err: err})
	}

	// Use a transaction to batch all the gets together.
	var ops []clientv3.Op
	for _, key := range keys {
		ops = append(ops, clientv3.OpGet(key))
	}

	log.Debug(ctx, GetAllRequest{Count: len(keys)})
	ctx, cancel := context.WithTimeout(ctx, c.optime)
	defer cancel()
	resp, err := kv.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return nil, log.Alert(GetAllError{Err: err})
	}
	log.Debug(ctx, GetAllResponse{Response: resp})

	values = make(map[string][]byte)
	for _, rop := range resp.Responses {
		if kvs := rop.GetResponseRange().GetKvs(); len(kvs) > 0 {
			values[string(kvs[0].Key)] = kvs[0].Value
		}
	}
	return
}

func (c *client) GetWithPrefix(ctx context.Context, prefix string) (
	<-chan storage.GetWithPrefixResult, <-chan error) {

	ch := make(chan storage.GetWithPrefixResult)

	// Since there are likely too many keys to get in a single request, we
	// need to perform multiple queries. In order to speed this up, perform
	// them in parallel using workers up to the number of threads.
	workers := runtime.NumCPU()
	errc := make(chan error, workers+1) // All goroutines can send without blocking.

	kv, err := c.kv(ctx)
	if err != nil {
		close(ch)
		errc <- log.Alert(GetWithPrefixKVError{Err: err})
		return ch, errc
	}

	// First get the number of keys with the prefix to know how many jobs
	// are needed to get them all.
	log.Debug(ctx, GetWithPrefixCountRequest{Prefix: prefix})
	mctx, cancel := context.WithTimeout(ctx, c.optime)
	defer cancel()
	resp, err := kv.Get(mctx, prefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		close(ch)
		errc <- log.Alert(GetWithPrefixCountError{Err: err})
		return ch, errc
	}
	log.Debug(ctx, GetWithPrefixCountResponse{Response: resp})

	if resp.Count == 0 { // No such keys, stop now.
		close(ch)
		close(errc)
		return ch, errc
	}

	// Determine the number of requests that must be performed given the
	// count of keys and block size. If less than the number of workers,
	// reduce workers. No need to reduce errc buffer.
	if jobs := (resp.Count + blockSize - 1) / blockSize; jobs < int64(workers) {
		workers = int(jobs)
	}

	// Divide the keyspace into *roughly* equal segments between all
	// workers. Simplify this into dividing the first byte into segments:
	// assumes that the keys are equally distributed, segments will be
	// unequal otherwise.
	step := byte(256 / workers) // Overflows if workers == 1, but will not be used then.
	var current byte
	segments := make([]string, workers+1)
	segments[0] = prefix
	for i := 1; i < workers; i++ {
		current += step
		segments[i] = prefix + string(current)
	}
	segments[len(segments)-1] = clientv3.GetPrefixRangeEnd(prefix)

	// Start worker pool where each goroutine sends the key-values from
	// their segment on blockc.
	blockc := make(chan *clientv3.GetResponse)
	wctx, cancel := context.WithCancel(ctx) // Canceled on worker error.

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < len(segments)-1; i++ {
		go func(from, to string) {
			defer wg.Done()
			if err := c.getRange(wctx, kv, blockc, from, to); err != nil {
				errc <- err
				cancel() // Notify other workers of failure.
			}
		}(segments[i], segments[i+1])
	}
	go func() {
		defer cancel() // Free worker context resources on success.
		wg.Wait()
		close(blockc) // All workers are done.
	}()

	// Stream values from blocks to the caller.
	go func() {
		defer close(ch)
		for block := range blockc {
			for _, kv := range block.Kvs {
				select {
				case ch <- storage.GetWithPrefixResult{
					Key:   string(kv.Key),
					Value: kv.Value,
				}:
				// Check ctx instead of wctx so that we still
				// send all values that we got before a worker
				// error occured.
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}
		}
		// Only close on success and after all workers are done, i.e.,
		// blockc is closed.
		close(errc)
	}()

	return ch, errc
}

// etcd does not have an iterating API so getRange queries data in blocks using
// WithRange and WithLimit. Through testing on a real-life setup we determined
// 6144 to be a good balance between speed and query time. Note that this
// number was reached when using an OpTimeout of 10 seconds. If a smaller
// timeout is used, then this probably needs to be reduced as well.
const blockSize = 6144

// getRange sends the keys in range [from, to) that were created in or before
// revision rev on ch in blocks of blockSize. It sends the results in blocks so
// that it does not have to wait sending individual keys and can fetch the next
// block while the previous one is being processed.
func (c *client) getRange(ctx context.Context, kv clientv3.KV,
	ch chan<- *clientv3.GetResponse, from, to string) error {

	// Although etcd provides a WithMaxCreateRev option so that we would
	// get only keys that existed at a given time, it slows down requests
	// drastically. Rather, accept that the results can contain some keys
	// that did not exist when getRange was called.
	ops := []clientv3.OpOption{
		clientv3.WithRange(to),
		clientv3.WithLimit(blockSize),
	}
	var n int // Keep count of keys retrieved by this thread for debugging.
	key := from
	for {
		// Get the next block of key-values.
		log.Debug(ctx, GetRangeRequest{Key: key, Range: to})
		opctx, cancel := context.WithTimeout(ctx, c.optime)
		resp, err := kv.Get(opctx, key, ops...)
		cancel() // We do not want to defer in a loop.
		if err != nil {
			return log.Alert(GetRangeError{Key: key, Range: to, Err: err})
		}
		// Only log the three first and three last values as to not
		// explode the logs.
		var first, last string
		if len(resp.Kvs) > 0 {
			first = string(resp.Kvs[0].Key)
			last = string(resp.Kvs[len(resp.Kvs)-1].Key)
		}
		log.Debug(ctx, GetRangeResponse{
			Header: resp.Header,
			Count:  len(resp.Kvs), // resp.Count was not asked for.
			First:  first,
			Last:   last,
			More:   resp.More,
		})
		n += len(resp.Kvs)

		if len(resp.Kvs) == 0 {
			log.Debug(ctx, GetRangeCount{From: from, To: to, Count: n})
			return nil // No more keys.
		}

		select {
		case ch <- resp:
		case <-ctx.Done():
			return ctx.Err()
		}

		// The next block will be queried from the key immediately
		// succeeding the last key from this block.
		key = nextKey(resp.Kvs[len(resp.Kvs)-1].Key)
	}
}

// nextKey returns the key that lexicographically immediately follows key.
func nextKey(key []byte) string {
	return string(append(key, 0))
}

func (c *client) CAS(ctx context.Context, cas string, old, new []byte) (err error) {
	kv, err := c.kv(ctx)
	if err != nil {
		return log.Alert(CASKVError{Err: err})
	}

	log.Debug(ctx, CASRequest{CAS: cas, Old: old, New: new})
	resp, err := c.doRetry(ctx, func(ctx context.Context) (*clientv3.TxnResponse, error) {
		return kv.Txn(ctx).
			If(clientv3.Compare(clientv3.Value(cas), "=", string(old))).
			Then(clientv3.OpPut(cas, string(new))).
			Else(clientv3.OpGet(cas)).
			Commit()
	})
	if err != nil {
		return log.Alert(CASError{CAS: cas, Err: err})
	}
	log.Debug(ctx, CASResponse{Response: resp})

	if !resp.Succeeded {
		// Assume resp has the correct amount of Reponses.
		if kvs := resp.Responses[0].GetResponseRange().GetKvs(); len(kvs) > 0 {
			return storage.UnexpectedValueError{
				Key: cas,
				Err: CASValueMismatchError{
					Have: string(kvs[0].Value),
					Want: string(old),
				},
			}
		}
		return storage.NotExistError{
			Key: cas,
			Err: CASMissingCASKeyError{},
		}
	}
	return
}

// doRetry performs a transaction, automatically retrying up to three times if
// there are timeout errors.
//
// doRetry takes a function which constructs and commits the transaction. The
// transaction needs to be reconstructed each time, so that it can use a new
// context per retry, refreshing the operation timeout each time.
//
// Usually clientv3.Txn.Commit retries any operations itself until the context
// expires or the operation succeeds, but not in case of write requests.
// However, we still wish to retry if the error was caused by a timeout error:
// this works around any write errors caused by network issues or leader
// changes.
func (c *client) doRetry(ctx context.Context, f func(context.Context) (*clientv3.TxnResponse, error)) (
	resp *clientv3.TxnResponse, err error) {

	for i := 0; i < 3; i++ {
		rctx, cancel := context.WithTimeout(ctx, c.optime)
		defer cancel()
		switch resp, err = f(rctx); err {
		case rpctypes.ErrTimeout,
			rpctypes.ErrTimeoutDueToLeaderFail,
			rpctypes.ErrTimeoutDueToConnectionLost:

			time.Sleep(300 * time.Millisecond) // Give time to recover.
		default:
			return
		}
	}
	return // Out of retries, return resp and err from last attempt.
}
