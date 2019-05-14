// +build development

package storage

import (
	"context"

	"ivxv.ee/command/status"
)

// This file contains storage.Client methods which are only available in
// development mode. These provide direct access to the storage service and
// must be disabled in production code.

// PutAll exports the internal c.putAll method used for optimally storing many
// keys at once.
func (c *Client) PutAll(ctx context.Context, prefix string,
	values map[string][]byte, ensure bool, progress status.Add) error {

	return c.putAll(ctx, prefix, values, ensure, progress)
}
