package flightrecorderreceiver

import "go.opentelemetry.io/collector/scraper/scraperhelper"

type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`

	Include string `mapstructure:"include"`

	// prevent unkeyed literal initialization
	_ struct{}
}
