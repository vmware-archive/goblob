package goblob

import (
	"fmt"
	"path"

	"github.com/xchapter7x/lo"
)

func (b *Blob) Equal(blobToCompare Blob) bool {
	if b.Filename == blobToCompare.Filename && b.Path == blobToCompare.Path {
		if b.Checksum == blobToCompare.Checksum {
			return true
		} else {
			lo.G.Info(fmt.Sprintf("checksum [%s] does not match [%s] for [%s]", b.Checksum, blobToCompare.Checksum, path.Join(b.Path, b.Filename)))
			return false
		}
	}
	return false
}
