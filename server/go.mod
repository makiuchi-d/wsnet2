module wsnet2

go 1.20

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/dolthub/go-mysql-server v0.16.0
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/go-cmp v0.5.9
	github.com/jmoiron/sqlx v1.3.5
	github.com/pelletier/go-toml v1.9.5
	github.com/shiguredo/websocket v1.6.0
	github.com/spf13/cobra v1.7.0
	github.com/vmihailenco/msgpack/v5 v5.3.5
	go.uber.org/zap v1.24.0
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

replace github.com/dolthub/go-mysql-server => ../../../go-mysql-server

require (
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/dolthub/flatbuffers/v23 v23.3.3-dh.2 // indirect
	github.com/dolthub/go-icu-regex v0.0.0-20230524105445-af7e7991c97e // indirect
	github.com/dolthub/jsonpath v0.0.2-0.20230525180605-8dc13778fd72 // indirect
	github.com/dolthub/vitess v0.0.0-20230718053226-42bab255733a // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lestrrat-go/strftime v1.0.4 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tetratelabs/wazero v1.1.0 // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	gopkg.in/src-d/go-errors.v1 v1.0.0 // indirect
)
