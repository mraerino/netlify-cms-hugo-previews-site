package hugojs

import (
	"bytes"
	"os"
	"syscall/js"

	"github.com/pkg/errors"
)

type jsFile struct {
	fs *jsFS

	path string
	fi   *jsFileInfo

	body *bytes.Reader
}

func (f *jsFile) Close() error {
	return nil
}

func (f *jsFile) initBody() error {
	if f.body != nil {
		return nil
	}

	if f.fi.isDir {
		return errors.New("cannot read from a directory")
	}

	if f.fs.backendFS.Get("readFile").Type() != js.TypeFunction {
		return errors.New("invalid type: readFile")
	}

	cb, resCh := jsCallback()
	go f.fs.backendFS.Call("readFile", f.path, cb)
	res := <-resCh
	cb.Release()
	if res.err != nil {
		return res.err
	}

	bodyVal := res.vals[0]
	if bodyVal.Type() != js.TypeObject {
		return errors.New("got invalid return value")
	}

	body := make([]byte, 0)
	js.CopyBytesToGo(body, bodyVal)
	f.body = bytes.NewReader(body)
	return nil
}

func (f *jsFile) Read(p []byte) (n int, err error) {
	if err := f.initBody(); err != nil {
		return 0, err
	}

	return f.body.Read(p)
}

func (f *jsFile) ReadAt(p []byte, off int64) (n int, err error) {
	if err := f.initBody(); err != nil {
		return 0, err
	}

	return f.body.ReadAt(p, off)
}

func (f *jsFile) Seek(offset int64, whence int) (int64, error) {
	if err := f.initBody(); err != nil {
		return 0, err
	}

	return f.body.Seek(offset, whence)
}

func (f *jsFile) Write(p []byte) (n int, err error) {
	return 0, ErrReadOnly
}

func (f *jsFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, ErrReadOnly
}

func (f *jsFile) Name() string {
	return f.fi.name
}

func (f *jsFile) listFiles(num int) ([]os.FileInfo, error) {
	if !f.fi.isDir {
		return nil, errors.New("file is not a directory")
	}

	if f.fs.backendFS.Get("listDir").Type() != js.TypeFunction {
		return nil, errors.New("invalid type: listDir")
	}

	cb, resCh := jsCallback()
	go f.fs.backendFS.Call("listDir", f.path, cb)
	res := <-resCh
	cb.Release()
	if res.err != nil {
		return nil, res.err
	}

	filesVal := res.vals[0]
	if filesVal.Type() != js.TypeObject {
		return nil, errors.New("invalid type for return value")
	}

	filesLen := filesVal.Length()
	if filesLen > num {
		filesLen = num
	}

	fileInfos := make([]os.FileInfo, filesLen)
	for i := 0; i < filesLen; i++ {
		fi, err := fileInfoFromValue(filesVal.Index(i))
		if err != nil {
			return fileInfos, errors.Wrap(err, "failed to get file info")
		}
		fileInfos[i] = fi
	}
	return fileInfos, nil
}

func (f *jsFile) Readdir(count int) ([]os.FileInfo, error) {
	return f.listFiles(count)
}

func (f *jsFile) Readdirnames(n int) ([]string, error) {
	fileInfos, err := f.listFiles(n)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(fileInfos))
	for i, info := range fileInfos {
		names[i] = info.Name()
	}
	return names, nil
}

func (f *jsFile) Stat() (os.FileInfo, error) {
	return f.fs.statFile(f.path)
}

func (f *jsFile) Sync() error {
	return nil
}

func (f *jsFile) Truncate(size int64) error {
	return ErrReadOnly
}

func (f *jsFile) WriteString(s string) (ret int, err error) {
	return 0, ErrReadOnly
}
