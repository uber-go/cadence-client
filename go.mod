module go.uber.org/cadence/v2

go 1.13

require (
	github.com/apache/thrift v0.13.0
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/mock v1.4.4
	github.com/kisielk/errcheck v1.2.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pborman/uuid v0.0.0-20160209185913-a97ce2ca70fa
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/uber-go/tally v3.3.15+incompatible
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/tchannel-go v1.16.0
	go.uber.org/atomic v1.7.0
	go.uber.org/fx v1.10.0 // indirect
	go.uber.org/goleak v1.0.0
	go.uber.org/multierr v1.6.0
	go.uber.org/thriftrw v1.25.0
	go.uber.org/yarpc v1.53.2
	go.uber.org/zap v1.13.0
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/time v0.0.0-20170927054726-6dc17368e09b
	honnef.co/go/tools v0.0.1-2019.2.3
)

// yarpc brings up new thrift version, which is not compatible with current codebase
replace github.com/apache/thrift => github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7
