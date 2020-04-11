package hugojs

import (
	"os"
	"syscall/js"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

var ErrReadOnly = errors.New("FS is read-only")

type jsFS struct {
	backendFS js.Value
}

func (jsFS) Name() string {
	return "jsFS"
}

func NewJSFS(backendFS js.Value) (afero.Fs, error) {
	if backendFS.Type() != js.TypeObject {
		return nil, errors.New("Invalid backend")
	}
	return &jsFS{
		backendFS: backendFS,
	}, nil
}

func (fs *jsFS) statFile(path string) (*jsFileInfo, error) {
	if fs.backendFS.Get("stat").Type() != js.TypeFunction {
		return nil, errors.New("invalid type: stat")
	}

	cb, resCh := jsCallback()
	go fs.backendFS.Call("stat", path, cb)
	res := <-resCh
	cb.Release()
	if res.err != nil {
		return nil, res.err
	}

	return fileInfoFromValue(res.vals[0])
}

func (fs *jsFS) Open(name string) (afero.File, error) {
	fi, err := fs.statFile(name)
	if err != nil {
		return nil, err
	}
	return &jsFile{
		fs:   fs,
		path: name,
		fi:   fi,
	}, nil
}

func (fs *jsFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_WRONLY) != 0 {
		return nil, ErrReadOnly
	}

	return fs.Open(name)
}

func (fs *jsFS) Stat(name string) (os.FileInfo, error) {
	return fs.statFile(name)
}

// all below are not supported

func (fs *jsFS) Create(name string) (afero.File, error) {
	return nil, ErrReadOnly
}

func (fs *jsFS) Mkdir(name string, perm os.FileMode) error {
	return ErrReadOnly
}

func (fs *jsFS) MkdirAll(path string, perm os.FileMode) error {
	return ErrReadOnly
}

func (fs *jsFS) Remove(name string) error {
	return ErrReadOnly
}

func (fs *jsFS) RemoveAll(path string) error {
	return ErrReadOnly
}

func (fs *jsFS) Rename(oldname, newname string) error {
	return ErrReadOnly
}

func (fs *jsFS) Chmod(name string, mode os.FileMode) error {
	return ErrReadOnly
}

func (fs *jsFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrReadOnly
}
