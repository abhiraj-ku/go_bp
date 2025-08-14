package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/abhiraj-ku/go_bp/internal/config"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/zerologWriter"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type LoggerService struct {
	nrApp *newrelic.Application
}

// NewLogger creates a new logger Service with newRelic integration
func NewLoggerService(cfg *config.ObservabilityConfig) *LoggerService {
	service := &LoggerService{}

	if cfg.NewRelic.LicenseKey == "" {
		fmt.Println("NewRelic license key needed")
		return service
	}

	// Inits NewRelic service
	var NewRelicConfig []newrelic.ConfigOption
	NewRelicConfig = append(NewRelicConfig,
		newrelic.ConfigAppName(cfg.ServiceName),
		newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(cfg.NewRelic.AppLogForwardingEnabled),
		newrelic.ConfigDistributedTracerEnabled(cfg.NewRelic.DistributedTracingEnabled),
	)
	if cfg.NewRelic.DebugLogging {
		NewRelicConfig = append(NewRelicConfig, newrelic.ConfigDebugLogger(os.Stdout))
	}

	app, err := newrelic.NewApplication(NewRelicConfig...)
	if err != nil {
		fmt.Println("Failed to init newRelic..%s\n", err)
	}

	service.nrApp = app
	fmt.Println("New relic app initialized: %s\n", cfg.ServiceName)
	return service

}

// shutdown logger service
func (ls *LoggerService) ShutDown() {
	if ls.nrApp != nil {
		ls.nrApp.Shutdown(10 * time.Second)

	}
}

// GetApplication returns the New Relic application instance
func (ls *LoggerService) GetNewRelic() *newrelic.Application {
	return ls.nrApp
}

func NewLogger(level string, isProd bool) zerolog.Logger {
	return NewLoggerWithService(&config.ObservabilityConfig{
		Logging: config.LoggingConfig{
			Level: level,
		},
		Environment: func() string {
			if isProd {
				return "prod"
			}
			return "dev"

		}(),
	}, nil)
}

func NewLoggerWithConfig(cfg *config.ObservabilityConfig) zerolog.Logger {
	return NewLoggerWithService(cfg, nil)
}

func NewLoggerWithService(cfg *config.ObservabilityConfig, loggerService *LoggerService) zerolog.Logger {

	var logLevel zerolog.Level

	level := cfg.GetLogLevel()
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel

	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var writer io.Writer

	var baseWriter io.Writer

	// checks if prod env or local/dev
	if cfg.IsProduction() && cfg.Logging.Level == "json" {
		baseWriter = os.Stdout

		if loggerService != nil && loggerService.nrApp != nil {
			nWriter := zerologWriter.New(baseWriter, loggerService.nrApp)
			writer = nWriter
		} else {
			writer = baseWriter
		}
	} else {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		writer = consoleWriter
	}
	// newrelic log forwarding is done by zerolog integration
	logger := zerolog.New(writer).Level(logLevel).With().Timestamp().Str("service", cfg.ServiceName).Str("env", cfg.Environment).Logger()

	if !cfg.IsProduction() {
		logger = logger.With().Stack().Logger()
	}
	return logger
}

// WithTraceContext adds new relic transaction context to logger
func WithTraceContext(logger zerolog.Logger, txn *newrelic.Transaction) zerolog.Logger {
	if txn == nil {
		return logger
	}
	// Get trace metadata from transaction
	metadata := txn.GetTraceMetadata()

	return logger.With().
		Str("trace.id", metadata.TraceID).
		Str("span.id", metadata.SpanID).
		Logger()
}

// NewPgxLogger creates a database logger
func NewPgxLogger(level zerolog.Level) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatFieldValue: func(i any) string {
			switch v := i.(type) {
			case string:
				// Clean and format SQL for better readability
				if len(v) > 200 {
					// Truncate very long SQL statements
					return v[:200] + "..."
				}
				return v
			case []byte:
				var obj interface{}
				if err := json.Unmarshal(v, &obj); err == nil {
					pretty, _ := json.MarshalIndent(obj, "", "    ")
					return "\n" + string(pretty)
				}
				return string(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}

	return zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("component", "database").
		Logger()
}

// GetPgxTraceLogLevel converts zerolog level to pgx tracelog level
func GetPgxTraceLogLevel(level zerolog.Level) int {
	switch level {
	case zerolog.DebugLevel:
		return 6 // tracelog.LogLevelDebug
	case zerolog.InfoLevel:
		return 4 // tracelog.LogLevelInfo
	case zerolog.WarnLevel:
		return 3 // tracelog.LogLevelWarn
	case zerolog.ErrorLevel:
		return 2 // tracelog.LogLevelError
	default:
		return 0 // tracelog.LogLevelNone
	}
}
