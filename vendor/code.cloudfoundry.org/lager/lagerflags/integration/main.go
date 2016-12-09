package main

import (
	"errors"
	"flag"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"code.cloudfoundry.org/boshlog/boshlogflags"
)

func main() {
	boshlogflags.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := boshlogflags.New("cf-boshlog-integration")

	logger.Debug("component-does-action", boshlog.Data{"debug-detail": "foo"})
	logger.Info("another-component-action", boshlog.Data{"info-detail": "bar"})
	logger.Error("component-failed-something", errors.New("error"), boshlog.Data{"error-detail": "baz"})
	logger.Fatal("component-failed-badly", errors.New("fatal"), boshlog.Data{"fatal-detail": "quux"})
}
