module github.com/florianl/flightrecorderreceiver

go 1.25.0

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/open-telemetry/sig-profiling/profcheck v0.0.0-20260416173450-2addcd15bcf9
	github.com/stretchr/testify v1.11.1
	github.com/zeebo/xxh3 v1.1.0
	go.opentelemetry.io/collector/component v1.56.0
	go.opentelemetry.io/collector/component/componenttest v0.150.0
	go.opentelemetry.io/collector/confmap v1.56.0
	go.opentelemetry.io/collector/consumer v1.56.0
	go.opentelemetry.io/collector/consumer/consumertest v0.150.0
	go.opentelemetry.io/collector/consumer/xconsumer v0.150.0
	go.opentelemetry.io/collector/pdata v1.56.0
	go.opentelemetry.io/collector/pdata/pprofile v0.150.0
	go.opentelemetry.io/collector/receiver v1.56.0
	go.opentelemetry.io/collector/receiver/receivertest v0.150.0
	go.opentelemetry.io/collector/receiver/xreceiver v0.150.0
	go.opentelemetry.io/collector/scraper/scraperhelper v0.150.0
	go.opentelemetry.io/otel v1.43.0
	go.opentelemetry.io/proto/otlp/profiles/v1development v0.3.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.1
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.4 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/cmd/mdatagen v0.149.0 // indirect
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.55.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.150.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.56.0 // indirect
	go.opentelemetry.io/collector/filter v0.149.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.150.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.56.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.150.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.150.0 // indirect
	go.opentelemetry.io/collector/scraper v0.150.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/telemetry v0.0.0-20260409153401-be6f6cb8b1fa // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/grpc v1.80.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.7.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)

tool (
	go.opentelemetry.io/collector/cmd/mdatagen
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
	mvdan.cc/gofumpt
)
