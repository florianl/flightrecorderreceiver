package flightrecorderreceiver

import (
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`

	// Include specifies the glob pattern for flight recorder files to process
	Include string `mapstructure:"include"`

	// Internal fields for receiver coordination (not exposed via config)
	receiver     *flightRecorderReceiver
	receiverOnce sync.Once

	// prevent unkeyed literal initialization
	_ struct{}
}

// getOrCreateReceiver returns the receiver instance, creating it if necessary.
// This ensures that both profiles and metrics pipelines share the same receiver.
func (c *Config) getOrCreateReceiver(settings component.TelemetrySettings) *flightRecorderReceiver {
	c.receiverOnce.Do(func() {
		c.receiver = newFlightRecorderReceiver(c, settings)
	})
	return c.receiver
}
