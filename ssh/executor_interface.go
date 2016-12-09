package ssh

import "io"

type Executor interface {
	ExecuteForWrite(dest io.Writer, command string) error
	ExecuteForRead(command string) (io.Reader, error)
	SecureCopy(src string, dest io.Writer) (err error)
}
