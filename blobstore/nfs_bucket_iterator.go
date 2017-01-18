package blobstore

type nfsBucketIterator struct {
	blobCh chan *Blob
	doneCh chan struct{}
	errCh  chan error
}

func (i *nfsBucketIterator) Next() (*Blob, error) {
	if i.blobCh == nil {
		return nil, ErrIteratorDone
	}

	select {
	case blob, ok := <-i.blobCh:
		if !ok {
			i.blobCh = nil
			return nil, ErrIteratorDone
		}

		return blob, nil
	case err := <-i.errCh:
		if err != nil {
			return nil, err
		}
		return nil, ErrIteratorDone
	}
}

func (i *nfsBucketIterator) Done() {
	i.blobCh = nil
	close(i.doneCh)
}
