package goblob

import (
	"fmt"

	"github.com/xchapter7x/lo"
)

func (b *Blob) Equal(blobToCompare Blob) bool {
	if b.Filename == blobToCompare.Filename {
		lo.G.Debug(fmt.Sprintf("Checksums [%s] comparing to [%s]", b.Checksum, blobToCompare.Checksum))
		if b.Checksum == blobToCompare.Checksum {
			return true
		} else {
			lo.G.Info(b.Filename, "checksum does not match re-uploading")
			return false
		}

	}
	return false
}
