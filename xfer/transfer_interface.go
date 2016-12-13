package xfer

type BlobTransferService interface {
	Transfer(buckets []string, destinationPath string) error
}
