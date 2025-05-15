package cache

import (
	"crypto/tls"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/config"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
	"github.com/valkey-io/valkey-go"
)

const TTLSeconds = 3600  // 3 hours
const LonelyMinutes = 15 // 15 minutes

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

func (v Valkey) Close() {
	v.client.Close()
}
