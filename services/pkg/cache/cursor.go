package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
	"github.com/valkey-io/valkey-go"
)

const cursorKey = "cursor"

// SaveCursor saves the Jetstream cursor to the cache, as the key 'cursor'.
// The cursor expires after 2 minutes in the cache.
func (v Valkey) SaveCursor(cursor int64) error {
	value := strconv.FormatInt(cursor, 10)

	cmd := v.client.B().Set().Key(cursorKey).Value(value).Ex(time.Second * 120).Build()
	err := v.client.Do(context.Background(), cmd).Error()
	if err != nil {
		return util.WrapErr("failed to save cursor", err)
	}

	return nil
}

func (v Valkey) ReadCursor() (int64, error) {
	cmd := v.client.B().Get().Key(cursorKey).Build()
	resp := v.client.Do(context.Background(), cmd)
	if err := resp.Error(); err != nil {
		if err == valkey.Nil {
			return 0, nil
		}
		return 0, util.WrapErr("failed to execute get command", err)
	}

	bytes, err := resp.AsBytes()
	if err != nil {
		return 0, util.WrapErr("failed to convert response to bytes", err)
	}

	cursor, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return 0, util.WrapErr("failed to parse cursor", err)
	}

	return cursor, nil
}
