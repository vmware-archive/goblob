package blobstore

import "errors"

var ErrIteratorDone = errors.New("no more items in iterator")
var ErrIteratorAborted = errors.New("iterator aborted")

//go:generate counterfeiter . BucketIterator

type BucketIterator interface {
	Next() (*Blob, error)
	Done()
}
