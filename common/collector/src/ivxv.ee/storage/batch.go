package storage

import (
	"bytes"
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"ivxv.ee/command/status"
	"ivxv.ee/errors"
	"ivxv.ee/log"
)

// getAll calls GetAll in batches of BatchSize if the storage protocol
// implements Batcher, otherwise fallbacks to just getting the keys one-by-one.
//
// getAll currently does not utilize worker pools, since we do not yet use
// it for any large gets (we have GetWithPrefix for that). If this changes,
// then we can start using them here as well.
func (c *Client) getAll(ctx context.Context, keys ...string) (values map[string][]byte, err error) {
	values = make(map[string][]byte)
	if batcher, ok := c.prot.(Batcher); ok {
		size := batcher.BatchSize()
		var index int
		for index < len(keys) {
			end := index + size
			if end > len(keys) {
				end = len(keys)
			}

			bval, err := batcher.GetAll(ctx, keys[index:end]...)
			if err != nil {
				return nil, GetAllError{Err: err}
			}
			for key, value := range bval {
				values[key] = value
			}

			index = end
		}
		return
	}

	for _, key := range keys {
		value, err := c.prot.Get(ctx, key)
		switch {
		case err == nil:
			values[key] = value
		case errors.CausedBy(err, new(NotExistError)) != nil:
			// Ignore missing keys to match the real GetAll.
		default:
			return nil, GetAllSingleError{Key: key, Err: err}
		}
	}
	return
}

// getAllStrict calls getAll and returns an error if any key is missing.
func (c *Client) getAllStrict(ctx context.Context, keys ...string) (
	values map[string][]byte, err error) {

	if values, err = c.getAll(ctx, keys...); err != nil {
		return nil, err
	}
	for _, key := range keys {
		if _, ok := values[key]; !ok {
			var ne NotExistError
			ne.Key = key
			ne.Err = GetAllStrictMissingError{}
			return nil, ne
		}
	}
	return values, nil
}

// putAll attempts to optimize putting many keys at once. It starts a worker
// pool which calls PutAll in batches of BatchSize if the storage protocol
// implements Batcher, otherwise fallbacks to just putting the keys one-by-one.
//
// Additionally, putAll prepends prefix to all keys: this allows to do this
// common operation here, where we need to iterate over the maps anyway.
//
// If ensure is true, then existing keys are allowed, as long as they have the
// value as the one being put.
//
// Progress of the operation is reported to progress as well as logged
// periodically.
func (c *Client) putAll(ctx context.Context, prefix string,
	values map[string][]byte, ensure bool, progress status.Add) error {

	// Child context to stop goroutines early on errors.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Define a common job type.
	type job struct {
		// Used by Batchers.
		batch []PutAllRequest
		keys  []string

		// Used by non-Batchers.
		key   string
		value []byte
	}

	// Start feeding jobs into a channel.
	jobc := make(chan job)

	batcher, isBatcher := c.prot.(Batcher)
	jobs := len(values)
	if isBatcher {
		size := batcher.BatchSize()
		jobs = (jobs + size - 1) / size
	}

	workers := runtime.NumCPU()
	errc := make(chan error, workers+1) // All goroutines can error without blocking.
	if isBatcher {
		go func() {
			defer close(jobc)
			batch := make([]PutAllRequest, 0, batcher.BatchSize())
			keys := make([]string, 0, cap(batch))

			for key, value := range values {
				batch = append(batch, PutAllRequest{prefix + key, value})
				keys = append(keys, prefix+key)
				if len(batch) < cap(batch) {
					continue
				}

				select {
				case jobc <- job{batch: batch, keys: keys}:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}

				// Create new slices instead of modifying the
				// ones sent to workers.
				batch = make([]PutAllRequest, 0, cap(batch))
				keys = make([]string, 0, cap(batch))
			}
			if len(batch) > 0 {
				select {
				case jobc <- job{batch: batch, keys: keys}:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}
		}()
	} else {
		go func() {
			defer close(jobc)
			for key, value := range values {
				select {
				case jobc <- job{key: prefix + key, value: value}:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}
		}()
	}

	// Determine the actual function to call per job.
	var work func(job) (int, error)
	switch {
	case isBatcher && ensure:
		work = func(j job) (int, error) { return len(j.batch), c.ensureBatch(ctx, j.batch, j.keys) }
	case isBatcher && !ensure:
		work = func(j job) (int, error) { return len(j.batch), batcher.PutAll(ctx, j.batch...) }
	case !isBatcher && ensure:
		work = func(j job) (int, error) { return 1, c.ensure(ctx, j.key, j.value) }
	case !isBatcher && !ensure:
		work = func(j job) (int, error) { return 1, c.prot.Put(ctx, j.key, j.value) }
	}

	// Start an ad-hoc worker pool to perform jobs from the channel. Use
	// min(jobs, threads) workers. No need to reduce errc buffer.
	if jobs < workers {
		workers = jobs
	}
	var wg sync.WaitGroup
	wg.Add(workers)

	const logstep = 10000 // Log progress after each logstep.
	logat := uint64(logstep)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			countlog := PutAllProgress{Current: 0}
			for job := range jobc {
				n, err := work(job)
				if err != nil {
					errc <- err
					return
				}

				total := progress(uint64(n))
				if old := atomic.LoadUint64(&logat); total >= old &&
					atomic.CompareAndSwapUint64(&logat, old, old+logstep) {

					countlog.Current = total
					log.Log(ctx, countlog)
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(errc)
	}()

	// Either we got an error or wg.Wait closed errc returning nil.
	return <-errc
}

// ensureBatch is used by putAll to perform a single Batcher.PutAll with ensure
// logic, i.e., allowing keys to exist if their values match the ones being
// put.
func (c *Client) ensureBatch(ctx context.Context, batch []PutAllRequest, keys []string) error {
	batcher := c.prot.(Batcher)

	// First try to blindly put everything, hoping that this will succeed
	// in the majority of cases.
	switch err := batcher.PutAll(ctx, batch...); {
	case err == nil:
		return nil
	case errors.CausedBy(err, new(ExistError)) == nil:
		return PutAllError{Err: err}
	}

	// Some keys in the batch already exist. Get their values and ensure
	// that they match the ones being put.
	existing, err := batcher.GetAll(ctx, keys...)
	if err != nil {
		return EnsureBatchGetExistingError{Err: err}
	}

	// Batch and keys for values that are missing.
	var mbatch []PutAllRequest
	var mkeys []string

	for _, kv := range batch {
		val, ok := existing[kv.Key]
		if !ok {
			mbatch = append(mbatch, kv)
			mkeys = append(mkeys, kv.Key)
			continue
		}

		if !bytes.Equal(val, kv.Value) {
			return EnsureBatchExistingMismatchError{
				Key:      kv.Key,
				Existing: val,
				New:      kv.Value,
			}
		}
	}

	if len(mbatch) > 0 {
		// Recurse to try to put the keys that did not already exist.
		return c.ensureBatch(ctx, mbatch, mkeys)
	}

	return nil
}
