package config

import (
	"errors"
	"log/slog"
	"time"

	"github.com/rs/cors"
	"go.uber.org/fx"
)

var (
	// ErrConfigNotLoaded is returned when configuration access is attempted
	// before the configuration has been fully loaded.
	ErrConfigNotLoaded = errors.New("config not loaded")
)

// ExtractSectionsResult is a container of individual configuration sections
// returned from ExtractSections for dependency injection via fx.Out.
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

// RootConfig represents the top-level configuration container loaded from
// external sources. Implementations must provide access to specific
// configuration sections used throughout the application.
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

// ServerConfig contains runtime configuration for the HTTP/gRPC server and
// related components.
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

// ServerHTTPConfig defines the configuration for the embedded HTTP server.
type ServerHTTPConfig interface {
	GetAddr() string
	GetCors() cors.Options
}

// ServerDBConfig captures database connection settings.
type ServerDBConfig interface {
	GetDriver() string
	GetSource() string
}

// ServerRedisConfig describes the Redis connection configuration.
type ServerRedisConfig interface {
	GetInitAddr() string
	GetSelectDB() int
}

// ServerLockerConfig defines options for the distributed locking mechanism.
type ServerLockerConfig interface {
	GetKeyPrefix() string
	GetKeyValidity() time.Duration
	GetExtendInterval() time.Duration
	GetTryNextAfter() time.Duration
	GetKeyMajority() int32
	GetNoLoopTracking() bool
	GetFallbackSETPX() bool
}

// ServerMinIOConfig specifies settings for the MinIO object storage backend.
type ServerMinIOConfig interface {
	GetEndpoint() string
	GetAccessKey() string
	GetSecretKey() string
	GetUseSSL() bool
}

// ServerNATSConfig holds the NATS connection options.
type ServerNATSConfig interface {
	GetSeedURL() string
	GetNoEcho() bool
}

// SnowflakeConfig exposes configuration for the Snowflake ID generator.
type SnowflakeConfig interface {
	GetIdEpoch() int64
	GetClusterId() int64
	GetWorkerId() int64
	GetWorkerSeqKey() string
	GetClusterIdBits() int32
	GetWorkerIdBits() int32
	GetSequenceBits() int32
}

// LoggingConfig selects logging implementations and presets for the service.
type LoggingConfig interface {
	GetZapLogger() LoggingZapLoggerConfig
	GetSlogLogger() LoggingSlogLoggerConfig
}

// LoggingZapLoggerConfig configures the zap logger preset.
type LoggingZapLoggerConfig interface {
	GetPreset() string
}

// LoggingSlogLoggerConfig controls settings for slog logging.
type LoggingSlogLoggerConfig interface {
	GetHandler() string
	GetAddSource() bool
	GetLeveler() slog.Leveler
}

// TraceConfig describes OpenTelemetry tracing options.
type TraceConfig interface {
	GetExporterConfig() TraceExporterConfig
}

// TraceExporterConfig defines options for the trace exporter.
type TraceExporterConfig interface {
	GetProtocol() string
	GetEndpoint() string
	GetInsecure() bool
}

// MetricConfig describes OpenTelemetry metrics configuration.
type MetricConfig interface {
	GetExporterConfig() MetricExporterConfig
}

// MetricExporterConfig defines options for the metrics exporter.
type MetricExporterConfig interface {
	GetProtocol() string
	GetEndpoint() string
	GetInsecure() bool
}

// SecureConfig encompasses security and token-related settings.
type SecureConfig interface {
	GetToken() SecureTokenConfig
}

// SecureTokenConfig provides parameters for issuing and verifying tokens.
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
