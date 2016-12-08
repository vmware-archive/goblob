package tar

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const cmdExtractorLogTag = "CmdExtractor"

type CmdExtractor struct {
	runner boshsys.CmdRunner
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewCmdExtractor(
	runner boshsys.CmdRunner,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) CmdExtractor {
	return CmdExtractor{
		runner: runner,
		fs:     fs,
		logger: logger,
	}
}

func (e CmdExtractor) Extract(path string) (string, error) {
	extractPath, err := e.fs.TempDir("tar-CmdExtractor")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating extract destination")
	}

	e.logger.Debug(cmdExtractorLogTag, "Extract tar %s to %s", path, extractPath)

	_, _, _, err = e.runner.RunCommand("tar", "-C", extractPath, "-xzf", path)
	if err != nil {
		return "", bosherr.WrapError(err, "Running tar")
	}

	return extractPath, nil
}

func (e CmdExtractor) CleanUp(path string) error {
	return e.fs.RemoveAll(path)
}
