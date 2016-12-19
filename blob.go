package goblob

func (b *Blob) Equal(blobToCompare Blob) bool {
	if b.Filename == blobToCompare.Filename && b.Checksum == blobToCompare.Checksum {
		return true
	}
	return false
}
