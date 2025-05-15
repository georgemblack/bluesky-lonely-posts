package config

import (
	"encoding/json"
	"log/slog"

	"github.com/georgemblack/bluesky-lonely-posts/pkg/secrets"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

type Config struct {
	ValkeyAddress      string
	ValkeyTLSEnabled   bool
	CloudflareAPIToken string
	CloudflareZoneID   string
	ServerPort         string
}

func New() (Config, error) {
	sm, err := secrets.New()
	if err != nil {
		return Config{}, util.WrapErr("failed to create secrets manager", err)
	}

	apiToken, err := sm.GetCloudflareAPIToken()
	if err != nil {
		return Config{}, util.WrapErr("failed to get cloudflare api token", err)
	}

	zoneID, err := sm.GetCloudflareZoneID()
	if err != nil {
		return Config{}, util.WrapErr("failed to get cloudflare zone id", err)
	}

	result := Config{
		ValkeyAddress:      util.GetEnvStr("VALKEY_ADDRESS", "127.0.0.1:6379"),
		ValkeyTLSEnabled:   util.GetEnvBool("VALKEY_TLS_ENABLED", false),
		CloudflareAPIToken: apiToken,
		CloudflareZoneID:   zoneID,
		ServerPort:         util.GetEnvStr("SERVER_PORT", "8080"),
	}

	// Marshal to JSON and print if debug is enabled
	data, err := json.Marshal(result)
	if err != nil {
		slog.Warn(util.WrapErr("failed to marshal config", err).Error())
	}
	slog.Debug("generated config", "config", string(data))

	return result, nil
}
