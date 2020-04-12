package githubfs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/go-github/v31/github"
)

type githubFile struct {
	isDir bool
	path  string
	size  int64

	files []*github.RepositoryContent
	body  *bytes.Reader
}

func (f *githubFile) Close() error {
	return nil
}

// ErrIsDir is returned whenever the caller tries to use a read operation on a directory
var ErrIsDir = errors.New("Cannot do read operations on directories")

func (f *githubFile) Read(p []byte) (n int, err error) {
	if f.isDir {
		return 0, ErrIsDir
	}

	return f.body.Read(p)
}

func (f *githubFile) ReadAt(p []byte, off int64) (n int, err error) {
	if f.isDir {
		return 0, ErrIsDir
	}

	return f.body.ReadAt(p, off)
}

func (f *githubFile) Seek(offset int64, whence int) (int64, error) {
	if f.isDir {
		return 0, ErrIsDir
	}

	return f.body.Seek(offset, whence)
}

func (f *githubFile) Name() string {
	return f.path
}

func (f *githubFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.isDir {
		return nil, errors.New("Cannot list files, is not a directory")
	}

	infos := make([]os.FileInfo, len(f.files))
	for i, file := range f.files {
		infos[i] = &githubFileInfo{
			isDir: file.GetType() == "dir",
			name:  file.GetName(),
			size:  int64(file.GetSize()),
		}
	}
	return infos, nil
}

func (f *githubFile) Readdirnames(n int) ([]string, error) {
	if !f.isDir {
		return nil, errors.New("Cannot list files, is not a directory")
	}

	names := make([]string, len(f.files))
	for i, file := range f.files {
		names[i] = file.GetName()
	}
	return names, nil
}

func (f *githubFile) Stat() (os.FileInfo, error) {
	return &githubFileInfo{
		isDir: f.isDir,
		name:  filepath.Base(f.path),
		size:  f.size,
	}, nil
}

func (f *githubFile) Sync() error {
	return nil
}

// not supported, because read-olny

func (f *githubFile) Write(p []byte) (n int, err error) {
	return 0, ErrReadOnly
}

func (f *githubFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, ErrReadOnly
}

func (f *githubFile) Truncate(size int64) error {
	return ErrReadOnly
}

func (f *githubFile) WriteString(s string) (ret int, err error) {
	return 0, ErrReadOnly
}
