package config

import (
	"encoding/json"
	"log/slog"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

type Config struct {
	ValkeyAddress    string
	ValkeyTLSEnabled bool
}

func New() (Config, error) {
	result := Config{
		ValkeyAddress:    util.GetEnvStr("VALKEY_ADDRESS", "127.0.0.1:6379"),
		ValkeyTLSEnabled: util.GetEnvBool("VALKEY_TLS_ENABLED", false),
	}

	// Marshal to JSON and print if debug is enabled
	data, err := json.Marshal(result)
	if err != nil {
		slog.Warn(util.WrapErr("failed to marshal config", err).Error())
	}
	slog.Debug("generated config", "config", string(data))

	return result, nil
}
