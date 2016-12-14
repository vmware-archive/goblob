package ssh

import "io"

//go:generate counterfeiter -o ./fakes/fake_executor.go . Executor
type Executor interface {
	ExecuteForWrite(dest io.Writer, command string) error
	ExecuteForRead(command string) (io.Reader, error)
	SecureCopy(src string, dest io.Writer) (err error)
}
