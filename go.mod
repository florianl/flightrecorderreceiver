module github.com/florianl/flightrecorderreceiver

go 1.26.7

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/open-telemetry/sig-profiling/profcheck v0.0.0-20260903071220-e05703c9b676
	github.com/stretchr/testify v1.12.1
	github.com/zeebo/xxh3 v1.1.0
	go.opentelemetry.io/collector/component v1.66.0
	go.opentelemetry.io/collector/component/componenttest v0.160.0
	go.opentelemetry.io/collector/confmap v1.66.0
	go.opentelemetry.io/collector/consumer v1.66.0
	go.opentelemetry.io/collector/consumer/consumertest v0.160.0
	go.opentelemetry.io/collector/consumer/xconsumer v0.160.0
	go.opentelemetry.io/collector/pdata v1.66.0
	go.opentelemetry.io/collector/pdata/pprofile v0.160.0
	go.opentelemetry.io/collector/receiver v1.66.0
	go.opentelemetry.io/collector/receiver/receivertest v0.160.0
	go.opentelemetry.io/collector/receiver/xreceiver v0.160.0
	go.opentelemetry.io/collector/scraper/scraperhelper v0.160.0
	go.opentelemetry.io/otel v1.46.0
	go.opentelemetry.io/proto/otlp/profiles/v1development v0.4.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.28.0
	golang.org/x/exp v0.0.0-20260824195058-e88cd73687aa
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/iimos/ucum v0.0.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/knadh/koanf/maps v0.1.3 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.1 // indirect
	github.com/knadh/koanf/v2 v2.3.6 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/cmd/mdatagen v0.160.0 // indirect
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.66.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.160.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.66.0 // indirect
	go.opentelemetry.io/collector/filter v0.160.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.160.0 // indirect
	go.opentelemetry.io/collector/internal/schemagen v0.160.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.66.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.160.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.160.0 // indirect
	go.opentelemetry.io/collector/scraper v0.160.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/telemetry v0.0.0-20260811182544-a038080d80e5 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	golang.org/x/vuln v1.7.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260803160001-6ac0973c030d // indirect
	google.golang.org/grpc v1.83.2 // indirect
	honnef.co/go/tools v0.8.1 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)

tool (
	go.opentelemetry.io/collector/cmd/mdatagen
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
	mvdan.cc/gofumpt
)
