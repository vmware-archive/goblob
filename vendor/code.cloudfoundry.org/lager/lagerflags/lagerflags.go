package boshlogflags

import (
	"flag"
	"fmt"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	ERROR = "error"
	FATAL = "fatal"
)

type boshlogConfig struct {
	LogLevel string `json:"log_level,omitempty"`
}

func DefaultboshlogConfig() boshlogConfig {
	return boshlogConfig{
		LogLevel: string(INFO),
	}
}

var minLogLevel string

func AddFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(
		&minLogLevel,
		"logLevel",
		string(INFO),
		"log level: debug, info, error or fatal",
	)
}

func New(component string) (boshlog.Logger, *boshlog.ReconfigurableSink) {
	return newLogger(component, minLogLevel)
}

func NewFromConfig(component string, config boshlogConfig) (boshlog.Logger, *boshlog.ReconfigurableSink) {
	return newLogger(component, config.LogLevel)
}

func newLogger(component, minLogLevel string) (boshlog.Logger, *boshlog.ReconfigurableSink) {
	var minboshlogLogLevel boshlog.LogLevel

	switch minLogLevel {
	case DEBUG:
		minboshlogLogLevel = boshlog.DEBUG
	case INFO:
		minboshlogLogLevel = boshlog.INFO
	case ERROR:
		minboshlogLogLevel = boshlog.ERROR
	case FATAL:
		minboshlogLogLevel = boshlog.FATAL
	default:
		panic(fmt.Errorf("unknown log level: %s", minLogLevel))
	}

	logger := boshlog.NewLogger(component)

	sink := boshlog.NewReconfigurableSink(boshlog.NewWriterSink(os.Stdout, boshlog.DEBUG), minboshlogLogLevel)
	logger.RegisterSink(sink)

	return logger, sink
}
