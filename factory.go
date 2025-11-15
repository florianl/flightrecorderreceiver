package flightrecorderreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/xreceiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
	"go.opentelemetry.io/collector/scraper/scraperhelper/xscraperhelper"
	"go.opentelemetry.io/collector/scraper/xscraper"
)

var compType = component.MustNewType("flightrecorder")

func NewFactory() receiver.Factory {
	return xreceiver.NewFactory(
		compType,
		createDefaultConfig,
		xreceiver.WithProfiles(createProfilesReceiver, component.StabilityLevelDevelopment),
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
	fileStatsConfig := cfg.(*Config)

	mp := newScraper(fileStatsConfig, settings)
	s, err := xscraper.NewProfiles(mp.scrape)
	if err != nil {
		return nil, err
	}

	opt := xscraperhelper.AddFactoryWithConfig(
		xscraper.NewFactory(compType, nil,
			xscraper.WithProfiles(
				func(context.Context, scraper.Settings, component.Config) (xscraper.Profiles, error) {
					return s, nil
				}, component.StabilityLevelDevelopment)),
		nil)

	return xscraperhelper.NewProfilesController(
		&fileStatsConfig.ControllerConfig,
		settings,
		consumer,
		opt,
	)
}
