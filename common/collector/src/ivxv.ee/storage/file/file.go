//ivxv:development

/*
Package file implements a storage protocol which read and writes data from the
local filesystem.

The keys are Base64-encoded and stored as files in the working directory.

Note! This protocol does not provide any synchronization between concurrent
users and is not suitable for production. It should only be used for testing.
*/
package file // import "ivxv.ee/storage/file"

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"ivxv.ee/log"
	"ivxv.ee/storage"
	"ivxv.ee/yaml"
)

func init() {
	storage.Register(storage.File, func(n yaml.Node, _ *storage.Services) (
		s storage.PutGetter, err error) {

		var f F
		if err = yaml.Apply(n, &f); err != nil {
			return nil, ConfigurationError{Err: err}
		}
		// nolint: gosec, we want group permissions so that all
		// services access the files.
		if err = os.MkdirAll(f.WD, 0770); err != nil {
			return nil, WorkingDirectoryError{Path: f.WD, Err: err}
		}
		return f, nil
	})
}

// F implements the file storage protocol.
type F struct {
	WD string // All keys are stored in this working directory.
}

// Put implements the storage.PutGetter interface.
func (f F) Put(ctx context.Context, key string, value []byte) (err error) {
	// nolint: gosec, we want group permissions so that all services can
	// access the files.
	fp, err := os.OpenFile(filepath.Join(f.WD, encode(key)),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0660)
	if err != nil {
		if os.IsExist(err) {
			return storage.ExistError{Key: key, Err: PutExistingKeyError{Err: err}}
		}
		return log.Alert(PutKeyError{Key: key, Err: err})
	}
	defer func() {
		if cerr := fp.Close(); cerr != nil && err == nil {
			err = log.Alert(PutCloseError{Key: key, Err: cerr})
		}
	}()

	if _, err = fp.Write(value); err != nil {
		return log.Alert(PutValueError{Key: key, Err: err})
	}
	return
}

// Get implements the storage.PutGetter interface.
func (f F) Get(ctx context.Context, key string) (value []byte, err error) {
	if value, err = ioutil.ReadFile(filepath.Join(f.WD, encode(key))); err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NotExistError{Key: key, Err: GetMissingKeyError{Err: err}}
		}
		return nil, log.Alert(GetError{Key: key, Err: err})
	}
	return
}

// GetWithPrefix implements the storage.PutGetter interface.
func (f F) GetWithPrefix(ctx context.Context, prefix string) (
	<-chan storage.GetWithPrefixResult, <-chan error) {

	c := make(chan storage.GetWithPrefixResult)
	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		defer close(c)

		d, err := os.Open(f.WD)
		if err != nil {
			errc <- log.Alert(GetWithPrefixOpenWDError{Err: err})
			return
		}
		defer d.Close() // nolint: errcheck, ignore close failure of read-only fd.

		for {
			// Read next set of keys. Do not read all at once in
			// case we have a lot of keys.
			keys, err := d.Readdirnames(128) // n chosen arbitrarily.
			switch err {
			case nil:
			case io.EOF:
				return // No more keys.
			default:
				errc <- log.Alert(GetWithPrefixReadDirError{Err: err})
				return
			}

			// Get the values for matching keys and stream them to
			// the channel.
			for _, key := range keys {
				var r storage.GetWithPrefixResult
				if r.Key, err = decode(key); err != nil {
					errc <- log.Alert(GetWithPrefixDecodeError{
						Key: key,
						Err: err,
					})
					return
				}

				if !strings.HasPrefix(r.Key, prefix) {
					continue
				}

				if r.Value, err = f.Get(ctx, r.Key); err != nil {
					errc <- log.Alert(GetWithPrefixGetError{
						Key: key,
						Err: err,
					})
					return
				}

				select {
				case c <- r:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}
		}
	}()
	return c, errc
}

// CAS implements the storage.PutGetter interface.
//
// Note! This method is not actually synchronized.
func (f F) CAS(ctx context.Context, cas string, old, new []byte) (err error) {

	// Read the old value and compare it to the expected one.
	fp, err := os.OpenFile(filepath.Join(f.WD, encode(cas)), os.O_RDWR, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return storage.NotExistError{
				Key: cas,
				Err: CASMissingCASKeyError{Err: err},
			}
		}
		return log.Alert(CASKeyError{Key: cas, Err: err})
	}
	defer func() {
		if cerr := fp.Close(); cerr != nil && err == nil {
			err = log.Alert(CASCloseError{Key: cas, Err: err})
		}
	}()

	value, err := ioutil.ReadAll(fp)
	if err != nil {
		return log.Alert(CASReadValueError{Key: cas, Err: err})
	}
	if !bytes.Equal(value, old) {
		return storage.UnexpectedValueError{
			Key: cas,
			Err: CASValueMismatchError{
				Have:     value,
				Expected: old,
			}}
	}

	// Set the new value.
	if _, err = fp.WriteAt(new, 0); err != nil {
		return log.Alert(CASWriteValueError{Err: err})
	}
	if err = fp.Truncate(int64(len(new))); err != nil {
		return log.Alert(CASTruncateValueError{Err: err})
	}
	return
}

func encode(key string) string {
	return base64.URLEncoding.EncodeToString([]byte(key))
}

func decode(key string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(key)
	return string(b), err
}
