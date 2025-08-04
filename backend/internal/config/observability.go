package config

import (
	"fmt"
	"time"
)

type ObservabilityConfig struct {
	ServiceName  string             `koanf:"service_name" validate:"required"`
	Environment  string             `koanf:"environment" validate:"required"`
	Logging      LoggingConfig      `koanf:"logging" validate:"required"`
	NewRelic     NewRelicConfig     `koanf:"new_relic" validate:"required"`
	HealthChecks HealthChecksConfig `koanf:"health_checks" validate:"required"`
}

type LoggingConfig struct {
	Level              string        `koanf:"level" validate:"required"`
	Format             string        `koanf:"format" validate:"required"`
	SlowQueryThreshold time.Duration `koanf:"slow_query_threshold"`
}

type NewRelicConfig struct {
	LicenseKey                string `koanf:"license_key" validate:"required"`
	AppLogForwardingEnabled   bool   `koanf:"app_log_forwarding_enabled"`
	DistributedTracingEnabled bool   `koanf:"distributed_tracing_enabled"`
	DebugLogging              bool   `koanf:"debug_logging"`
}

type HealthChecksConfig struct {
	Enabled  bool          `koanf:"enabled"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
	Checks   []string      `koanf:"checks"`
}

func DefaultObsConf() *ObservabilityConfig {
	return &ObservabilityConfig{
		ServiceName: "go_bp",
		Environment: "dev",
		Logging: LoggingConfig{
			Level:              "info",
			Format:             "json",
			SlowQueryThreshold: 100 * time.Millisecond,
		},
		NewRelic: NewRelicConfig{
			LicenseKey:                "",
			AppLogForwardingEnabled:   true,
			DistributedTracingEnabled: true,
			DebugLogging:              true,
		},
		HealthChecks: HealthChecksConfig{
			Enabled:  true,
			Interval: time.Duration(30 * time.Second),
			Timeout:  time.Duration(5 * time.Second),
			Checks:   []string{"database", "redis"},
		},
	}
}

// Validates ensure we have recieved the correct fields
func (o *ObservabilityConfig) Validate() error {
	if o.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	// validate log levels
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[o.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug,info,warn,error)", o.Logging.Level)
	}
	return nil
}

func (o *ObservabilityConfig) GetLogLevel() string {
	switch o.Logging.Level {
	case "prod":
		if o.Logging.Level == "" {
			return "info"
		}
	case "dev":
		if o.Logging.Level == "" {
			return "debug"
		}
	}
	return o.Logging.Level
}

func (o *ObservabilityConfig) IsProduction() bool {
	return o.Environment == "prod"
}
