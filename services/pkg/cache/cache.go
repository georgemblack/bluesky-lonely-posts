package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/config"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
	"github.com/valkey-io/valkey-go"
	"github.com/vmihailenco/msgpack/v5"
)

const TTLSeconds = 7200 // 12 hours

type Valkey struct {
	client valkey.Client
}

// New creates a new Valkey client.
func New(cfg config.Config) (Valkey, error) {
	var tlsConfig *tls.Config // nil by default
	if cfg.ValkeyTLSEnabled {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: false, // Validate the server's certificate
		}
	}

	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.ValkeyAddress},
		TLSConfig:   tlsConfig,
	})
	if err != nil {
		return Valkey{}, util.WrapErr("failed to create valkey client", err)
	}

	return Valkey{client: client}, nil
}

// SavePost saves a post record to the cache.
func (v Valkey) SavePost(hash string, post PostRecord) error {
	bytes, err := msgpack.Marshal(post)
	if err != nil {
		return util.WrapErr("failed to marshal record", err)
	}

	key := fmt.Sprintf("post:%s", hash)
	cmd := v.client.B().Set().Key(key).Value(string(bytes)).Ex(time.Second * TTLSeconds).Build()
	err = v.client.Do(context.Background(), cmd).Error()
	if err != nil {
		return util.WrapErr("failed to set key", err)
	}

	return nil
}

// ReadPost reads a post record from the cache. If the record does not exist, return an empty record.
func (v Valkey) ReadPost(hash string) (PostRecord, error) {
	key := fmt.Sprintf("post:%s", hash)
	cmd := v.client.B().Get().Key(key).Build()
	resp := v.client.Do(context.Background(), cmd)
	if err := resp.Error(); err != nil {
		if err == valkey.Nil {
			return PostRecord{}, nil
		}
		return PostRecord{}, util.WrapErr("failed to execute get command", err)
	}

	bytes, err := resp.AsBytes()
	if err != nil {
		return PostRecord{}, util.WrapErr("failed to convert response to bytes", err)
	}

	var record PostRecord
	err = msgpack.Unmarshal(bytes, &record)
	if err != nil {
		return PostRecord{}, util.WrapErr("failed to unmarshal record", err)
	}

	return record, nil
}

// DeletePost deletes a post record from the cache.
func (v Valkey) DeletePost(hash string) error {
	key := fmt.Sprintf("post:%s", hash)
	cmd := v.client.B().Del().Key(key).Build()
	err := v.client.Do(context.Background(), cmd).Error()
	if err != nil {
		return util.WrapErr("failed to delete key", err)
	}
	return nil
}

func (v Valkey) Close() {
	v.client.Close()
}

func (v Valkey) FindPosts(cursor uint64) ([]PostRecord, uint64, error) {
	cmd := v.client.B().Scan().Cursor(cursor).Match("post:*").Count(10).Build()
	resp := v.client.Do(context.Background(), cmd)
	if err := resp.Error(); err != nil {
		return nil, 0, util.WrapErr("failed to execute scan command", err)
	}

	res, err := resp.ToArray()
	if err != nil {
		return nil, 0, util.WrapErr("failed to convert response to array", err)
	}
	if len(res) < 2 {
		return nil, 0, util.WrapErr("invalid scan response format", nil)
	}

	next, err := res[0].AsUint64()
	if err != nil {
		return nil, 0, util.WrapErr("failed to parse cursor", err)
	}

	keys, err := res[1].AsStrSlice()
	if err != nil {
		return nil, 0, util.WrapErr("failed to parse keys", err)
	}

	records := make([]PostRecord, 0, len(keys))
	for _, key := range keys {
		hash := key[5:]
		record, err := v.ReadPost(hash)
		if err != nil {
			return nil, 0, util.WrapErr(fmt.Sprintf("failed to read post with hash %s", hash), err)
		}
		records = append(records, record)
	}

	return records, next, nil
}
