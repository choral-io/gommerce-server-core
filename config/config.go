package config

import (
	"errors"
	"log/slog"
	"time"

	"github.com/rs/cors"
	"go.uber.org/fx"
)

var (
	ErrConfigNotLoaded = errors.New("config not loaded")
)

type ExtractSectionsResult struct {
	fx.Out

	ServerConfig       ServerConfig
	ServerHTTPConfig   ServerHTTPConfig
	ServerDBConfig     ServerDBConfig
	ServerRedisConfig  ServerRedisConfig
	ServerLockerConfig ServerLockerConfig
	ServerMinIOConfig  ServerMinIOConfig
	ServerNATSConfig   ServerNATSConfig
	SnowflakeConfig    SnowflakeConfig
	LoggingConfig      LoggingConfig
	TraceConfig        TraceConfig
	MetricConfig       MetricConfig
	SecureConfig       SecureConfig
	SecureTokenConfig  SecureTokenConfig
}

// ExtractSections extracts sections from RootConfig.
// It is used for dependency injection.
func ExtractSections(cfg RootConfig) ExtractSectionsResult {
	result := ExtractSectionsResult{
		ServerConfig:       cfg.GetServerConfig(),
		ServerHTTPConfig:   cfg.GetServerConfig().GetHTTPConfig(),
		ServerDBConfig:     cfg.GetServerConfig().GetDBConfig(),
		ServerRedisConfig:  cfg.GetServerConfig().GetRedisConfig(),
		ServerLockerConfig: cfg.GetServerConfig().GetLockerConfig(),
		ServerMinIOConfig:  cfg.GetServerConfig().GetMinIOConfig(),
		ServerNATSConfig:   cfg.GetServerConfig().GetNATSConfig(),
		SnowflakeConfig:    cfg.GetSnowflakeConfig(),
		LoggingConfig:      cfg.GetLoggingConfig(),
		TraceConfig:        cfg.GetTraceConfig(),
		MetricConfig:       cfg.GetMetricConfig(),
		SecureConfig:       cfg.GetSecureConfig(),
		SecureTokenConfig:  cfg.GetSecureConfig().GetToken(),
	}
	return result
}

type RootConfig interface {
	// GetValue retrieves a value from the configuration by its path.
	// The path is a valid JSON path, e.g. "$.server.http.addr".
	GetValue(path string, value any) error
	GetServerConfig() ServerConfig
	GetSnowflakeConfig() SnowflakeConfig
	GetLoggingConfig() LoggingConfig
	GetTraceConfig() TraceConfig
	GetMetricConfig() MetricConfig
	GetSecureConfig() SecureConfig
}

type ServerConfig interface {
	GetDebug() bool
	GetName() string
	GetVersion() string
	GetInstanceId() string
	GetHTTPConfig() ServerHTTPConfig
	GetDBConfig() ServerDBConfig
	GetRedisConfig() ServerRedisConfig
	GetLockerConfig() ServerLockerConfig
	GetMinIOConfig() ServerMinIOConfig
	GetNATSConfig() ServerNATSConfig
}

type ServerHTTPConfig interface {
	GetAddr() string
	GetCors() cors.Options
}

type ServerDBConfig interface {
	GetDriver() string
	GetSource() string
}

type ServerRedisConfig interface {
	GetInitAddr() string
	GetSelectDB() int
}

type ServerLockerConfig interface {
	GetKeyPrefix() string
	GetKeyValidity() time.Duration
	GetExtendInterval() time.Duration
	GetTryNextAfter() time.Duration
	GetKeyMajority() int32
	GetNoLoopTracking() bool
	GetFallbackSETPX() bool
}

type ServerMinIOConfig interface {
	GetEndpoint() string
	GetAccessKey() string
	GetSecretKey() string
	GetUseSSL() bool
}

type ServerNATSConfig interface {
	GetSeedURL() string
	GetNoEcho() bool
}

type SnowflakeConfig interface {
	GetIdEpoch() int64
	GetClusterId() int64
	GetWorkerId() int64
	GetWorkerSeqKey() string
	GetClusterIdBits() int32
	GetWorkerIdBits() int32
	GetSequenceBits() int32
}

type LoggingConfig interface {
	GetZapLogger() LoggingZapLoggerConfig
	GetSlogLogger() LoggingSlogLoggerConfig
}

type LoggingZapLoggerConfig interface {
	GetPreset() string
}

type LoggingSlogLoggerConfig interface {
	GetHandler() string
	GetAddSource() bool
	GetLeveler() slog.Leveler
}

type TraceConfig interface {
	GetExporterConfig() TraceExporterConfig
}

type TraceExporterConfig interface {
	GetProtocol() string
	GetEndpoint() string
	GetInsecure() bool
}

type MetricConfig interface {
	GetExporterConfig() MetricExporterConfig
}

type MetricExporterConfig interface {
	GetProtocol() string
	GetEndpoint() string
	GetInsecure() bool
}

type SecureConfig interface {
	GetToken() SecureTokenConfig
}

type SecureTokenConfig interface {
	GetStore() string
	GetBucket() string
	GetIssuer() string
	GetAudience() string
	GetSigningMethod() string
	GetAccessTokenTTL() time.Duration
	GetRefreshTokenTTL() time.Duration
	GetPublicKey() []byte
	GetPrivateKey() []byte
}
