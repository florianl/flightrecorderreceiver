package flightrecorderreceiver

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
)

// flightRecorderReceiver is a custom receiver that parses flight recorder
// files and emits both profiles and metrics to their respective consumers.
type flightRecorderReceiver struct {
	cfg    *Config
	logger *zap.Logger

	profilesConsumer xconsumer.Profiles // may be nil
	metricsConsumer  consumer.Metrics   // may be nil

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// newFlightRecorderReceiver creates a new flight recorder receiver instance.
func newFlightRecorderReceiver(cfg *Config, settings component.TelemetrySettings) *flightRecorderReceiver {
	return &flightRecorderReceiver{
		cfg:    cfg,
		logger: settings.Logger,
	}
}

// Start begins the receiver's scraping loop in a background goroutine.
func (r *flightRecorderReceiver) Start(ctx context.Context, _ component.Host) error {
	ctx, r.cancel = context.WithCancel(ctx)

	r.wg.Go(func() {
		r.run(ctx)
	})

	return nil
}

// Shutdown stops the receiver and waits for the scraping goroutine to exit.
func (r *flightRecorderReceiver) Shutdown(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// run is the main loop that periodically scrapes flight recorder files.
func (r *flightRecorderReceiver) run(ctx context.Context) {
	// Initial delay before first scrape
	select {
	case <-time.After(r.cfg.ControllerConfig.InitialDelay):
	case <-ctx.Done():
		return
	}

	// Perform immediate first scrape
	if err := r.scrapeAndEmit(ctx); err != nil {
		r.logger.Error("failed to scrape flight recorder files", zap.Error(err))
	}

	// Create ticker for periodic scraping
	ticker := time.NewTicker(r.cfg.ControllerConfig.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.scrapeAndEmit(ctx); err != nil {
				r.logger.Error("failed to scrape flight recorder files", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

// scrapeAndEmit reads all matching flight recorder files, parses them,
// and emits the results to the profiles and metrics consumers.
func (r *flightRecorderReceiver) scrapeAndEmit(ctx context.Context) error {
	matches, err := doublestar.FilepathGlob(r.cfg.Include)
	if err != nil {
		return err
	}

	var scrapeErrors []error
	profiles := pprofile.NewProfiles()
	metrics := pmetric.NewMetrics()

	for _, match := range matches {
		f, err := os.Open(match)
		if err != nil {
			scrapeErrors = append(scrapeErrors, err)
			continue
		}

		newProfiles, newMetrics, convertErr := convert(ctx, r.logger, f)
		if convertErr != nil {
			scrapeErrors = append(scrapeErrors, convertErr)
			f.Close()
			continue
		}

		// Merge profiles
		if err := newProfiles.MergeTo(profiles); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}

		// Merge metrics
		mergeMetrics(newMetrics, metrics)

		f.Close()
	}

	// Emit profiles if consumer is configured
	if r.profilesConsumer != nil && profiles.ResourceProfiles().Len() > 0 {
		if err := r.profilesConsumer.ConsumeProfiles(ctx, profiles); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	// Emit metrics if consumer is configured
	if r.metricsConsumer != nil && metrics.ResourceMetrics().Len() > 0 {
		if err := r.metricsConsumer.ConsumeMetrics(ctx, metrics); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if len(scrapeErrors) > 0 {
		return errors.Join(scrapeErrors...)
	}

	return nil
}

// mergeMetrics merges metrics from src into dst.
func mergeMetrics(src, dst pmetric.Metrics) {
	src.ResourceMetrics().MoveAndAppendTo(dst.ResourceMetrics())
}
