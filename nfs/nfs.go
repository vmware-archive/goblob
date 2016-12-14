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

type NFS struct {
	executor ssh.Executor
	tempDir  string
	blobs    []goblob.Blob
}

func NewNFS(executor ssh.Executor, tempDir string) goblob.Store {
	return &NFS{
		executor: executor,
		tempDir:  tempDir,
	}
}

func (r *NFS) List() ([]goblob.Blob, error) {
	cmd := "cd /var/vcap/store/shared && tar -cz ."

	reader, err := r.executor.ExecuteForRead(cmd)
	if err != nil {
		return nil, err
	} else {
		err = goblob.ExtractTar(reader, r.tempDir)
		if err != nil {
			return nil, err
		}
		if err = filepath.Walk(r.tempDir, r.walk); err != nil {
			return nil, err
		}
		return r.blobs, err
	}
}

func (r *NFS) walk(path string, info os.FileInfo, e error) error {
	if !info.IsDir() {
		filePath := strings.Replace(path, "/"+info.Name(), "", -1)
		checksum, checksumErr := goblob.MD5(path)
		if (checksumErr) != nil {
			return checksumErr
		}
		r.blobs = append(r.blobs, goblob.Blob{
			Filename: info.Name(),
			Path:     filePath,
			Checksum: checksum,
		})
	}

	return e
}

func (r *NFS) Read(src goblob.Blob) (io.Reader, error) {
	return os.Open(path.Join(src.Path, src.Filename))
}
func (r *NFS) Write(dst goblob.Blob, src io.Reader) error {
	return errors.New("not implemented")
}
