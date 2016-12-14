package nfs

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/c0-ops/goblob"
	"github.com/c0-ops/goblob/ssh"
)

type NFSStore struct {
	executor ssh.Executor
	tempDir  string
	blobs    []goblob.Blob
}

func NewNFSStore(executor ssh.Executor, tempDir string) goblob.Store {
	return &NFSStore{
		executor: executor,
		tempDir:  tempDir,
	}
}

func (n *NFSStore) List() ([]goblob.Blob, error) {
	cmd := "cd /var/vcap/store/shared && tar -cz ."

	reader, err := n.executor.ExecuteForRead(cmd)
	if err != nil {
		return nil, err
	} else {
		err = goblob.ExtractTar(reader, n.tempDir)
		if err != nil {
			return nil, err
		}
		if err = filepath.Walk(n.tempDir, n.walk); err != nil {
			return nil, err
		}
		return n.blobs, err
	}
}

func (n *NFSStore) walk(path string, info os.FileInfo, e error) error {
	if !info.IsDir() {
		filePath := strings.Replace(path, "/"+info.Name(), "", -1)
		checksum, checksumErr := goblob.MD5(path)
		if (checksumErr) != nil {
			return checksumErr
		}
		n.blobs = append(n.blobs, goblob.Blob{
			Filename: info.Name(),
			Path:     filePath,
			Checksum: checksum,
		})
	}

	return e
}

func (n *NFSStore) Read(src goblob.Blob) (io.Reader, error) {
	return os.Open(path.Join(src.Path, src.Filename))
}
func (n *NFSStore) Write(dst goblob.Blob, src io.Reader) error {
	return errors.New("not implemented")
}
