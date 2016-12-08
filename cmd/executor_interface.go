package cmd

import "io"

type Executor interface {
	ExecuteForWrite(destination io.Writer, command string) error
	ExecuteForRead(command string) (io.Reader, error)
	SecureCopy(src string, w io.Writer) (err error)
}
