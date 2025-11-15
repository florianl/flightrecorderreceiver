package flightrecorderreceiver

import (
	"context"
	"errors"
	"os"

	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper/scrapererror"

	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"
)

type frScraper struct {
	include string
	logger  *zap.Logger
}

func (s *frScraper) scrape(ctx context.Context) (pprofile.Profiles, error) {
	matches, err := doublestar.FilepathGlob(s.include)
	if err != nil {
		return pprofile.Profiles{}, err
	}

	var scrapeErrors []error
	var profiles pprofile.Profiles

	for _, match := range matches {
		f, err := os.Open(match)
		if err != nil {
			scrapeErrors = append(scrapeErrors, err)
			continue
		}
		newProfile, convertErr := convert(ctx, f)
		if convertErr != nil {
			scrapeErrors = append(scrapeErrors, convertErr)
			f.Close()
			continue
		}
		if err := newProfile.MergeTo(profiles); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
		f.Close()
	}

	if len(scrapeErrors) != 0 {
		return profiles, scrapererror.NewPartialScrapeError(errors.Join(scrapeErrors...), len(scrapeErrors))
	}

	return profiles, nil
}

func newScraper(cfg *Config, settings receiver.Settings) *frScraper {
	return &frScraper{
		include: cfg.Include,
		logger:  settings.Logger,
	}
}
