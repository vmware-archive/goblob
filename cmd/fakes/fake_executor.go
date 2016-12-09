package fakes

import (
	"strings"
	"io"
	"errors"
)

//Taken from cfops

var (
	NfsSuccessString = "success nfs"
	NfsFailureString = "failed nfs"
	ErrMockNfsCommand = errors.New("error occurred")
)

type SuccessMockNFSExecuter struct {
	ActualCommand string
}

func (s *SuccessMockNFSExecuter) ExecuteForWrite(dest io.Writer, cmd string) (err error) {
	s.ActualCommand = cmd
	io.Copy(dest, strings.NewReader(NfsSuccessString))
	return
}

func (s *SuccessMockNFSExecuter) ExecuteForRead(cmd string) (io.Reader, error) {
	s.ActualCommand = cmd
	return strings.NewReader(NfsSuccessString), nil
}

func (s *SuccessMockNFSExecuter) SecureCopy(src string, w io.Writer) (err error) {
	io.Copy(w, strings.NewReader(NfsSuccessString))
	return
}

type FailureMockNFSExecuter struct{}

func (s *FailureMockNFSExecuter) ExecuteForWrite(dest io.Writer, cmd string) (err error) {
	io.Copy(dest, strings.NewReader(NfsFailureString))
	err = ErrMockNfsCommand
	return
}

func (s *FailureMockNFSExecuter) ExecuteForRead(cmd string) (io.Reader, error) {
	return nil, ErrMockNfsCommand
}

func (s *FailureMockNFSExecuter) SecureCopy(src string, w io.Writer) (err error) {
	io.Copy(w, strings.NewReader(NfsFailureString))
	return
}