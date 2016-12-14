package goblob

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/c0-ops/goblob/ssh"
)

type RemoteNFS struct {
	executor ssh.Executor
	tempDir  string
	blobs    []Blob
}

func NewRemoteNFS(executor ssh.Executor, tempDir string) Store {
	return &RemoteNFS{
		executor: executor,
		tempDir:  tempDir,
	}
}

func (r *RemoteNFS) List() ([]Blob, error) {
	cmd := "cd /var/vcap/store/shared && tar -cz ."

	reader, err := r.executor.ExecuteForRead(cmd)
	if err != nil {
		return nil, err
	} else {
		err = ExtractTar(reader, r.tempDir)
		if err != nil {
			return nil, err
		}
		if err = filepath.Walk(r.tempDir, r.walk); err != nil {
			return nil, err
		}
		return r.blobs, err
	}
}

func (r *RemoteNFS) walk(path string, info os.FileInfo, e error) error {
	if !info.IsDir() {
		filePath := strings.Replace(path, "/"+info.Name(), "", -1)
		checksum, checksumErr := MD5(path)
		if (checksumErr) != nil {
			return checksumErr
		}
		r.blobs = append(r.blobs, Blob{
			Filename: info.Name(),
			Path:     filePath,
			Checksum: checksum,
		})
	}

	return e
}

func (r *RemoteNFS) Read(src Blob) (io.Reader, error) {
	return os.Open(path.Join(src.Path, src.Filename))
}
func (r *RemoteNFS) Write(dst Blob, src io.Reader) error {
	return errors.New("not implemented")
}
