package flightrecorderreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/xreceiver"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

var compType = component.MustNewType("flightrecorder")

func NewFactory() receiver.Factory {
	return xreceiver.NewFactory(
		compType,
		createDefaultConfig,
		xreceiver.WithProfiles(createProfilesReceiver, component.StabilityLevelDevelopment),
		xreceiver.WithMetrics(createMetricsReceiver, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		ControllerConfig: scraperhelper.NewDefaultControllerConfig(),
	}
}

func createProfilesReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	consumer xconsumer.Profiles,
) (xreceiver.Profiles, error) {
	c := cfg.(*Config)

	// Get or create the shared receiver instance
	rcv := c.getOrCreateReceiver(settings.TelemetrySettings)
	rcv.profilesConsumer = consumer

	return rcv, nil
}

func createMetricsReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	c := cfg.(*Config)

	// Get or create the shared receiver instance
	rcv := c.getOrCreateReceiver(settings.TelemetrySettings)
	rcv.metricsConsumer = consumer

	return rcv, nil
}
