package flightrecorderreceiver

import (
	"io"
	"os"
	"runtime/trace"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func primeFactors(t *testing.T, n int) []int {
	t.Helper()
	factors := []int{}
	for i := 2; i*i <= n; i++ {
		for n%i == 0 {
			factors = append(factors, i)
			n /= i
		}
	}
	if n > 1 {
		factors = append(factors, n)
	}
	return factors
}

func fibonacci(t *testing.T, n uint32) uint32 {
	t.Helper()
	if n < 2 {
		return n
	}
	return fibonacci(t, n-1) + fibonacci(t, n-2)
}

func generateFlightrecord(t *testing.T) (io.Reader, func() error) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "flightrecord-*.out")
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() error {
		return f.Close()
	}

	if err := trace.Start(f); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Go(
		func() {
			trace.WithRegion(t.Context(), "primeFactors", func() {
				list := primeFactors(t, 73*73)
				_ = list
			})
		})

	wg.Go(
		func() {
			trace.WithRegion(t.Context(), "fibonacci", func() { fibonacci(t, 23) })
		})
	wg.Wait()

	trace.Stop()

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	return f, cleanup
}

func TestConvert(t *testing.T) {
	f, cleanup := generateFlightrecord(t)
	defer cleanup()

	logger := zap.NewNop()

	p, err := convert(t.Context(), logger, f)
	if err != nil {
		t.Fatal(err)
	}

	// Log the converted profiles for visual inspection
	for _, rp := range p.ResourceProfiles().All() {
		t.Logf("ResourceProfile: %+v\n", rp)
		for _, sp := range rp.ScopeProfiles().All() {
			t.Logf("  ScopeProfile: %+v\n", sp)
			for _, prof := range sp.Profiles().All() {
				t.Logf("    Profile:\n")
				t.Logf("      Timestamp : %s\n", prof.Time().String())
				t.Logf("      Duration: %d ns\n", prof.DurationNano())
				for _, sample := range prof.Samples().All() {
					t.Logf("      Sample:\n")
					var tsStr string
					for _, ts := range sample.TimestampsUnixNano().All() {
						if tsStr != "" {
							tsStr += ", "
						}
						tsStr += time.Unix(0, int64(ts)).String()
					}
					t.Logf("        Timestamps: [%s]\n", tsStr)
					t.Logf("        Values: %v\n", sample.Values().AsRaw())
					for _, li := range p.Dictionary().StackTable().At(int(sample.StackIndex())).LocationIndices().All() {
						loc := p.Dictionary().LocationTable().At(int(li))
						t.Logf("        Location: 0x%x\n", loc.Address())
						for _, ln := range loc.Lines().All() {
							fn := p.Dictionary().FunctionTable().At(int(ln.FunctionIndex()))
							funcName := p.Dictionary().StringTable().At(int(fn.NameStrindex()))
							fileName := p.Dictionary().StringTable().At(int(fn.FilenameStrindex()))
							t.Logf("          Line: %d, Function: %s, Filename: %s\n", ln.Line(), funcName, fileName)
						}
					}
					t.Logf("")
				}
			}
		}
	}
}
