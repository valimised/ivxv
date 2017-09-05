/*
Package etcd implements a storage protocol which reads and writes data from an
etcd cluster.
*/
package etcd // import "ivxv.ee/storage/etcd"

import (
	"context"
	"crypto/tls"
	"path/filepath"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"

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

		cert, err := tls.LoadX509KeyPair(
			filepath.Join(services.Sensitive, "tls.pem"),
			filepath.Join(services.Sensitive, "tls.key"))
		if err != nil {
			return nil, TLSKeyPairError{Err: err}
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

// kv returns the KV interface of the client, initalizing it if not already
// done. If initialization failed, then all calls to kv will return that error.
func (c *client) kv(ctx context.Context) (kv clientv3.KV, err error) {
	c.clionce.Do(func() {
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
	ctx, cancel := context.WithTimeout(ctx, c.optime)
	defer cancel()
	resp, err := kv.Txn(ctx).
		// clientv3/clientv3util.KeyMissing in newer etcd.
		If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, string(value))).
		Commit()
	if err != nil {
		return log.Alert(PutError{Key: key, Err: err})
	}
	log.Debug(ctx, PutResponse{Response: resp})

	if !resp.Succeeded {
		return storage.ExistError{Key: key, Err: PutExistingKeyError{}}
	}
	return
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

func (c *client) GetWithPrefix(ctx context.Context, prefix string) (
	<-chan storage.GetWithPrefixResult, <-chan error) {

	ch := make(chan storage.GetWithPrefixResult)
	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		defer close(ch)

		kv, err := c.kv(ctx)
		if err != nil {
			errc <- log.Alert(GetWithPrefixKVError{Err: err})
			return
		}

		// etcd does not have an iterating API so just query data in
		// blocks using WithRange and WithLimit (in case we have a lot
		// of keys).
		//
		// Although etcd provides a WithRev option which enables us to
		// query all blocks from the same revision, old revisions can
		// get compacted during this operation (even though unlikely)
		// resulting in errors.  Rather rely on the immutability
		// guarantee of our API that nothing gets removed behind us.
		// And on the other hand we are completely fine with new keys
		// being added during this operation.

		from := prefix // The key to start querying the block from.
		ops := []clientv3.OpOption{
			clientv3.WithRange(clientv3.GetPrefixRangeEnd(prefix)),
			clientv3.WithLimit(1024), // Block size, arbitrarily chosen.
		}

		for {
			// Get the next block of key-values.
			//
			// Even though the previous response has a More field
			// we can check to find out if there are more keys left
			// in the range, we will ignore it and still repeat the
			// query in case new keys were added after the last one
			// since our previous query.
			log.Debug(ctx, GetWithPrefixRequest{Prefix: prefix, From: from})
			opctx, cancel := context.WithTimeout(ctx, c.optime)
			resp, err := kv.Get(opctx, from, ops...)
			cancel() // We do not want to defer in a loop.
			if err != nil {
				errc <- log.Alert(GetWithPrefixError{
					Prefix: prefix,
					From:   from,
					Err:    err,
				})
				return
			}
			log.Debug(ctx, GetWithPrefixResponse{Response: resp})

			if len(resp.Kvs) == 0 {
				return // No more keys.
			}

			// Stream the block of key-values to the channel.
			for _, kv := range resp.Kvs {
				select {
				case ch <- storage.GetWithPrefixResult{
					Key:   string(kv.Key),
					Value: kv.Value,
				}:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}

			// The next block will be queried from the key
			// immediately succeeding the last key from
			// this block.
			from = string(append(resp.Kvs[len(resp.Kvs)-1].Key, '\x00'))
		}
	}()
	return ch, errc
}

func (c *client) CASAndGet(ctx context.Context, cas string, old, new []byte, keys ...string) (
	values map[string][]byte, err error) {

	kv, err := c.kv(ctx)
	if err != nil {
		return nil, log.Alert(CASAndGetKVError{Err: err})
	}

	cmp := clientv3.Compare(clientv3.Value(cas), "=", string(old))

	ops := make([]clientv3.Op, 0, len(keys)+1)
	ops = append(ops, clientv3.OpPut(cas, string(new)))
	for _, key := range keys {
		ops = append(ops, clientv3.OpGet(key))
	}

	fail := clientv3.OpGet(cas)

	log.Debug(ctx, CASAndGetRequest{CAS: cas, Old: old, New: new, Keys: keys})
	ctx, cancel := context.WithTimeout(ctx, c.optime)
	defer cancel()
	resp, err := kv.Txn(ctx).If(cmp).Then(ops...).Else(fail).Commit()
	if err != nil {
		return nil, log.Alert(CASAndGetError{CAS: cas, Keys: keys, Err: err})
	}
	log.Debug(ctx, CASAndGetResponse{Response: resp})

	// Assume resp has the correct amount of Reponses.

	if !resp.Succeeded {
		casresp := resp.Responses[0].GetResponseRange()
		if len(casresp.Kvs) == 0 {
			return nil, storage.NotExistError{
				Key: cas,
				Err: CASAndGetMissingCASKeyError{},
			}
		}
		return nil, storage.UnexpectedValueError{
			Key: cas,
			Err: CASAndGetValueMismatchError{
				Have: string(casresp.Kvs[0].Value),
				Want: string(old),
			},
		}
	}

	// The first response is for the OpPut, so skip it.
	keysresp := resp.Responses[1:]

	values = make(map[string][]byte)
	for i, opresp := range keysresp {
		getresp := opresp.GetResponseRange()
		if len(getresp.Kvs) == 0 {
			return nil, storage.NotExistError{
				Key: keys[i],
				Err: CASAndGetMissingKeyError{},
			}
		}
		values[string(getresp.Kvs[0].Key)] = getresp.Kvs[0].Value
	}
	return
}
