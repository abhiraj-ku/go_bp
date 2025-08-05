package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/abhiraj-ku/go_bp/internal/config"
	"github.com/newrelic/go-agent/v3/newrelic"
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
